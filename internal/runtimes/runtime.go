package runtimes

import (
	"errors"
	"log"
)

type Runtime interface {
	Exec() (int, error)
	Shell() (int, error)
	Start() (int, error)
	Stop() (int, error)
}

func GetRuntime(name string) (Runtime, error) {
	if name == "apptainerinstance" {
		log.Println("Returning apptainer instance")
		return &ApptainerInstanceRuntime{}, nil
	}

	return nil, errors.New("No runtime found")
}
