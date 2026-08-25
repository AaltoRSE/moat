/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	env "github.com/AaltoRSE/shark-tank/cmd/env"
	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage shark-tank environment templates",
	Long: `Manage shark-tank environment templates.

This command allows you to manage your shark-tank environment templates.
You can create and list environment templates.`,
}

func init() {
	env.EnvCmd.AddCommand(templateCmd)
}
