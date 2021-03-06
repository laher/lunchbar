package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/matryer/xbar/pkg/plugins"
)

const (
	osWindows = "windows"
	osLinux   = "linux"
	osDarwin  = "darwin"
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

// edit a file
// don't use a bare 'open' call directly, because it might run the thing instead of opening it
func osEditFile(filename string) error {
	editor := os.Getenv("LUNCHBAR_EDITOR")
	if editor == "" {
		log.Warn("LUNCHBAR_EDITOR not set ...")
	}

	ctx := context.Background()
	if editor == "" {
		// open in the default way
		switch runtime.GOOS {
		case osDarwin:
			// equivalent to 'open -a TextEdit file.txt'
			cmd := exec.CommandContext(ctx, "open", "-a", "TextEdit", filename)
			plugins.Setpgid(cmd)
			if err := cmd.Run(); err != nil {
				log.Errorf("could not open: %s. PATH=%s", err, os.Getenv("PATH"))
				return err
			}
			log.Info("opened with TextEdit")
			return nil
		case osLinux:
			// assume update-alternatives and gnome-text-editor (this should cover most Debian/Ubuntu and RedHat/Fedora flavours)
			// TODO other linux variants?
			mimeCmd := exec.CommandContext(ctx, "update-alternatives", "--list", "gnome-text-editor")
			buf := bytes.Buffer{}
			mimeCmd.Stdout = &buf
			plugins.Setpgid(mimeCmd)
			if err := mimeCmd.Run(); err != nil {
				return err
			}
			editors := strings.TrimSpace(buf.String())
			editorArr := strings.Split(editors, "\n")
			cmd := exec.CommandContext(ctx, editorArr[0], filename) //nolint:gosec
			plugins.Setpgid(cmd)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		case osWindows:
			// TODO ... notepad seems bad. Try Code/notepad++/?
			cmd := exec.CommandContext(ctx, "notepad.exe", filename)
			plugins.Setpgid(cmd)
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		default:
			return errUnsupportedPlatform
		}
	} else {
		cmd := exec.CommandContext(ctx, editor, filename)
		plugins.Setpgid(cmd)
		if err := cmd.Run(); err != nil {
			log.WithError(err).WithField("PATH", os.Getenv("PATH")).Errorf("error running " + editor)
			return err
		}
	}
	return nil
}

var errUnsupportedPlatform = fmt.Errorf("unsupported platform")

// TODO timeout? test to be sure
func osOpen(ctx context.Context, href string) error {
	var err error
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case osLinux:
		cmd = exec.CommandContext(ctx, "xdg-open", href)
	case osWindows:
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", href)
	case osDarwin:
		cmd = exec.CommandContext(ctx, "open", href)
	default:
		return errUnsupportedPlatform
	}
	plugins.Setpgid(cmd)
	err = cmd.Run()
	if err != nil {
		log.Warnf("ERR: action href: %s", err)
		return err
	}
	return nil
}
