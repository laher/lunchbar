package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func (r *pluginRunner) refreshItems(ctx context.Context) {
	if strings.HasSuffix(r.plugin.Command, ".elvish") || strings.HasSuffix(r.plugin.Command, ".elv") {
		// run it and parse output
		out, err := elvishRunScript(r.plugin.Command, os.Stdout, os.Stderr, []string{r.plugin.Command})
		if err != nil {
			r.plugin.OnErr(err)
			return
		}
		items, err := r.plugin.ParseOutput(ctx, r.plugin.Command, strings.NewReader(strings.Join(out, "\n")))
		if err != nil {
			r.plugin.OnErr(err)
			return
		}
		r.plugin.Items = items
	} else {
		r.plugin.Refresh(ctx)
	}
	r.sendIPC("I refreshed")
}

func newPluginRunner(bin string, ipcc *ipc.Client) *pluginRunner {
	p := plugins.NewPlugin(bin)
	r := &pluginRunner{
		plugin:      p,
		ipc:         ipcc,
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

const osWindows = "windows"

func (r *pluginRunner) refresh(ctx context.Context, initial bool) {
	r.refreshItems(ctx)
	title := r.plugin.Items.CycleItems[0].DisplayText()
	systray.SetTitle(title)   // doesn't do anything on windows.
	systray.SetTooltip(title) // not all platforms
	// necessary for windows - set title and icon ...
	ic, err := getTextIcon(title[0:1])
	if err == nil {
		if runtime.GOOS != osWindows || initial { // <- Windows seems to have problems with settings anything after icon
			systray.SetIcon(ic)
		}
	}

	if initial {
		r.addCrossbarMenu(title)
	} else if runtime.GOOS == osWindows {
		r.mainItem.SetTitle(title)
	}

	r.plugin.Debugf = r.log.Infof
	r.plugin.AppleScriptTemplate = appleScriptDefaultTemplate
	r.log.Infof("found %d items", len(r.plugin.Items.ExpandedItems))
	for index, item := range r.plugin.Items.ExpandedItems {
		r.loadItem(index, item)
	}
	if len(r.plugin.Items.ExpandedItems) < len(r.items) {
		for i := len(r.plugin.Items.ExpandedItems); i < len(r.items); i++ {
			r.items[i].trayItem.Hide()
		}
	}
}

func (r *pluginRunner) loadItem(index int, item *plugins.Item) {
	var itemW *itemWrap
	if item.Params.Separator {
		if len(r.items) < index+1 {
			itemW = &itemWrap{isSeparator: true, plugItem: item}
			itemW.trayItem = systray.AddMenuItem("----------", "separator")
			r.items = append(r.items, itemW)
			r.handleAction(itemW)
		} else {
			itemW = r.items[index]
			itemW.isSeparator = true
			itemW.plugItem = item
			itemW.trayItem.SetTitle("-------------")
			itemW.trayItem.Show()
		}
		// itemW.trayItem.Disable()
		r.items = append(r.items, itemW)
	} else {
		if len(r.items) < index+1 {
			itemW = &itemWrap{isSeparator: false, plugItem: item}
			itemW.trayItem = systray.AddMenuItem(item.DisplayText(), "tooltip")
			r.items = append(r.items, itemW)
			r.handleAction(itemW)
		} else {
			itemW = r.items[index]
			itemW.isSeparator = false
			itemW.trayItem.SetTitle(item.DisplayText())
			itemW.trayItem.Show()
		}
		if len(item.Items) > 0 {
			subitemWs := r.subitems[itemW]
			if subitemWs == nil {
				subitemWs = []*itemWrap{}
			}
			for subindex, subitem := range item.Items {
				r.loadSubitem(itemW, subitemWs, subindex, subitem)
			}
			r.subitems[itemW] = subitemWs

			if len(item.Items) < len(subitemWs) {
				for i := len(item.Items); i < len(subitemWs); i++ {
					subitemWs[i].trayItem.Hide()
				}
			}
		}
	}
}

func (r *pluginRunner) loadSubitem(itemW *itemWrap, subitemWs []*itemWrap, subindex int, subitem *plugins.Item) {
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
		if len(subitem.Items) < len(subsubitemWs) {
			for i := len(subitem.Items); i < len(subsubitemWs); i++ {
				subitemWs[i].trayItem.Hide()
			}
		}
	}
}

func (r *pluginRunner) handleAction(item *itemWrap) {
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
