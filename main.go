package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	ipc "github.com/james-barrow/golang-ipc"
	"github.com/matryer/xbar/pkg/plugins"
)

func rootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "crossbar")
}

// TODO - OS-dependent dir?
func pluginsDir() string {
	return filepath.Join(rootDir(), "plugins")
}

func isExecutable(fi os.FileMode) bool {
	return fi.Perm()&0111 != 0
}

type state struct {
	config config
	lock   sync.Mutex
}

func main() {
	flag.Parse()
	plugin := ""
	// for now don't do any flag parsing
	if len(os.Args) > 1 {
		plugin = os.Args[1]
	}
	pluginsDir := pluginsDir()
	st := &state{}
	if plugin != "" {
		sc, err := ipc.StartClient("crossbar", nil)
		if err != nil {
			log.Errorf("could not start IPC client: %s", err)
			return
		}
		bin := plugin
		if !strings.Contains(plugin, "/") {
			bin = filepath.Join(pluginsDir, plugin)
		}
		p := plugins.NewPlugin(bin)
		r := pluginRunner{plugin: p, st: st, ipc: sc}
		log.Infof("launching systray icon")
		systray.Run(r.init(), r.onExit)
	} else {
		sc, err := ipc.StartServer("crossbar", nil)
		if err != nil {
			log.Errorf("could not start IPC server: %s", err)
			return
		}
		log.Infof("Loading plugins from %s", pluginsDir)
		s, err := newSupervisor(sc)
		if err != nil {
			log.Errorf("could not set up supervisor: %s", err)
			return
		}
		s.Start()
		// TODO refresh occasionally?
		return
	}
}

func appMenu(mItem *systray.MenuItem) {
	mItem.AddSubMenuItem("Refresh", "")
	mItem.AddSubMenuItem("Open plugins dir", "")

}
