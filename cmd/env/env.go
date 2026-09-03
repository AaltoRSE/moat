/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/AaltoRSE/moat/cmd"
	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage moat environments",
	Long: `Manage moat environments.

This command allows you to manage your moat environments.
You can create and list environments.`,
}

func init() {
	cmd.RootCmd.AddCommand(EnvCmd)
}
