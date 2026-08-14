package types

type RuntimeSpec struct {
	Type     string `validate:"required"`
	ImageUrl string `validate:"required_if=Type=apptainer"`
	CacheDir string `validate:"required_if=Type=apptainer,filepath"`
}
