package types

// RuntimeSpec holds the configuration of a single runtime: its type,
// container executable, container image, image cache location, and runtime
// behavior flags. The executable is optional: when it is not set, the
// runtime engines fall back to their default executables (apptainer for
// the apptainer type, docker for the docker type).
//
// The overlay fields configure how the overlay mounts of the environments
// that use this runtime are provided (apptainer type only): OverlayOwner is
// "internal" (the container engine creates the overlays at run time) or
// "external" (moat pre-creates them on the host under OverlayDir),
// OverlayDriver selects the overlay filesystem driver, and OverlayDir is
// the managed root directory for externally owned overlays.
type RuntimeSpec struct {
	Type          string `yaml:"type" validate:"required"`
	Executable    string `yaml:"executable,omitempty"`
	ImageUrl      string `yaml:"imageurl,omitempty" validate:"required_if=Type apptainer"`
	CacheDir      string `yaml:"cachedir,omitempty" validate:"required_if=Type apptainer"`
	PassEnv       bool   `yaml:"passenv" validate:"required"`
	MountCWD      bool   `yaml:"mountcwd"`
	OverlayOwner  string `yaml:"overlayowner,omitempty" validate:"omitempty,oneof=internal external"`
	OverlayDriver string `yaml:"overlaydriver,omitempty" validate:"omitempty,oneof=fuse-overlayfs kernel-overlayfs btrfs"`
	OverlayDir    string `yaml:"overlaydir,omitempty"`
}

// RuntimeUpdate pairs a RuntimeSpec field with the value to set it to in
// the moat runtime set command. Field is the yaml field name of
// RuntimeSpec (type, executable, imageurl, cachedir, passenv, mountcwd,
// overlayowner, overlaydriver or overlaydir), and Value holds the value(s)
// as given on the command line.
type RuntimeUpdate struct {
	Field string
	Value []string
}
