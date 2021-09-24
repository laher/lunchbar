package main

import (
	"context"
	"runtime"

	_ "embed"

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

type pluginRunner struct {
	plugin *plugins.Plugin
	st     *state
	ipc    *ipc.Client
}

func (r *pluginRunner) init() func() {
	return func() {
		log.Infof("command %s", r.plugin.Command)
		ctx := context.Background()
		r.plugin.Refresh(ctx)
		err := r.ipc.Write(14, []byte("hello server. I refreshed"))
		if err != nil {
			log.Warnf("could not write to server: %s", err)
		}
		title := r.plugin.Items.CycleItems[0].DisplayText()
		systray.SetTitle(title)   // doesn't do anything on windows.
		systray.SetTooltip(title) // not all platforms
		// necessary for windows - set title and icon ...
		ic, err := getTextIcon(title[0:1])
		//ic, err := getTextIcon("C")
		if err == nil {
			systray.SetIcon(ic)
		}

		if runtime.GOOS == "windows" {
			systray.AddMenuItem(title, "tooltip")
			systray.AddSeparator()
		}

		r.plugin.Debugf = log.Infof
		//p.AppleScriptTemplate = appleScriptDefaultTemplate
		log.Infof("found %d items", len(r.plugin.Items.ExpandedItems))
		for _, item := range r.plugin.Items.ExpandedItems {
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
								r.handleAction(subsubitem, mSubSubItem.ClickedCh)
							}
						} else {
							r.handleAction(subitem, mSubItem.ClickedCh)
						}
					}
				} else {
					r.handleAction(item, mItem.ClickedCh)
				}
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
func (r *pluginRunner) handleAction(item *plugins.Item, clickedChan <-chan struct{}) {

	action := item.Action()
	go func() {
		for _ = range clickedChan {
			// only run one action at once. avoids stuck actions from accumulating
			log.Infof("Clicked item %+v", item)
			r.st.lock.Lock()
			ctx := context.Background()
			action(ctx)
			log.Infof("click action complete")
			r.st.lock.Unlock()
		}
	}()
}

func (r *pluginRunner) onExit() {
}
