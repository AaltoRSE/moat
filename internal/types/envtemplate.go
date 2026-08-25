package types

type EnvTemplate struct {
	HomeBase       string `validate:"required,dirpath"`
	Mounts         []string
	ReadOnlyMounts []string
	Runtime        string
	PassEnv        *bool
}
