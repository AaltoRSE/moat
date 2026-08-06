package types

type Config struct {
	Defaults Defaults   `validate:"required"`
	Envs     []SharkEnv `validate:"required"`
	Runtimes map[string]RuntimeSpec
}

type Defaults struct {
	Runtime  string `validate:"required"`
	Runtimes map[string]RuntimeSpec
}
