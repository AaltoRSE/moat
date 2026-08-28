/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	env "github.com/AaltoRSE/moat/cmd/env"
	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage moat environment templates",
	Long: `Manage moat environment templates.

This command allows you to manage your moat environment templates.
You can create and list environment templates.`,
}

func init() {
	env.EnvCmd.AddCommand(templateCmd)
}
