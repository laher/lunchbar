package main

import "context"

func (r *pluginRunner) menuThisPlugin() {

	mThisPlugin := r.mainItem.AddSubMenuItem("This plugin", "manage this plugin")
	mRefresh := mThisPlugin.AddSubMenuItem("Refresh", "Refresh script")
	go func() {
		for range mRefresh.ClickedCh {
			r.log.Info("Requesting refresh")
			ctx := context.Background()
			r.lock.Lock()
			r.refresh(ctx, false)
			r.lock.Unlock()
			r.log.Info("Finished refresh request")
		}
	}()

	mOpen := mThisPlugin.AddSubMenuItem("Edit plugin script", "edit script")
	go func() {
		for range mOpen.ClickedCh {
			r.log.Info("Requesting open file")
			r.lock.Lock()
			if err := osEditFile(r.plugin.Command); err != nil {
				r.log.WithError(err).Warn("failed to open file")
			}
			r.lock.Unlock()
			r.log.Info("Finished file open request")
		}
	}()

	// TODO - needed?
	mRestart := mThisPlugin.AddSubMenuItem("Restart plugin", "Restart plugin")
	go func() {
		for range mRestart.ClickedCh {
			r.log.Info("Requesting plugin restart")
			r.lock.Lock()
			r.sendIPC(&IPCMessage{Type: msgPluginRestartme})
			r.lock.Unlock()
			r.log.Info("Finished restart request")
		}
	}()

}
