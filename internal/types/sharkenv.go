package types

type SharkEnv struct {
	FakeHome       string   `validate:"required"`
	Mounts         []string `validate:"required,dive,required"`
	ReadOnlyMounts []string
	Runtime        string
	PassEnv        *bool
}
