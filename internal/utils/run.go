package utils

import (
	"log"
	"os"
	"os/exec"
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

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
