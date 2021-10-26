package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/apex/log"

	"github.com/matryer/xbar/pkg/plugins"
)

const (
	msgcodeDefault = 1
)

type supervisor struct {
	lock        sync.RWMutex
	listener    net.Listener
	connections map[string]net.Conn
	//ipcs        map[string]*ipc.Server
	log       *log.Entry
	processes []*exec.Cmd
}

func newSupervisor() *supervisor {
	s := &supervisor{
		connections: map[string]net.Conn{},
		//ipcs:      map[string]*ipc.Server{},
		log:       log.WithField("t", "supervisor").WithField("pid", os.Getpid()),
		processes: []*exec.Cmd{},
	}
	return s
}

func sockAddr() string {
	return filepath.Join(os.TempDir(), "lunchbar.sock")
	//return filepath.Join(os.Getenv("TEMP"), "lunchbar")
}

func (s *supervisor) listenForConnections() error {
	s.log.Info("remove all")
	if err := os.RemoveAll(sockAddr()); err != nil {
		s.log.WithError(err).Error("could not remove all")
		return err
	}
	var err error
	s.listener, err = net.Listen("unix", sockAddr())
	if err != nil {
		return err
	}
	go func() {
		defer s.listener.Close()
		for {
			// Accept new connections, dispatching them to echoServer
			// in a goroutine.
			conn, err := s.listener.Accept()
			if err != nil {
				log.Errorf("accept error:", err)
				os.Exit(1)
			}
			go s.handleConnection(context.Background(), conn)
		}
	}()
	return nil
}

func (s *supervisor) handleConnection(ctx context.Context, conn net.Conn) {
	key := "" // no key yet
	for {
		m := &IPCMessage{}
		err := m.Read(conn)
		if err != nil {
			if err == io.EOF {
				// TODO - restart?
				log.Infof("client closed connection")
				return
			}
			log.Errorf("error %s", err)
			return
		}

		log.Infof("[READ] message: %+v", m)
		if m.Type == msgPluginID {
			log.Infof("plugin connected with ID: %s", string(m.Data))
			key = string(m.Data)
			s.lock.Lock()
			s.connections[key] = conn
			s.lock.Unlock()
		} else {
			/*
				s := strings.ToUpper(string(m.Data))
				m.Length = len(s)
				m.Data = []byte(s)

				err = m.Write(conn)
				if err != nil {
					log.Errorf("Error writing %s", err)
					break
				}

				fmt.Println("[WRITE] ", m)
			*/
			s.handleMessage(ctx, key, m)
		}
	}
}

func (s *supervisor) handleMessage(ctx context.Context, key string, m *IPCMessage) {
	s.log.WithField("from-plugin", key).WithField("messageType", m.Type).WithField("body", string(m.Data)).Debug("supervisor received message")
	switch m.Type {
	case msgPluginRefreshAll:
		// TODO discovering new plugins
		// TODO killing old plugins
		s.broadcast(&IPCMessage{Type: msgSupervisorRefresh})
	case msgPluginRefreshComplete:
		// nothing to do
	case msgPluginRefreshError:
		// TODO - kill and relaunch, maybe? after a time...
	case msgPluginUnrecognised:
		// TODO - die here? / kill plugin?
		s.log.WithField("from-plugin", key).WithField("messageType", m.Type).WithField("body", string(m.Data)).Warn("plugin did not recognise previous command")
	case msgPluginRestartme:
		s.log.Info("send quit request to plugin which requested restart")
		s.sendIPC(key, &IPCMessage{Type: msgSupervisorQuit})
		// TODO timing. should we wait for plugins to respond?
		time.Sleep(time.Second * 1)
		p := filepath.Join(pluginsDir(), key)
		plugin := plugins.NewPlugin(p)
		s.startPlugin(ctx, plugin)

	case msgPluginQuit:
		s.log.Info("broadcasting quit request to all plugins")
		s.broadcast(&IPCMessage{Type: msgSupervisorQuit})
		// TODO timing. should we wait for plugins to respond?
		time.Sleep(time.Second * 2)
		s.log.Info("exiting now")
		os.Exit(0)
	default:
		s.log.WithField("from-plugin", key).WithField("messageType", m.Type).WithField("body", string(m.Data)).Error("supervisor received a message which is not handled")
		s.sendIPC(key, &IPCMessage{Type: msgSupervisorUnrecognised})
	}
}

func (s *supervisor) broadcast(m *IPCMessage) {
	s.lock.RLock()
	keys := []string{}
	for k := range s.connections {
		keys = append(keys, k)
	}
	s.lock.RUnlock()
	for _, k := range keys {
		s.sendIPC(k, m)
	}
}

func (s *supervisor) sendIPC(k string, m *IPCMessage) error {
	s.lock.RLock()
	c, ok := s.connections[k]
	s.lock.RUnlock()
	if ok {
		s.log.Infof("Sending quit message to plugin: %s", k)
		if err := m.Write(c); err != nil {
			// TODO tidyup?
			s.log.Warnf("could not write to client: %s", err)
			return err
		}
	} else {
		s.log.Warnf("could not find a connection for plugin: %s", k)
	}
	return nil
}

func (s *supervisor) Start() {
	s.log.Infof("Loading plugins from %s", pluginsDir())
	if err := os.MkdirAll(pluginsDir(), 0777); err != nil {
		s.log.Warnf("failed to create plugin directory: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		err := s.listenForConnections()
		if err != nil {
			s.log.WithError(err).Errorf("could not listen for connections %s", sockAddr())
			os.Exit(1)
		}
	}()
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
	//go s.Listen(ctx, key) // restart ?
	go func(plugin *plugins.Plugin) {
		cmd := exec.CommandContext(ctx, thisExecutable, "plugin", "-plugin", key)
		plugins.Setpgid(cmd) // sets process group id (Unix only). This ensures that the child processes get tidied up
		cmd.Dir = pluginsDir()
		cmd.Stderr = os.Stdout
		s.processes = append(s.processes, cmd)
		// TODO use Start instead?
		err := cmd.Run()
		if err != nil {
			s.log.WithField("plugin", filepath.Base(plugin.Command)).Errorf("error running %s: %s", thisExecutable, err)
			return
		}
	}(plugin)
}
