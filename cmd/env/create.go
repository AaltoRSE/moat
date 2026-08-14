/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"strings"

	"github.com/AaltoRSE/shark-tank/internal/env"
	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/cobra"
)

var name string
var home string
var mountString string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new shark-tank environment",
	Long: `Create a new shark-tank environment.

This command allows you to create and configure a new environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := env.CreateEnvironment(name, types.SharkEnv{
			FakeHome:       home,
			Mounts:         strings.Split(mountString, ","),
			ReadOnlyMounts: []string{},
		}); err != nil {
			log.Fatalf("could not create environment: %v", err)
		}
	},
}

func init() {
	createCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment")
	createCmd.Flags().StringVarP(&home, "home", "H", "", "Home directory for the environment")
	createCmd.Flags().StringVarP(&mountString, "mounts", "m", "", "Comma-separated list of project mounts")

	if err := createCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := createCmd.MarkFlagRequired("home"); err != nil {
		panic(err)
	}
	if err := createCmd.MarkFlagRequired("mounts"); err != nil {
		panic(err)
	}
	envCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
