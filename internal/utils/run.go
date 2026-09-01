package utils

import (
	"fmt"
	"io"
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

type OutputCapture struct {
	oldStdout *os.File
	readPipe  *os.File
}

func (sc *OutputCapture) StartCapture() {
	sc.oldStdout = os.Stdout
	sc.readPipe, os.Stdout, _ = os.Pipe()
}

func (sc *OutputCapture) StopCapture() (string, error) {
	if sc.oldStdout == nil || sc.readPipe == nil {
		return "", fmt.Errorf("StartCapture not called before StopCapture")
	}
	err := os.Stdout.Close()
	if err != nil {
		return "", err
	}
	os.Stdout = sc.oldStdout
	bytes, err := io.ReadAll(sc.readPipe)
	sc.readPipe = nil
	sc.oldStdout = nil
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
