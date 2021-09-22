package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	"github.com/matryer/xbar/pkg/plugins"
)

func rootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "crossbar")
}

// TODO - OS-dependent dir?
func pluginsDir() string {
	return filepath.Join(rootDir(), "plugins")
}

func isExecutable(fi os.FileMode) bool {
	return fi.Perm()&0111 != 0
}

func main() {

	flag.Parse()
	plugin := ""
	// for now don't do any flag parsing
	if len(os.Args) > 1 {
		plugin = os.Args[1]
	}
	pluginsDir := pluginsDir()
	st := &state{}
	if plugin != "" {
		bin := filepath.Join(pluginsDir, plugin)
		// run this plugin
		//c := command{Cmd: }

		p := plugins.NewPlugin(bin)
		/*
			log.Infof("cmd.Run(): %s", bin)
			cmd := exec.Command(bin)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}
			log.Infof("cmd.Run() complete")
			outs := strings.Split(string(out), "\n")
			t := &config{
				Title: outs[0],
				Items: []string{},
			}
			if len(outs) > 1 {
				t.Items = outs[1:]
			}

			process(t)
		*/
		log.Infof("launching systray icon")
		systray.Run(onReady(p, st), onExit)
	} else {
		log.Infof("Loading plugins from %s", pluginsDir)
		s, err := newSupervisor()
		if err != nil {
			panic(err)
		}
		s.Start()
		// TODO refresh occasionally?
		return
		// launch all the plugins
		des, err := os.ReadDir(pluginsDir)
		if err != nil {
			panic(err)
		}
		for _, de := range des {
			if !de.IsDir() {
				finf, err := de.Info()
				if err != nil {
					log.Warnf("cant get file info: %s, %v", err, de.Type())
				}
				if !isExecutable(finf.Mode()) {
					log.Warnf("cant check excutableness: %s, %v", err, finf.Mode())

				}
				//&& isExecutable(de.Type()) {
				log.Infof("Loading plugin %s", de.Name())
				/* TODO
				c := command{Cmd: filepath.Join(pluginsDir, de.Name())}
				cm := c.cmd()
				err := cm[0].Run()
				if err != nil {
					panic(err)
				}
				*/
			} else {
				log.Infof("Plugin file invalid: %+v, %v", de, de.Type())
			}
		}
	}
}

type state struct {
	config config
	lock   sync.Mutex
}

func appMenu(mItem *systray.MenuItem) {
	mItem.AddSubMenuItem("Refresh", "")
	mItem.AddSubMenuItem("Open plugins dir", "")

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

func onReady(p *plugins.Plugin, st *state) func() {
	return func() {
		ctx := context.Background()
		p.Refresh(ctx)
		title := p.Items.CycleItems[0].DisplayText()
		systray.SetTitle(title)
		// if windows, set title and icon ...
		if runtime.GOOS == "windows" {
			// TODO - always set an icon
			systray.SetTooltip(title)
		} else {
			systray.SetTooltip("Crossbar")
		}
		p.Debugf = log.Infof
		//p.AppleScriptTemplate = appleScriptDefaultTemplate
		log.Infof("found %d items", len(p.Items.ExpandedItems))
		for _, item := range p.Items.ExpandedItems {
			if item.Params.Separator {
				systray.AddSeparator()
			} else {

				//if item.Alternate != nil {
				//	parts := strings.Split(item, "|")
				//	mItem := systray.AddMenuItem(strings.TrimSpace(parts[0]), "tooltip")
				mItem := systray.AddMenuItem(item.DisplayText(), "tooltip")

				/*
					if len(parts) > 1 {
						pairStrs := strings.Split(parts[1], " ")
						pairs := map[string]string{}
						for _, pairStr := range pairStrs {
							pair := strings.Split(pairStr, "=")
							if len(pair) > 1 {
								pairs[pair[0]] = pair[1]
							} else {
								pairs[pair[0]] = ""
							}
						}
				*/

				if len(item.Items) > 0 {
					for _, subitem := range item.Items {
						mSubItem := mItem.AddSubMenuItem(subitem.DisplayText(), "tooltip")
						//mSubitem := systray.AddMenuItem(subitem.DisplayText(), "tooltip")
						if len(subitem.Items) > 0 {
							for _, subsubitem := range subitem.Items {
								mSubSubItem := mSubItem.AddSubMenuItem(subsubitem.DisplayText(), "tooltip")
								handleAction(st, subsubitem, mSubSubItem.ClickedCh)
							}
						} else {
							handleAction(st, subitem, mSubItem.ClickedCh)
						}
					}
				} else {
					handleAction(st, item, mItem.ClickedCh)
				}
				//}
			}
		}

		mCB := systray.AddMenuItem("Crossbar", "crossbar settings")
		mCB.AddSubMenuItem("Refresh", "Refresh script")
		mQuit := mCB.AddSubMenuItem("Quit", "Quit crossbar")
		go func() {
			<-mQuit.ClickedCh
			log.Info("Requesting quit")
			systray.Quit()
			log.Info("Finished quit request")
		}()

	}
}
func handleAction(st *state, item *plugins.Item, clickedChan <-chan struct{}) {

	action := item.Action()
	go func() {
		for _ = range clickedChan {
			// only run one action at once. avoids stuck actions from accumulating
			log.Infof("Clicked item %+v", item)
			st.lock.Lock()
			ctx := context.Background()
			action(ctx)
			log.Infof("click action complete")
			st.lock.Unlock()
		}
		/*
			for k, v := range pairs {
				switch k {
				case "href":
					// link
					log.Infof("open href with %s", v)
				case "shell":
					log.Infof("open shell with %s", v)
					// shell
				}
			}
		*/
	}()
}

func onExit() {
}
