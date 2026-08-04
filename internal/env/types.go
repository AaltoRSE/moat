package env

type Environment struct {
	Name           string
	FakeHome       string
	Mounts         []string
	ReadOnlyMounts []string
}
