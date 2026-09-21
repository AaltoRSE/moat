package types

// RuntimeSpec holds the configuration of a single runtime: its type,
// container image, image cache location, and runtime behavior flags.
type RuntimeSpec struct {
	Type     string `yaml:"type" validate:"required"`
	ImageUrl string `yaml:"imageurl,omitempty" validate:"required_if=Type apptainer"`
	CacheDir string `yaml:"cachedir,omitempty" validate:"required_if=Type apptainer"`
	PassEnv  bool   `yaml:"passenv" validate:"required"`
	MountCWD bool   `yaml:"mountcwd"`
}
