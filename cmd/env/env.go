package env

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newEnvNameFlagSet returns a new flag group containing the -n/--name flag,
// shared by the subcommands that operate on a single named environment.
func newEnvNameFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("env-name", pflag.ContinueOnError)
	fs.StringP("name", "n", "", "Name of the environment")
	return fs
}

// newEnvVarFlagSet returns a new flag group containing the -H/--home,
// -m/--mount, -r/--ro-mount and -C/--command flags, shared by the
// subcommands that set the variables of an environment.
func newEnvVarFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("env-vars", pflag.ContinueOnError)
	fs.StringP("home", "H", "", "Home directory for the environment")
	fs.StringArrayP("mount", "m", nil, "Mounted directory (can be specified multiple times)")
	fs.StringArrayP("ro-mount", "r", nil, "Read-only mounted directory (can be specified multiple times)")
	fs.StringP("command", "C", "", "Command to run in the environment (use quotes for multi-word commands)")
	return fs
}

func CreateEnvCmd() *cobra.Command {

	// EnvCmd represents the env command
	var EnvCmd = &cobra.Command{
		Use:   "env",
		Short: "Manage moat environments",
		Long: `Manage moat environments.

This command allows you to manage your moat environments.
You can create and list environments.`,
	}

	EnvCmd.AddCommand(CreateEnvCreateCmd())
	EnvCmd.AddCommand(CreateEnvCopyCmd())
	EnvCmd.AddCommand(CreateEnvListCmd())
	EnvCmd.AddCommand(CreateEnvRemoveCmd())
	EnvCmd.AddCommand(CreateEnvSetCmd())
	EnvCmd.AddCommand(CreateEnvShowCmd())
	return EnvCmd
}
