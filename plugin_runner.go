package main

import (
	"context"
	"runtime"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	ipc "github.com/james-barrow/golang-ipc"
	"github.com/matryer/xbar/pkg/plugins"
)

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

func initPluginRunner(ipcc *ipc.Client, p *plugins.Plugin, st *state) func() {
	return func() {
		log.Infof("command %s", p.Command)
		ctx := context.Background()
		p.Refresh(ctx)
		_ = ipcc.Write(14, []byte("hello server. I refreshed"))
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
				mItem := systray.AddMenuItem(item.DisplayText(), "tooltip")

				if len(item.Items) > 0 {
					for _, subitem := range item.Items {
						mSubItem := mItem.AddSubMenuItem(subitem.DisplayText(), "tooltip")
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
	}()
}
