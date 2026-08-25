/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/AaltoRSE/shark-tank/cmd"
	_ "github.com/AaltoRSE/shark-tank/cmd/config"
	_ "github.com/AaltoRSE/shark-tank/cmd/env"
	_ "github.com/AaltoRSE/shark-tank/cmd/env/template"
	_ "github.com/AaltoRSE/shark-tank/cmd/run"
)

func main() {
	cmd.Execute()
}
