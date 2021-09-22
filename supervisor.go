package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/apex/log"

	"github.com/matryer/xbar/pkg/plugins"
	"github.com/pkg/errors"
)

var (
	configFile = filepath.Join(rootDir(), "crossbar.config.json")
)

type supervisor struct {
	lock   sync.Mutex
	config *config
}

func newSupervisor() (*supervisor, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadConfig")
	}
	s := &supervisor{
		config: config,
	}
	return s, nil
}

func (s *supervisor) Start() {
	if err := os.MkdirAll(pluginsDir(), 0777); err != nil {
		log.Warnf("failed to create plugin directory: %s", err)
	}
	s.RefreshAll()
}

func (s *supervisor) RefreshAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	pls, err := plugins.Dir(pluginsDir())
	if err != nil {
		log.Warnf("Error loading plugins", err)
	}
	ctx := context.Background()
	for _, plugin := range pls {
		plugin.Refresh(ctx)
		// TODO handle errors
		if len(plugin.Items.CycleItems) > 0 {
			s := plugin.Items.CycleItems[0].DisplayText()
			log.Infof("Loading plugin %s", s)
		} else {
			log.Infof("Not loading plugin %s", plugin.Command)
		}
	}
}
