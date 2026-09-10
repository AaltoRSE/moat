package env

import (
	"fmt"
	"sort"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateEnvListCmd creates the list subcommand for EnvCmd.
func CreateEnvListCmd() *cobra.Command {

	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List moat environments",
		Long: `List moat environments.

This command allows you to view the names of your current moat environments.
Use the -n/--name flag to list only a single environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			// When a name is given, verify the environment exists and print only it
			if name != "" {
				if _, err := config.GetEnv(config.CmdConfig, name, false); err != nil {
					log.Error().Msgf("Error with the environment %q: %v", name, err)
					return
				}
				fmt.Println(name)
				return
			}

			// List envs from viper configuration
			viperEnvs := config.CmdConfig.Sub("envs")
			if viperEnvs == nil {
				fmt.Println("No environments configured.")
				return
			}
			// Unmarshal the environments into a map to get the top-level names
			var envs map[string]types.MoatEnv
			if err := viperEnvs.Unmarshal(&envs); err != nil {
				log.Error().Msgf("Error unmarshalling environments: %v", err)
				return
			}
			// Print the names of the configured environments, sorted for stable output
			names := make([]string, 0, len(envs))
			for name := range envs {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Println(name)
			}
		},
	}

	listCmd.Flags().AddFlagSet(newEnvNameFlagSet())

	return listCmd
}
