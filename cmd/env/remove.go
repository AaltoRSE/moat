/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/AaltoRSE/shark-tank/internal/env"
	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/cobra"
)

func init() {

	var name string

	// removeCmd represents the remove command
	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove an environment from shark-tank",
		Long: `Remove an existing shark-tank environment.

This command allows you to delete and remove a previously created environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Print("remove called")
			if err := env.RemoveEnvironment(types.SharkEnv{Name: name}); err != nil {
				log.Print("Error removing environment: %v\n", err)
				return
			}
			log.Print("Environment successfully removed.\n")
		},
	}

	removeCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment to remove")

	if err := removeCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	envCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
