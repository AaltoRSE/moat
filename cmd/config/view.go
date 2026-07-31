/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	internal_config "github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/spf13/cobra"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:     "view",
	Aliases: []string{"show"},
	Short:   "view all shark-tank configurations",
	Long: `view all shark-tank configurations.

This command allows you to view all your current shark-tank configurations in a structured format.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(internal_config.GetConfigAsString())
	},
}

func init() {
	configCmd.AddCommand(viewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// viewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// viewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
