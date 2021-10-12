package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/apex/log"
	"github.com/laher/lunchbox/lunch"
	"github.com/matryer/xbar/pkg/plugins"
)

func homeDir() string {
	var homeDir string
	if runtime.GOOS == osWindows {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}

// TODO - OS-dependent dir? (xbar uses ~/Library/Application\ Support )
func rootDir() string {
	return filepath.Join(homeDir(), ".config", "lunchbar")
}

func pluginsDir() string {
	return filepath.Join(rootDir(), "plugins")
}

// TODO isExecutable is OS-dependent
// func isExecutable(fi os.FileMode) bool {
//	return fi.Perm()&0111 != 0
// }

// claimAsLunchboxProvider - any child process should use this executable to provide lunchbox functionality
// NOTE if LUNCHBOX_BIN is already set, respect that
func claimAsLunchboxProvider() {
	if os.Getenv("LUNCHBOX_BIN") == "" {
		os.Setenv("LUNCHBOX_BIN", os.Args[0])
	}
}

func main() {
	claimAsLunchboxProvider()
	flag.Parse()

	subcommand := ""
	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}
	ctx := context.Background()
	switch subcommand {
	case "plugin":
		err := runPlugin(ctx, os.Args[2:])
		if err != nil {
			log.Errorf("Error loading plugin: %s", err)
		}
	case "list-plugins":
		pls, err := plugins.Dir(pluginsDir())
		if err != nil {
			log.Warnf("Error loading plugins: %s", err)
		}
		for _, plugin := range pls {
			key := filepath.Base(plugin.Command)
			fmt.Println(key)
		}
	case "", "supervisor":
		s := newSupervisor()
		s.Start()
	default:
		ctx := lunch.Context{
			Ctx: context.Background(),
		}
		lun, ok := lunch.Get(subcommand)

		if !ok {
			fmt.Println("expected a valid subcommand")
			os.Exit(1)
		}
		if err := lun(ctx, os.Args[2:]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
