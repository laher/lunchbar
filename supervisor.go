package main

import (
	"context"
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
	ipcs   map[string]*ipc.Server
	log    *log.Entry
}

func newSupervisor() (*supervisor, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadConfig")
	}
	s := &supervisor{
		config: config,
		ipcs:   map[string]*ipc.Server{},
		log:    log.WithField("t", "supervisor").WithField("pid", os.Getpid()),
	}
	return s, nil
}

func (s *supervisor) Listen(key string, ipcs *ipc.Server) {
	for {
		s.log.WithField("plugin", key).Infof("listen for messages")
		m, err := ipcs.Read()
		if err != nil {
			s.log.Errorf("IPC server error %s", err)
			break
		}
		s.log.WithField("plugin", key).WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Infof("Server received message")
		s.sendIPC("OK", key)
	}
}

func (s *supervisor) sendIPC(m string, key string) {
	err := s.ipcs[key].Write(14, []byte(m))
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
	time.Sleep(time.Hour * 24) // TODO refresh plugins list indefinitely instead
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

	for _, plugin := range pls {
		key := filepath.Base(plugin.Command)
		s.log.Infof("starting %s %s", commandExec, key)
		sc, err := ipc.StartServer("crossbar_"+key, nil)
		if err != nil {
			log.Errorf("could not start IPC server: %s", err)
			return
		}
		s.ipcs[key] = sc
		go s.Listen(key, sc)
		go func(plugin *plugins.Plugin) {
			cmd := exec.CommandContext(ctx, commandExec, plugin.Command)
			cmd.Dir = pluginsDir()
			cmd.Stderr = os.Stdout
			err := cmd.Run()
			if err != nil {
				s.log.Errorf("error running %s: %s,%s", commandExec, err)
				return
			}
			err = sc.Write(14, []byte("hello client. I refreshed"))
			if err != nil {
				s.log.Warnf("could not write to client %s", err)
			}
		}(plugin)
	}
}
