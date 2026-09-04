package types

type MoatEnv struct {
	Home                  string `validate:"required,dirpath"`
	Mounts                []string
	ReadOnlyMounts        []string
	MountWorkingDirectory *bool
	Runtime               *string
	PassEnv               *bool
}
