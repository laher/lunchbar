package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	ipc "github.com/james-barrow/golang-ipc"
)

// TODO - OS-dependent dir? (xbar uses ~/Library/Application\ Support )
func rootDir() string {
	var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}

	return filepath.Join(homeDir, ".config", "crossbar")
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
		key := "crossbar_" + filepath.Base(plugin)
		sc, err := ipc.StartClient(key, nil)
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
		s, err := newSupervisor()
		if err != nil {
			log.Errorf("could not set up supervisor: %s", err)
			return
		}
		s.Start()
	}
}

func appMenu(mItem *systray.MenuItem) {
	mItem.AddSubMenuItem("Refresh", "")
	mItem.AddSubMenuItem("Open plugins dir", "")

}
