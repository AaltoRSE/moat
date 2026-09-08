/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"

	_ "github.com/AaltoRSE/moat/cmd/config"
	_ "github.com/AaltoRSE/moat/cmd/env"
	"github.com/AaltoRSE/moat/cmd/root"
	_ "github.com/AaltoRSE/moat/cmd/run"
)

func main() {
	err := root.CreateRootCmd().Execute()
	if err != nil {
		fmt.Println("Error encountered while running moat: ", err)
		os.Exit(1)
	}
	os.Exit(0)
}
