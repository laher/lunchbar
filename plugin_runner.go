package main

import (
	"context"
	"flag"
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
	"github.com/laher/lunchbox/lunch"
	"github.com/matryer/xbar/pkg/plugins"
)

func runPlugin(ctx context.Context, args []string) error {
	if err := os.Chdir(pluginsDir()); err != nil {
		return err
	}
	var pluginCmd = flag.NewFlagSet("plugin", flag.ExitOnError)
	var pluginPtr = pluginCmd.String("plugin", "", "name of plugin to run")
	err := pluginCmd.Parse(args)
	if err != nil {
		return err
	}
	// TODO should we do dotenv?
	// godotenv.Load(filepath.Base(*pluginPtr) + ".env")
	key := "lunchbar_" + filepath.Base(*pluginPtr)
	sc, err := ipc.StartClient(key, nil)
	if err != nil {
		log.Errorf("could not start IPC client: %s", err)
		return err
	}
	bin := *pluginPtr
	if !strings.Contains(*pluginPtr, "/") {
		bin = filepath.Join(pluginsDir(), *pluginPtr)
	}
	r := newPluginRunner(bin, sc)
	go r.Listen()
	systray.Run(r.init(ctx), r.onExit)
	return nil
}

type pluginRunner struct {
	plugin   *plugins.Plugin
	lock     sync.Mutex
	ipc      *ipc.Client
	mainItem *systray.MenuItem
	items    []*itemWrap
	log      *log.Entry
	//subitems    [][]*itemWrap
	//subsubitems [][]*itemWrap
}

func (r *pluginRunner) refreshItems(ctx context.Context) error {
	if err := os.Chdir(pluginsDir()); err != nil {
		r.plugin.OnErr(err)
		return err
	}
	if strings.HasSuffix(r.plugin.Command, ".elvish") || strings.HasSuffix(r.plugin.Command, ".elv") {
		// run it and parse output
		out, err := lunch.ElvishRunScript(ctx, []string{r.plugin.Command})
		if err != nil {
			r.plugin.OnErr(err)
			return err
		}
		items, err := r.plugin.ParseOutput(ctx, r.plugin.Command, strings.NewReader(strings.Join(out, "\n")))
		if err != nil {
			r.plugin.OnErr(err)
			return err
		}
		r.plugin.Items = items
	} else {
		r.plugin.Refresh(ctx)
	}
	r.sendIPC(msgPluginRefreshComplete)
	return nil
}

const (
	msgPluginRefreshAll      = "refresh-all"
	msgPluginRefreshComplete = "refresh-complete"
	msgPluginRefreshError    = "refresh-error"
	msgPluginUnrecognised    = "unrecognised"
	msgPluginQuit            = "quit"
	msgPluginRestartme       = "restartme"
)

func newPluginRunner(bin string, ipcc *ipc.Client) *pluginRunner {
	p := plugins.NewPlugin(bin)
	r := &pluginRunner{
		plugin: p,
		ipc:    ipcc,
		items:  []*itemWrap{},
		log:    log.WithField("plugin", filepath.Base(bin)).WithField("pid", os.Getpid()),
	}
	r.log.Infof("plugin runner initialised. full path: %s", bin)
	return r
}

func (r *pluginRunner) Listen() {
	r.log.Infof("listen for messages")
	for {
		m, err := r.ipc.Read()
		if err != nil {
			r.log.Errorf("IPC server error %s", err)
			break
		}
		r.log.WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Debug("plugin runner received message")

		if m.MsgType < 0 {
			// diagnostical
			continue
		}
		switch string(m.Data) {
		case msgSupervisorRefresh:
			r.lock.Lock()
			r.refresh(context.Background(), false)
			r.lock.Unlock()
		case msgSupervisorUnrecognised:
			// TODO die here?
			r.log.WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Warnf("supervisor did not recognise previous command")
		case msgSupervisorQuit:
			r.log.Info("Requesting systray quit")
			systray.Quit()
			r.log.Info("Finished quit request")
		default:
			r.log.WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Error("Plugin received a message type which it doesn't recognise")
			r.sendIPC(msgPluginUnrecognised)
		}
	}
}

func (r *pluginRunner) init(ctx context.Context) func() {
	r.log.Debug("launching systray icon")
	return func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		r.log.Debugf("initialise plugin with command %s", r.plugin.Command)
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
		r.log.Infof("refresh every %v", r.plugin.RefreshInterval.Duration().String())
		for {
			time.Sleep(r.plugin.RefreshInterval.Duration())
			r.lock.Lock()
			r.refresh(ctx, false)
			r.lock.Unlock()
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

func (r *pluginRunner) refreshAll(ctx context.Context, initial bool) {
	r.sendIPC(msgPluginRefreshAll)

}

// TODO allow icon configuration (via dotenv?)
func (r *pluginRunner) iconOnly() bool {
	return runtime.GOOS == osWindows
}

func (r *pluginRunner) refresh(ctx context.Context, initial bool) {
	r.log.Debug("refresh items")
	err := r.refreshItems(ctx)
	if err != nil {
		r.sendIPC(msgPluginRefreshError)
	}
	r.log.Debug("items refreshed")
	title := r.plugin.Items.CycleItems[0].DisplayText()
	systray.SetTitle(title)   // doesn't do anything on windows.
	systray.SetTooltip(title) // not all platforms

	lunchbarTitle := "Lunchbar"
	if r.iconOnly() {
		// necessary for windows - set icon ...
		// it's a bit ugly right now so just leave it for other platforms for now ...
		ic, err := getTextIcon(strings.TrimSpace(title)[:1])
		if err == nil {
			if runtime.GOOS != osWindows || initial { // <- Windows seems to have problems with settings anything after icon
				systray.SetIcon(ic)
			}
		}
		lunchbarTitle = title
	}

	if initial {
		r.LunchbarMenu(lunchbarTitle)
	} else {
		r.mainItem.SetTitle(lunchbarTitle)
	}

	// override the default logger
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
	var title = ""
	if item.Params.Separator {
		title = "----------"
	} else {
		title = item.DisplayText()
	}
	if len(r.items) < index+1 {
		itemW = &itemWrap{plugItem: item, subitems: []*itemWrap{}}
		itemW.trayItem = systray.AddMenuItem(title, item.DisplayText())
		r.items = append(r.items, itemW)
		r.handleAction(itemW)
	} else {
		itemW = r.items[index]
		itemW.trayItem.SetTitle(item.DisplayText())
	}
	itemW.plugItem = item
	itemW.isSeparator = item.Params.Separator
	itemW.trayItem.Show()
	if len(item.Items) > 0 {
		itemW.trayItem.Enable()
		for subindex, subitem := range item.Items {
			r.loadSubitem(itemW, subindex, subitem)
		}
		if len(item.Items) < len(itemW.subitems) {
			for i := len(item.Items); i < len(itemW.subitems); i++ {
				itemW.subitems[i].trayItem.Hide()
			}
		}
	} else {
		if !item.Params.Separator && itemW.plugItem.Action() != nil {
			itemW.trayItem.Enable()
		} else {
			itemW.trayItem.Disable()
		}
	}
}

func (r *pluginRunner) loadSubitem(itemW *itemWrap, subindex int, subitem *plugins.Item) {
	var subitemW *itemWrap
	if len(itemW.subitems) == subindex { // need to add it
		subitemW = &itemWrap{isSeparator: false, plugItem: subitem, subitems: []*itemWrap{}}
		subitemW.trayItem = itemW.trayItem.AddSubMenuItem(subitem.DisplayText(), "tooltip")
		itemW.subitems = append(itemW.subitems, subitemW)
		r.handleAction(subitemW)
	} else if len(itemW.subitems) > subindex {
		subitemW = itemW.subitems[subindex]
		subitemW.trayItem.SetTitle(subitem.DisplayText())
		subitemW.trayItem.Show()
	} else {
		// error
		r.log.WithFields(log.Fields{
			"subindex":    subindex,
			"displaytext": subitem.DisplayText(),
			"len":         len(itemW.subitems),
		}).Fatal("not enough subitems. Unexpected. Die")
	}
	if len(subitem.Items) > 0 {
		subitemW.trayItem.Enable()
		for subsubindex, subsubitem := range subitem.Items {
			var subsubitemW *itemWrap
			if len(subitemW.subitems) < subsubindex+1 {
				subsubitemW = &itemWrap{isSeparator: false, plugItem: subsubitem, subitems: []*itemWrap{}}
				subsubitemW.trayItem = subitemW.trayItem.AddSubMenuItem(subsubitem.DisplayText(), "tooltip")
				subitemW.subitems = append(subitemW.subitems, subsubitemW)
				r.handleAction(subsubitemW)
			} else {
				subsubitemW = subitemW.subitems[subsubindex]
				subsubitemW.trayItem.SetTitle(subsubitem.DisplayText())
				subsubitemW.trayItem.Show()
			}
		}
		if len(subitem.Items) < len(subitemW.subitems) {
			for i := len(subitem.Items); i < len(subitemW.subitems); i++ {
				itemW.subitems[i].trayItem.Hide()
			}
		}
	} else if subitemW.plugItem.Action() != nil {
		subitemW.trayItem.Enable()
	} else {
		subitemW.trayItem.Disable()
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
