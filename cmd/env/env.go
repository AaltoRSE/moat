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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// EnvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// EnvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
