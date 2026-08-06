package types

type RuntimeSpec struct {
	Type string `validate:"required"`
}

type ApptainerRuntimeSpec struct {
	Type     string `validate:"required"`
	ImageUrl string `validate:"required"`
	CacheDir string `validate:"required,filepath"`
}
