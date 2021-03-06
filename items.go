package main

import (
	"context"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	"github.com/matryer/xbar/pkg/plugins"
)

type itemWrap struct {
	plugItem    *plugins.Item
	trayItem    *systray.MenuItem
	isSeparator bool
	// parent      *itemWrap

	// could override standard xbar behaviour
	action plugins.ActionFunc

	subitems []*itemWrap
}

func (item *itemWrap) DoAction(ctx context.Context) {
	// only handle actions where necessary
	if !item.isSeparator {
		if item.action != nil {
			item.action(ctx)
		} else if item.plugItem != nil && len(item.plugItem.Items) < 1 {
			action := item.plugItem.Action()
			if action != nil {
				log.Debug("run item action")
				action(ctx)
				log.Debug("completed item action")
			}
		}
	}
}
