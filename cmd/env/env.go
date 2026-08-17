/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/shark-tank/cmd"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage shark-tank environments",
	Long: `Manage shark-tank environments.

This command allows you to manage your shark-tank environments.
You can create and list environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Print("env called")
	},
}

func init() {
	cmd.RootCmd.AddCommand(envCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// envCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// envCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
