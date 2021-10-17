package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/apex/log"
	ipc "github.com/james-barrow/golang-ipc"

	"github.com/matryer/xbar/pkg/plugins"
)

const (
	msgcodeDefault = 1
)

type supervisor struct {
	lock      sync.Mutex
	ipcs      map[string]*ipc.Server
	log       *log.Entry
	processes []*exec.Cmd
}

func newSupervisor() *supervisor {
	s := &supervisor{
		ipcs:      map[string]*ipc.Server{},
		log:       log.WithField("t", "supervisor").WithField("pid", os.Getpid()),
		processes: []*exec.Cmd{},
	}
	return s
}

const (
	msgSupervisorRefresh      = "refresh"
	msgSupervisorUnrecognised = "unrecognised"
	msgSupervisorQuit         = "quit"
)

func (s *supervisor) Listen(ctx context.Context, key string, ipcs *ipc.Server) {
	s.log.WithField("plugin", key).Infof("listen for messages")
	for {
		m, err := ipcs.Read()
		if err != nil {
			s.log.Errorf("IPC server error %s", err)
			break
		}
		s.log.WithField("from-plugin", key).WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Debug("supervisor received message")
		if m.MsgType < 0 {
			// -2 is error. -1 seems to mean 'ok'
			continue
		}
		switch string(m.Data) {
		case msgPluginRefreshAll:
			// TODO discovering new plugins
			// TODO killing old plugins
			s.broadcast(msgSupervisorRefresh)
		case msgPluginRefreshComplete:
			// nothing to do
		case msgPluginRefreshError:
			// TODO - kill and relaunch, maybe? after a time...
		case msgPluginUnrecognised:
			// TODO - die here? / kill plugin?
			s.log.WithField("from-plugin", key).WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Warn("plugin did not recognise previous command")
		case msgPluginRestartme:
			s.log.Info("send quit request to plugin which requested restart")
			s.sendIPC(key, msgSupervisorQuit)
			// TODO timing. should we wait for plugins to respond?
			time.Sleep(time.Second * 1)
			p := filepath.Join(pluginsDir(), key)
			plugin := plugins.NewPlugin(p)
			s.startPlugin(ctx, plugin)

		case msgPluginQuit:
			s.log.Info("broadcasting quit request to all plugins")
			s.broadcast(msgSupervisorQuit)
			// TODO timing. should we wait for plugins to respond?
			time.Sleep(time.Second * 2)
			s.log.Info("exiting now")
			os.Exit(0)
		default:
			s.log.WithField("from-plugin", key).WithField("messageType", m.MsgType).WithField("body", string(m.Data)).Error("supervisor received a message which is not handled")
			s.sendIPC(msgSupervisorUnrecognised, key)
		}
	}
}

func (s *supervisor) broadcast(m string) {
	for k := range s.ipcs {
		s.sendIPC(m, k)
	}
}

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	s.StartAll(ctx)

	log.Info("main thread running until cancellation of interrupt context")
	<-ctx.Done()
	// TODO should we do any more tidy up?
	log.Info("exit after signal context cancelled")
}

func (s *supervisor) StartAll(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	pls, err := plugins.Dir(pluginsDir())
	if err != nil {
		s.log.Warnf("Error loading plugins: %s", err)
	}
	for _, plugin := range pls {
		s.startPlugin(ctx, plugin)
	}
}

func (s *supervisor) startPlugin(ctx context.Context, plugin *plugins.Plugin) {
	thisExecutable, err := os.Executable()
	if err != nil {
		s.log.Errorf("Error getting current exe")
		return
	}

	key := filepath.Base(plugin.Command)
  s.log.Infof("starting plugin %s: %s", key, thisExecutable)
	sc, err := ipc.StartServer("lunchbar_"+key, nil)
	if err != nil {
		log.Errorf("could not start IPC server: %s", err)
		return
	}
	s.ipcs[key] = sc
	go s.Listen(ctx, key, sc)
	go func(plugin *plugins.Plugin) {
		cmd := exec.CommandContext(ctx, thisExecutable, "plugin", "-plugin", key)
		plugins.Setpgid(cmd) // sets process group id (Unix only). This ensures that the child processes get tidied up
		cmd.Dir = pluginsDir()
		cmd.Stderr = os.Stdout
		s.processes = append(s.processes, cmd)
		// TODO use Start instead?
		cmd.Start()
		err := cmd.Run()
		if err != nil {
			s.log.WithField("plugin", filepath.Base(plugin.Command)).Errorf("error running %s: %s", thisExecutable, err)
			return
		}
	}(plugin)
}
