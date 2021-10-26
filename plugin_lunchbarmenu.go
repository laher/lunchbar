package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	xbarplugins "github.com/laher/lunchbox/xbar-plugins"
	"github.com/matryer/xbar/pkg/plugins"
)

func (r *pluginRunner) LunchbarMenu(title string) {
	r.mainItem = systray.AddMenuItem(title, "Lunchbar menu")
	systray.AddSeparator()

	mRefresh := r.mainItem.AddSubMenuItem("Refresh", "Refresh script")
	go func() {
		<-mRefresh.ClickedCh
		r.log.Info("Requesting refresh")
		r.lock.Lock()
		defer r.lock.Unlock()
		ctx := context.Background()
		r.refresh(ctx, false)
		r.log.Info("Finished refresh request")
	}()

	mRestart := r.mainItem.AddSubMenuItem("Restart plugin", "Restart plugin")
	go func() {
		<-mRestart.ClickedCh
		r.log.Info("Requesting plugin restart")
		r.lock.Lock()
		defer r.lock.Unlock()
		r.sendIPC(&IPCMessage{Type: msgPluginRestartme})
		r.log.Info("Finished restart request")
	}()

	mRefreshAll := r.mainItem.AddSubMenuItem("Refresh All", "Refresh all plugins")
	go func() {
		<-mRefreshAll.ClickedCh
		r.log.Info("Requesting refresh-all")
		r.lock.Lock()
		defer r.lock.Unlock()
		ctx := context.Background()
		r.refreshAll(ctx, false)
		r.log.Info("Finished refresh-all request")
	}()
	// TODO - what to do if EDITOR not set. VISUAL? Look for a default program? TextEdit/notepad/etc?
	editor := os.Getenv("VISUAL")
	if editor == "" {
		log.Warn("VISUAL not set")
	} else {
		mOpen := r.mainItem.AddSubMenuItem("Edit plugin script", "edit script")
		go func() {
			<-mOpen.ClickedCh
			r.log.Info("Requesting open file")
			r.lock.Lock()
			defer r.lock.Unlock()
			ctx := context.Background()

			// for now ... use VISUAL?
			log.Infof("running %s", editor)
			item := &plugins.Item{
				Params: plugins.ItemParams{
					Shell:       editor,
					ShellParams: []string{r.plugin.Command},
					Terminal:    false,
				},
				Plugin: r.plugin,
			}
			af := item.Action()
			af(ctx)
			r.log.Info("Finished file open request")
		}()
	}

	mOpenDir := r.mainItem.AddSubMenuItem("Open plugin scripts dir", "open scripts dir")
	go func() {
		<-mOpenDir.ClickedCh
		r.log.Info("Requesting open dir")
		r.lock.Lock()
		defer r.lock.Unlock()
		item := &plugins.Item{
			Params: plugins.ItemParams{
				// href handler uses 'open' etc.
				Href: filepath.Dir(r.plugin.Command),
			},
			Plugin: r.plugin,
		}
		ctx := context.Background()
		af := item.Action()
		af(ctx)
		r.log.Info("Finished open dir request")
	}()

	pis, err := xbarplugins.List(context.TODO())
	if err != nil {

	} else {
		mAddPlugin := r.mainItem.AddSubMenuItem("Add a plugin", "add a plugin")
		for _, pi := range pis {
			mAddPlugin.AddSubMenuItem(pi, pi)
		}
	}

	mQuit := r.mainItem.AddSubMenuItem("Quit Lunchbar", "Quit lunchbar")
	go func() {
		<-mQuit.ClickedCh
		r.sendIPC(&IPCMessage{Type: msgPluginQuit})
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}
