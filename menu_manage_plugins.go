package main

import (
	"context"
	"path/filepath"

	xbarplugins "github.com/laher/lunchbox/xbar-plugins"
)

func (r *pluginRunner) menuManagePlugins() {

	mManagePlugins := r.mainItem.AddSubMenuItem("Manage plugins", "Add and manage plugins")
	// open dir to create a file
	mOpenDir := mManagePlugins.AddSubMenuItem("Open plugin scripts dir", "open scripts dir to manage plugin files")
	go func() {
		for range mOpenDir.ClickedCh {
			dir := filepath.Dir(r.plugin.Command)
			r.log.WithField("dir", dir).Info("Requesting open dir")
			r.lock.Lock()
			ctx := context.Background()
			if err := osOpen(ctx, dir); err != nil {
				r.log.WithError(err).Warn("failed to open dir")
			}
			r.lock.Unlock()
			r.log.Info("Finished open dir request")
		}
	}()
	mBrowseXbar := mManagePlugins.AddSubMenuItem("Browse Xbar plugins", "browse to xbar plugins")
	go func() {
		for range mBrowseXbar.ClickedCh {
			r.log.Info("Browse to xbar plugins")
			ctx := context.Background()
			r.lock.Lock()
			if err := osOpen(ctx, "https://github.com/matryer/xbar-plugins"); err != nil {
				r.log.WithError(err).Warn("failed to open link")
			}
			r.lock.Unlock()
		}
	}()

	mRefreshAll := mManagePlugins.AddSubMenuItem("Refresh All", "Refresh all plugins")
	go func() {
		for range mRefreshAll.ClickedCh {
			r.log.Info("Requesting refresh-all")
			ctx := context.Background()
			r.lock.Lock()
			r.refreshAll(ctx, false)
			r.lock.Unlock()
			r.log.Info("Finished refresh-all request")
		}
	}()
	mManagePlugins.AddSubMenuItem("", "")
	mTitle := mManagePlugins.AddSubMenuItem("Install a bundled plugin:", "Install a bundled plugin")
	mTitle.Disable()

	pis, err := xbarplugins.List(context.TODO())
	if err != nil {
		// nothing. dont add any
		r.log.WithError(err).Warn("failed to list available plugins")
	} else {
		for _, pi := range pis {
			pluginItem := mManagePlugins.AddSubMenuItem(pi, pi)
			go func(pi string) {
				for range pluginItem.ClickedCh {
					r.log.WithField("for-plugin", pi).Info("todo - install plugin")
				}
			}(pi)
		}
	}
}
