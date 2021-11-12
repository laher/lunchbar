package main

import (
	"context"
	"io"
	"os"
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
			r.refreshAll(ctx)
			r.lock.Unlock()
			r.log.Info("Finished refresh-all request")
		}
	}()
	mManagePlugins.AddSubMenuItem("", "")
	mTitle := mManagePlugins.AddSubMenuItem("Install a bundled plugin:", "Install a bundled plugin")
	mTitle.Disable()
	ctx := context.TODO()

	pis, err := xbarplugins.List(ctx)
	if err != nil {
		// nothing. dont add any
		r.log.WithError(err).Warn("failed to list available plugins")
	} else {
		for _, pi := range pis {
			dest := filepath.Join(pluginsDir(), filepath.Base(pi))
			if _, err := os.Stat(dest); err != nil {
				// doesnt exist
				r.log.WithError(err).Warn("plugin file doesnt exist")
				continue
			}
			pluginItem := mManagePlugins.AddSubMenuItem(pi, pi)
			go func(pi string) {
				for range pluginItem.ClickedCh {
					if err := r.installPlugin(ctx, pi); err != nil {
						r.log.WithError(err).Error("failed to install plugin")
					}
				}
			}(pi)
		}
	}
}

func (r *pluginRunner) installPlugin(ctx context.Context, pi string) error {
	rdr, err := xbarplugins.Open(ctx, pi)
	if err != nil {
		return err
	}
	dest := filepath.Join(pluginsDir(), filepath.Base(pi))
	w, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, rdr); err != nil {
		w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	r.log.WithField("for-plugin", pi).Info("installed plugin")
	return nil
}
