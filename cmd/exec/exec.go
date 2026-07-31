/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/AaltoRSE/shark-tank/cmd"
	runtimes "github.com/AaltoRSE/shark-tank/internal/runtimes"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in shark-tank",
	Long: `Execute a command in shark-tank.

This command allows you to run a specific command within the shark-tank environment.`,
	DisableFlagParsing: true,

	Run: func(cmd *cobra.Command, args []string) {
		log.Print("exec called with args:", args)

		var runtime runtimes.Runtime
		var err error

		runtime, err = runtimes.GetRuntime("apptainerinstance")
		log.Print(runtime)
		if err == nil {
			runtime.Exec()
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// execCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// execCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
