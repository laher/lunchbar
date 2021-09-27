package main

import (
	"errors"
	"os"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/prog"
)

func elvish(bin string, out, stderr *os.File, args []string) ([]string, error) {
	f, err := os.ReadFile(bin)
	if err != nil {
		return []string{}, err
	}
	s := parse.Source{Name: bin, Code: string(f), IsFile: true}
	e := eval.NewEvaler()
	capture, fetcher, err := eval.StringCapturePort()
	if err != nil {
		return []string{}, err
	}
	cfg := eval.EvalCfg{
		PutInFg: true,
		Ports:   []*eval.Port{eval.DummyInputPort, capture, eval.DummyOutputPort}, // TODO stop using dummy output
	}

	/* load env
	variable := eval.MakeVarFromName(name)
	err := variable.Set(val)
	if err != nil {
		return err
	}
	*/

	err = e.Eval(s, cfg)
	if err != nil {
		return []string{}, err
	}

	return fetcher(), nil

	/*
		os.Exit(prog.Run(
			[3]*os.File{in, out, stderr}, args,
			buildinfo.Program, daemonStub{}, shell.Program{}))
	*/
	// ? prog.Composite(buildinfo.Program, daemonStub{}, shell.Program{})))
}

var errNoDaemon = errors.New("daemon is not supported in this build")

type daemonStub struct{}

func (daemonStub) ShouldRun(f *prog.Flags) bool { return f.Daemon }

func (daemonStub) Run(fds [3]*os.File, f *prog.Flags, args []string) error {
	return errNoDaemon
}
