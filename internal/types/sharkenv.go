package types

type SharkEnv struct {
	Name           string   `validate:"required"`
	FakeHome       string   `validate:"required"`
	Mounts         []string `validate:"required,dive,required"`
	ReadOnlyMounts []string
	Runtime        string
}
