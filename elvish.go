package main

import (
	"errors"
	"os"

	"src.elv.sh/pkg/buildinfo"
	"src.elv.sh/pkg/prog"
	"src.elv.sh/pkg/shell"
)

func elvish() {
	os.Exit(prog.Run(
		[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, os.Args,
		buildinfo.Program, daemonStub{}, shell.Program{}))
	// ? prog.Composite(buildinfo.Program, daemonStub{}, shell.Program{})))
}

var errNoDaemon = errors.New("daemon is not supported in this build")

type daemonStub struct{}

func (daemonStub) ShouldRun(f *prog.Flags) bool { return f.Daemon }

func (daemonStub) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	return errNoDaemon
}
