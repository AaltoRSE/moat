package types

type Config struct {
	Defaults Defaults   `validate:"required"`
	Envs     []SharkEnv `validate:"required"`
}

type ApptainerInstanceRuntime struct {
	ImageUrl string `validate:"required"`
	CacheDir string `validate:"required,filepath"`
}

type Defaults struct {
	Runtime       string                   `validate:"required"`
	RuntimeConfig ApptainerInstanceRuntime `validate:"required_if=Runtime apptainerinstance"`
}
