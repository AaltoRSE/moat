/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"list", "view"},
	Short:   "Show the current moat configuration",
	Long: `Show the current moat configuration.

This command allows you to show the current moat configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.GetConfigAsString())
	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
