package runtime

import (
	"fmt"
	"sort"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateRuntimeListCmd creates the list subcommand for runtimeCmd.
func CreateRuntimeListCmd() *cobra.Command {

	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List moat runtimes",
		Long: `List moat runtimes.

This command allows you to view the names of your current moat runtimes,
including the default runtimes.
Use the -n/--name flag to list only a single runtime.`,
		Run: func(cmd *cobra.Command, args []string) {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			runtimes := config.GetRuntimes(config.CmdConfig)

			// When a name is given, verify the runtime exists and print only it
			if name != "" {
				if _, ok := runtimes[name]; !ok {
					log.Error().Msgf("Error with the runtime %q: runtime %q not found", name, name)
					return
				}
				fmt.Println(name)
				return
			}

			// Print the names of the available runtimes, sorted for stable output
			names := make([]string, 0, len(runtimes))
			for name := range runtimes {
				names = append(names, name)
			}
			if len(names) == 0 {
				fmt.Println("No runtimes configured.")
				return
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Println(name)
			}
		},
	}

	listCmd.Flags().StringP("name", "n", "", "Name of the runtime")

	return listCmd
}
