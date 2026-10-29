package env

import (
	"fmt"
	"strconv"

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
// -m/--mount, -r/--ro-mount, -C/--command and --mountcwd flags, shared by
// the subcommands that set the variables of an environment.
func newEnvVarFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("env-vars", pflag.ContinueOnError)
	fs.StringP("home", "H", "", "Home directory for the environment")
	fs.StringArrayP("mount", "m", nil, "Mounted directory (can be specified multiple times)")
	fs.StringArrayP("ro-mount", "r", nil, "Read-only mounted directory (can be specified multiple times)")
	fs.StringP("command", "C", "", "Command to run in the environment (use quotes for multi-word commands)")
	fs.String("mountcwd", "", "Bind-mount the caller's current working directory into the container (true or false, empty means true)")
	return fs
}

// getMountCWDValue returns the value of the --mountcwd flag as a pointer to
// a boolean. It returns nil when the flag was not given, meaning that the
// environment does not define mountcwd of its own and falls back to the
// setting of the runtime it uses. An empty flag value defaults to true;
// any other value must be a valid boolean string. It returns an error if
// the flag value is not a valid boolean.
func getMountCWDValue(cmd *cobra.Command) (*bool, error) {
	if !cmd.Flags().Changed("mountcwd") {
		return nil, nil
	}
	value, err := cmd.Flags().GetString("mountcwd")
	if err != nil {
		return nil, fmt.Errorf("could not get mountcwd flag: %w", err)
	}
	if value == "" {
		enabled := true
		return &enabled, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value %q for --mountcwd: %w", value, err)
	}
	return &parsed, nil
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
