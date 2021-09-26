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
)

// TODO - OS-dependent dir? (xbar uses ~/Library/Application\ Support )
func rootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "crossbar")
}

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
		r := newPluginRunner(bin, sc)
		go r.Listen()
		systray.Run(r.init(), r.onExit)
	} else {
		sc, err := ipc.StartServer("crossbar", nil)
		if err != nil {
			log.Errorf("could not start IPC server: %s", err)
			return
		}
		s, err := newSupervisor(sc)
		if err != nil {
			log.Errorf("could not set up supervisor: %s", err)
			return
		}
		go s.Listen()
		s.Start()

		// TODO refresh occasionally?
		return
	}
}

func appMenu(mItem *systray.MenuItem) {
	mItem.AddSubMenuItem("Refresh", "")
	mItem.AddSubMenuItem("Open plugins dir", "")

}
