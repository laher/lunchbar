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
)

type supervisor struct {
	lock sync.Mutex
	ipcs map[string]*ipc.Server
	log  *log.Entry
}

func newSupervisor() (*supervisor, error) {
	s := &supervisor{
		ipcs: map[string]*ipc.Server{},
		log:  log.WithField("t", "supervisor").WithField("pid", os.Getpid()),
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

const (
	msgcodeDefault = 1
)

func (s *supervisor) sendIPC(m, key string) {
	err := s.ipcs[key].Write(msgcodeDefault, []byte(m))
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
		s.log.Warnf("Error loading plugins: %s", err)
	}
	thisExecutable, err := os.Executable()
	if err != nil {
		s.log.Warnf("Error getting current exe")
		return
	}

	for _, plugin := range pls {
		ctx := context.Background()
		key := filepath.Base(plugin.Command)
		s.log.Infof("starting %s %s", thisExecutable, key)
		sc, err := ipc.StartServer("crossbar_"+key, nil)
		if err != nil {
			log.Errorf("could not start IPC server: %s", err)
			return
		}
		s.ipcs[key] = sc
		go s.Listen(key, sc)
		go func(plugin *plugins.Plugin) {
			cmd := exec.CommandContext(ctx, thisExecutable, "-plugin", key)
			cmd.Dir = pluginsDir()
			cmd.Stderr = os.Stdout
			err := cmd.Run()
			if err != nil {
				s.log.WithField("plugin", filepath.Base(plugin.Command)).Errorf("error running %s: %s", thisExecutable, err)
				return
			}
			err = sc.Write(msgcodeDefault, []byte("hello client. I refreshed"))
			if err != nil {
				s.log.Warnf("could not write to client %s", err)
			}
		}(plugin)
	}
}
