package types

type Config struct {
	Defaults Defaults           `validate:"required"`
	Envs     map[string]MoatEnv `validate:"required"`
	Runtimes map[string]RuntimeSpec
}

type Defaults struct {
	Runtime  string `validate:"required"`
	Runtimes map[string]RuntimeSpec
}
