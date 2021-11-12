package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
)

func (r *pluginRunner) LunchbarMenu(title string) {
	r.mainItem = systray.AddMenuItem(title, "Lunchbar menu")
	systray.AddSeparator()

	r.menuThisPlugin()
	r.menuManageLunchbar()
	r.menuManagePlugins()

	mQuit := r.mainItem.AddSubMenuItem("Quit Lunchbar", "Quit lunchbar")
	go func() {
		// should only be once really
		<-mQuit.ClickedCh
		r.sendIPC(&IPCMessage{Type: msgPluginQuit})
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}

func (r *pluginRunner) menuManageLunchbar() {
	mManageLunchbar := r.mainItem.AddSubMenuItem("Manage lunchbar", "manage lunchbar itself")
	mOpen := mManageLunchbar.AddSubMenuItem("Edit .env file", "edit config file")
	go func() {
		for range mOpen.ClickedCh {
			r.log.Info("Requesting open file")
			r.lock.Lock()
			if err := osEditFile(filepath.Join(rootDir(), ".env")); err != nil {
				r.log.WithError(err).Warn("failed to open file")
			}
			r.lock.Unlock()
			r.log.Info("Finished file open request")
		}
	}()
}
