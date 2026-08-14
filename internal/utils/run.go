package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type RunArgs struct {
	Command  string
	Args     []string
	Env      []string
	AddOsEnv bool
}

func Run(runargs RunArgs) error {
	cmd := exec.Command(runargs.Command, runargs.Args...)
	var env []string
	if runargs.AddOsEnv {
		env = append(env, os.Environ()...)
	}
	cmd.Env = append(env, runargs.Env...)

	fmt.Printf("Running command: %s %s\n", runargs.Command, strings.Join(runargs.Args, " "))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func RunCapture(runargs RunArgs) (string, error) {
	cmd := exec.Command(runargs.Command, runargs.Args...)
	var env []string
	if runargs.AddOsEnv {
		env = append(env, os.Environ()...)
	}
	cmd.Env = append(env, runargs.Env...)

	fmt.Printf("Running command: %s %s\n", runargs.Command, strings.Join(runargs.Args, " "))

	stdoutStderr, err := cmd.CombinedOutput()
	return string(stdoutStderr), err
}
