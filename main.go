/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/AaltoRSE/shark-tank/cmd"
	_ "github.com/AaltoRSE/shark-tank/cmd/config"
	_ "github.com/AaltoRSE/shark-tank/cmd/env"
	_ "github.com/AaltoRSE/shark-tank/cmd/run"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cmd.Execute()
}
