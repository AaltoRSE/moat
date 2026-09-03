/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/AaltoRSE/moat/cmd"
	_ "github.com/AaltoRSE/moat/cmd/config"
	_ "github.com/AaltoRSE/moat/cmd/env"
	_ "github.com/AaltoRSE/moat/cmd/run"
)

func main() {
	cmd.Execute()
}
