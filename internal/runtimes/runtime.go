package runtimes

import (
	"errors"
	"log"

	"github.com/AaltoRSE/shark-tank/internal/types"
)

type Runtime interface {
	Exec(env types.SharkEnv, args []string) (int, error)
	Shell(env types.SharkEnv) (int, error)
	Start(env types.SharkEnv) (int, error)
	Stop(env types.SharkEnv) (int, error)
}

func GetRuntime(name string) (Runtime, error) {
	if name == "apptainerinstance" {
		log.Println("Returning apptainer instance")
		return &ApptainerInstanceRuntime{}, nil
	}

	return nil, errors.New("No runtime found")
}
