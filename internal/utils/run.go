package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	fmt.Printf("Running command: %s %s\n", runargs.Command, strings.Join(runargs.Args, " "))

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

	fmt.Printf("Running command: %s %s\n", runargs.Command, strings.Join(runargs.Args, " "))

	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}
