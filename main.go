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
	elvishPluginPtr := flag.String("elvish", "", "run an elvish plugin as a standard command")
	flag.Parse()
	if *elvishPluginPtr != "" { // load from plugin
		godotenv.Load(filepath.Base(*elvishPluginPtr) + ".env")
		bin := *elvishPluginPtr
		if !strings.Contains(*elvishPluginPtr, "/") {
			bin = filepath.Join(pluginsDir(), *elvishPluginPtr)
		}
		/*
			f, err := os.Open(bin)
			//f, err := os.ReadFile(bin)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		*/
		elvish(bin, os.Stdout, os.Stderr, append([]string{""}, flag.Args()...))
		/*
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, bin)
			cmd.Dir = pluginsDir()
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		*/
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
