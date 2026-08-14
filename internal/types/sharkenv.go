package types

type SharkEnv struct {
	Home           string   `validate:"required"`
	Mounts         []string `validate:"required,dive,required"`
	ReadOnlyMounts []string
	Runtime        string
	PassEnv        *bool
}
