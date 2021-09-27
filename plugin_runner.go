package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	_ "embed"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	ipc "github.com/james-barrow/golang-ipc"
	"github.com/matryer/xbar/pkg/plugins"
)

type pluginRunner struct {
	plugin      *plugins.Plugin
	lock        sync.Mutex
	ipc         *ipc.Client
	mainItem    *systray.MenuItem
	items       []*itemWrap
	subitems    map[*itemWrap][]*itemWrap
	subsubitems map[*itemWrap][]*itemWrap
	log         *log.Entry
}

func newPluginRunner(bin string, ipc *ipc.Client) *pluginRunner {
	p := plugins.NewPlugin(bin)
	r := &pluginRunner{
		plugin:      p,
		ipc:         ipc,
		items:       []*itemWrap{},
		subitems:    map[*itemWrap][]*itemWrap{},
		subsubitems: map[*itemWrap][]*itemWrap{},
		log:         log.WithField("plugin", filepath.Base(bin)).WithField("pid", os.Getpid()),
	}
	r.log.Infof("plugin runner initialised. full path: %s", bin)
	return r
}

func (r *pluginRunner) Listen() {
	for {
		r.log.Infof("listen for messages ")
		m, err := r.ipc.Read()
		if err != nil {
			r.log.Errorf("IPC server error %s", err)
			break
		}
		r.log.WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Infof("plugin runner received message")
	}
}

func (r *pluginRunner) init() func() {
	r.log.Infof("launching systray icon")
	return func() {
		r.log.Infof("command %s", r.plugin.Command)
		ctx := context.Background()
		r.refresh(ctx, true)
		go func() {
			time.Sleep(5 * time.Second)
			r.loop()
		}()

	}

}

func (r *pluginRunner) loop() {
	if r.plugin.RefreshInterval.Duration() > 0 {
		ctx := context.Background()
		for {
			r.log.Infof("refresh every %v", r.plugin.RefreshInterval.Duration().String())
			time.Sleep(r.plugin.RefreshInterval.Duration())
			r.refresh(ctx, false)
		}
	} else {
		// nothing to do. this plugin is static
		r.log.Info("no refresh")
	}
}

func (r *pluginRunner) sendIPC(s string) {
	err := r.ipc.Write(msgcodeDefault, []byte(s))
	if err != nil {
		r.log.Warnf("could not write to server: %s", err)
	}
}

func (r *pluginRunner) refresh(ctx context.Context, initial bool) {
	r.plugin.Refresh(ctx)
	r.sendIPC("I refreshed")
	title := r.plugin.Items.CycleItems[0].DisplayText()
	systray.SetTitle(title)   // doesn't do anything on windows.
	systray.SetTooltip(title) // not all platforms
	// necessary for windows - set title and icon ...
	ic, err := getTextIcon(title[0:1])
	//ic, err := getTextIcon("C")
	if err == nil {
		if runtime.GOOS != "windows" || initial { // <- Windows seems to have problems with settings anything after icon
			systray.SetIcon(ic)
		}
	}

	if initial {
		r.addCrossbarMenu(title)
	} else if runtime.GOOS == "windows" {
		r.mainItem.SetTitle(title)
	}

	r.plugin.Debugf = r.log.Infof
	r.plugin.AppleScriptTemplate = appleScriptDefaultTemplate
	r.log.Infof("found %d items", len(r.plugin.Items.ExpandedItems))
	for index, item := range r.plugin.Items.ExpandedItems {
		var itemW *itemWrap
		if item.Params.Separator {
			if len(r.items) < index+1 {
				itemW = &itemWrap{isSeparator: true, plugItem: item}
				itemW.trayItem = systray.AddMenuItem("----------", "separator")
				r.items = append(r.items, itemW)
			} else {
				itemW = r.items[index]
				itemW.plugItem = item
				itemW.trayItem.SetTitle("-------------")
				itemW.trayItem.Show()
			}
			itemW.isSeparator = true
			//itemW.trayItem.Disable()
			r.items = append(r.items, itemW)
		} else {
			if len(r.items) < index+1 {
				itemW = &itemWrap{isSeparator: false, plugItem: item}
				itemW.trayItem = systray.AddMenuItem(item.DisplayText(), "tooltip")
				r.items = append(r.items, itemW)
				r.handleAction(itemW)
			} else {
				itemW = r.items[index]
				itemW.trayItem.SetTitle(item.DisplayText())
				itemW.trayItem.Show()
			}
			if len(item.Items) > 0 {
				subitemWs := r.subitems[itemW]
				if subitemWs == nil {
					subitemWs = []*itemWrap{}
				}
				for subindex, subitem := range item.Items {
					var subitemW *itemWrap
					if len(subitemWs) < subindex+1 {
						subitemW = &itemWrap{isSeparator: false, plugItem: subitem}
						subitemW.trayItem = itemW.trayItem.AddSubMenuItem(subitem.DisplayText(), "tooltip")
						subitemWs = append(subitemWs, subitemW)
						r.handleAction(subitemW)
					} else {
						subitemW = subitemWs[subindex]
						subitemW.trayItem.SetTitle(subitem.DisplayText())
						subitemW.trayItem.Show()
					}
					if len(subitem.Items) > 0 {
						subsubitemWs := r.subsubitems[subitemW]
						if subsubitemWs == nil {
							subsubitemWs = []*itemWrap{}
						}
						for subsubindex, subsubitem := range subitem.Items {
							var subsubitemW *itemWrap
							if len(subsubitemWs) < subsubindex+1 {
								subsubitemW = &itemWrap{isSeparator: false, plugItem: subsubitem}
								subsubitemW.trayItem = subitemW.trayItem.AddSubMenuItem(subsubitem.DisplayText(), "tooltip")
								subsubitemWs = append(subsubitemWs, subsubitemW)
								r.handleAction(subsubitemW)
							} else {
								subsubitemW = subsubitemWs[subsubindex]
								subsubitemW.trayItem.SetTitle(subsubitem.DisplayText())
								subsubitemW.trayItem.Show()
							}
						}
						r.subsubitems[subitemW] = subsubitemWs
					}
				}
				r.subitems[itemW] = subitemWs
			}
		}
	}
	if len(r.plugin.Items.ExpandedItems) < len(r.items) {
		for i := len(r.plugin.Items.ExpandedItems); i < len(r.items); i++ {
			r.items[i].trayItem.Hide()
		}
	}

}

func (r *pluginRunner) handleAction(item *itemWrap) {
	//item *plugins.Item, clickedChan <-chan struct{}) {

	go func() {
		for range item.trayItem.ClickedCh {
			// only run one action at once. avoids stuck actions from accumulating
			r.log.Infof("Clicked item %+v", item)
			r.lock.Lock()
			ctx := context.Background()
			item.DoAction(ctx)
			r.log.Infof("click action complete")
			r.lock.Unlock()
		}
	}()
}

func (r *pluginRunner) onExit() {
}

const appleScriptDefaultTemplate = `
			set quotedScriptName to quoted form of "{{ .Command }}"
		{{ if .Params }}
			set commandLine to {{ .Vars }} & " " & quotedScriptName & " " & {{ .Params }}
		{{ else }}
			set commandLine to {{ .Vars }} & " " & quotedScriptName
		{{ end }}
			if application "Terminal" is running then 
				tell application "Terminal"
					do script commandLine
					activate
				end tell
			else
				tell application "Terminal"
					do script commandLine in window 1
					activate
				end tell
			end if
		`
