package types

type MoatEnv struct {
	Home           string   `validate:"required,dirpath"`
	Mounts         []string `yaml:"mounts,omitempty"`
	ReadOnlyMounts []string `yaml:"readonlymounts,omitempty"`
	Mountcwd       *bool    `yaml:"mountcwd,omitempty"`
	Runtime        *string  `yaml:"runtime,omitempty"`
	PassEnv        *bool    `yaml:"passenv,omitempty"`
	Command        *string  `yaml:"command,omitempty"`
}
