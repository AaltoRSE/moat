package types

// MoatEnv describes a single named environment: its fake home directory,
// its mounts, and its behavior overrides. OverlayMounts lists the
// container-side paths that get a writable overlay (see the runtime's
// overlayowner/overlaydriver/overlaydir settings). Name is not part of the
// configuration: it is set by the config lookup functions so that runtime
// engines can key per-environment state (e.g. externally owned overlay
// directories) by environment name.
type MoatEnv struct {
	Home           string   `validate:"required,dirpath"`
	Mounts         []string `yaml:"mounts,omitempty"`
	ReadOnlyMounts []string `yaml:"readonlymounts,omitempty"`
	OverlayMounts  []string `yaml:"overlaymounts,omitempty"`
	MountCWD       *bool    `yaml:"mountcwd,omitempty"`
	Runtime        *string  `yaml:"runtime,omitempty"`
	PassEnv        *bool    `yaml:"passenv,omitempty"`
	Command        *string  `yaml:"command,omitempty"`
	Name           string   `yaml:"-" mapstructure:"-"`
}
