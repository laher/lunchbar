package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

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
	log    *log.Entry
}

func newSupervisor(ipc *ipc.Server) (*supervisor, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadConfig")
	}
	s := &supervisor{
		config: config,
		ipc:    ipc,
		log:    log.WithField("t", "supervisor").WithField("pid", os.Getpid()),
	}
	return s, nil
}

func (s *supervisor) Listen() {
	for {
		s.log.Infof("listen for messages ")
		m, err := s.ipc.Read()
		if err != nil {
			s.log.Errorf("IPC server error %s", err)
			break
		}
		s.log.WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Infof("Server received message")
		s.sendIPC("OK", "originator")
	}
}

func (s *supervisor) sendIPC(m string, t string) {
	err := s.ipc.Write(14, []byte(fmt.Sprintf("from:supervisor, to:%s: %s", m, m)))
	if err != nil {
		s.log.Warnf("could not write to client: %s", err)
	}
}

func (s *supervisor) Start() {
	s.log.Infof("Loading plugins from %s", pluginsDir())
	if err := os.MkdirAll(pluginsDir(), 0777); err != nil {
		s.log.Warnf("failed to create plugin directory: %s", err)
	}
	s.StartAll()
	time.Sleep(time.Hour * 24) // TODO refresh plugins list
}

func (s *supervisor) StartAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	pls, err := plugins.Dir(pluginsDir())
	if err != nil {
		s.log.Warnf("Error loading plugins", err)
	}
	ctx := context.Background()
	commandExec, err := os.Executable()
	if err != nil {
		s.log.Warnf("Error getting current exe")
		return
	}

	for _, plu := range pls {
		plugin := plu
		s.log.Infof("starting %s %s", commandExec, filepath.Base(plugin.Command))
		go func(plugin *plugins.Plugin) {
			cmd := exec.CommandContext(ctx, commandExec, plugin.Command)
			cmd.Dir = pluginsDir()
			cmd.Stderr = os.Stdout
			err := cmd.Run()
			if err != nil {
				s.log.Errorf("error running %s: %s,%s", commandExec, err)
				return
			}
			err = s.ipc.Write(14, []byte("hello client. I refreshed"))
			if err != nil {
				s.log.Warnf("could not write to client %s", err)
			}
		}(plugin)
	}
}
