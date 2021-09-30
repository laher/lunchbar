package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/getlantern/systray"
	ipc "github.com/james-barrow/golang-ipc"
	"github.com/joho/godotenv"
	"github.com/matryer/xbar/pkg/plugins"
)

// TODO - OS-dependent dir? (xbar uses ~/Library/Application\ Support )
func rootDir() string {
	var homeDir string
	if runtime.GOOS == osWindows {
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

func main() {
	pluginPtr := flag.String("plugin", "", "run plugin by name")
	listPlugins := flag.Bool("list", false, "list plugins")
	elvishScriptPtr := flag.String("elvish-script", "", "run an elvish plugin as a standard command")
	elvishShellPtr := flag.Bool("elvish", false, "run an elvish shell prompt")
	flag.Parse()
	if *elvishShellPtr { // load from plugin
		elvishPrompt(append([]string{""}, flag.Args()...))
	} else if *elvishScriptPtr != "" { // load from plugin
		godotenv.Load(filepath.Base(*elvishScriptPtr) + ".env")
		bin := *elvishScriptPtr
		if !strings.Contains(*elvishScriptPtr, "/") {
			bin = filepath.Join(pluginsDir(), *elvishScriptPtr)
		}
		elvishRunScript(bin, os.Stdout, os.Stderr, append([]string{""}, flag.Args()...))
	} else if *pluginPtr != "" { // load from plugin
		godotenv.Load(filepath.Base(*pluginPtr) + ".env")
		key := "crossbar_" + filepath.Base(*pluginPtr)
		sc, err := ipc.StartClient(key, nil)
		if err != nil {
			log.Errorf("could not start IPC client: %s", err)
			return
		}
		bin := *pluginPtr
		if !strings.Contains(*pluginPtr, "/") {
			bin = filepath.Join(pluginsDir(), *pluginPtr)
		}
		r := newPluginRunner(bin, sc)
		go r.Listen()
		systray.Run(r.init(), r.onExit)
	} else if *listPlugins {
		pls, err := plugins.Dir(pluginsDir())
		if err != nil {
			log.Warnf("Error loading plugins: %s", err)
		}
		for _, plugin := range pls {
			key := filepath.Base(plugin.Command)
			fmt.Println(key)
		}
	} else {
		s, err := newSupervisor()
		if err != nil {
			log.Errorf("could not set up supervisor: %s", err)
			return
		}
		s.Start()
	}
}
