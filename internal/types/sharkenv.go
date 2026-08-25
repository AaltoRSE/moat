package types

type SharkEnv struct {
	Home           string   `validate:"required,dirpath"`
	Mounts         []string `validate:"required,dive,required"`
	ReadOnlyMounts []string
	Runtime        string
	PassEnv        *bool
}
