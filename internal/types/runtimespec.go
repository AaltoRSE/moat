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

// RuntimeUpdate pairs a RuntimeSpec field with the value to set it to in
// the moat runtime set command. Field is the yaml field name of
// RuntimeSpec (type, imageurl, cachedir, passenv or mountcwd), and Value
// holds the value(s) as given on the command line.
type RuntimeUpdate struct {
	Field string
	Value []string
}
