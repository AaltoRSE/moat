package utils

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

type RunArgs struct {
	Command string
	Args    []string
	Env     []string
	PassEnv bool
}

func Run(runargs RunArgs) error {
	cmd := exec.Command(runargs.Command, runargs.Args...)
	if runargs.PassEnv {
		cmd.Env = append(os.Environ(), runargs.Env...)
	} else {
		cmd.Env = runargs.Env
	}

	log.Debug().Str("command", runargs.Command).Strs("args", runargs.Args).Msg("Running command")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func RunCapture(runargs RunArgs) (string, error) {
	cmd := exec.Command(runargs.Command, runargs.Args...)
	if runargs.PassEnv {
		cmd.Env = append(os.Environ(), runargs.Env...)
	} else {
		cmd.Env = runargs.Env
	}

	log.Debug().Str("command", runargs.Command).Strs("args", runargs.Args).Msg("Running command")

	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}
