/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/AaltoRSE/shark-tank/cmd"
	_ "github.com/AaltoRSE/shark-tank/cmd/config"
	_ "github.com/AaltoRSE/shark-tank/cmd/env"
	_ "github.com/AaltoRSE/shark-tank/cmd/run"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cmd.Execute()
}
