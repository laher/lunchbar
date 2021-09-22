package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/apex/log"
	ipc "github.com/james-barrow/golang-ipc"

	"github.com/matryer/xbar/pkg/plugins"
	"github.com/pkg/errors"
)

var (
	configFile = filepath.Join(rootDir(), "crossbar.config.json")
)

type supervisor struct {
	lock   sync.Mutex
	config *config
	ipc    *ipc.Server
}

func newSupervisor(ipc *ipc.Server) (*supervisor, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadConfig")
	}
	s := &supervisor{
		config: config,
		ipc:    ipc,
	}
	return s, nil
}

func (s *supervisor) Listen() {
	for {
		log.Infof("listen for messages ")
		m, err := s.ipc.Read()
		if err != nil {
			log.Errorf("IPC server error %s", err)
			break
		}
		if m.MsgType > 0 {
			log.Infof("Server recieved %s - Message type: %s", string(m.Data), m.MsgType)
		}
	}
}

func (s *supervisor) Start() {
	if err := os.MkdirAll(pluginsDir(), 0777); err != nil {
		log.Warnf("failed to create plugin directory: %s", err)
	}
	s.StartAll()
	s.Listen()
}

func (s *supervisor) StartAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	pls, err := plugins.Dir(pluginsDir())
	if err != nil {
		log.Warnf("Error loading plugins", err)
	}
	ctx := context.Background()
	commandExec, err := os.Executable()
	if err != nil {
		log.Warnf("Error getting current exe")
		return
	}

	for _, plu := range pls {
		plugin := plu
		log.Infof("starting %s %s", commandExec, filepath.Base(plugin.Command))
		go func(plugin *plugins.Plugin) {
			cmd := exec.CommandContext(ctx, commandExec, plugin.Command)
			cmd.Dir = pluginsDir()
			cmd.Stderr = os.Stdout
			err := cmd.Run()
			if err != nil {
				log.Errorf("error running %s: %s,%s", commandExec, err)
				return
			}
		}(plugin)
	}
}
