package runtimes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/AaltoRSE/shark-tank/internal/utils"
)

func NewApptainerRuntimeSpec(imageUrl string, cacheDir string) *types.RuntimeSpec {
	return &types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: imageUrl,
		CacheDir: cacheDir,
	}
}

func NewApptainerRuntimeFromSpec(spec *types.RuntimeSpec) *ApptainerRuntime {
	return &ApptainerRuntime{
		ImageUrl: spec.ImageUrl,
		CacheDir: spec.CacheDir,
	}
}

type ApptainerRuntime struct {
	ImageUrl string
	CacheDir string
}

type ApptainerImage struct {
	Name string
	Path string
	Url  string
}

func (f *ApptainerRuntime) GetImage() (ApptainerImage, error) {

	cacheDir, err := filepath.Abs(os.ExpandEnv(f.CacheDir))
	if err != nil {
		fmt.Println("Invalid cache directory: ", cacheDir)
		return ApptainerImage{}, err
	}

	if !utils.CheckFolderExists(cacheDir) {
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			fmt.Println("Could not create cache directory: ", cacheDir)
			return ApptainerImage{}, err
		}
	}

	imageUrl := f.ImageUrl

	var name string

	imageUrlSplit := strings.Split(imageUrl, "://")
	if len(imageUrlSplit) > 1 {
		name = imageUrlSplit[1]
	} else {
		name = imageUrl
		imageUrl = "docker://" + imageUrl
	}

	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = name + ".sif"
	path := filepath.Join(cacheDir, name)

	return ApptainerImage{Name: name, Url: imageUrl, Path: path}, nil
}

func (f *ApptainerRuntime) Pull() (string, error) {

	var (
		args []string
	)
	image, err := f.GetImage()
	if err != nil {
		return "", err
	}

	args = append(args, "pull", image.Path, image.Url)

	output, err := utils.RunCapture(utils.RunArgs{Command: "apptainer", Args: args, Env: []string{}, AddOsEnv: true})

	if err != nil && strings.Contains(output, "Image file already exists") {
		fmt.Printf("Image already exists: %s\n", image.Path)
		return image.Path, nil
	}

	if err != nil {
		fmt.Printf("Error pulling image: %v\n", err)
		fmt.Printf("Output: %s\n", output)
		return "", err
	}

	return image.Path, err
}

func (f *ApptainerRuntime) Run(env types.SharkEnv, args []string) (int, error) {
	log.Print("Run called")
	var (
		apptainerArgs []string
		cmdEnv        []string
	)

	if len(args) < 1 {
		return 0, fmt.Errorf("no command provided to run")
	}

	imagePath, err := f.Pull()

	if err != nil {
		return 0, err
	}

	mounts := []string{}

	for _, mount := range env.Mounts {
		mounts = append(mounts, "--bind", mount)
	}

	homeMount := fmt.Sprintf("%s:%s", env.FakeHome, os.Getenv("HOME"))

	// Add mounts and home mount to the apptainer arguments
	apptainerArgs = append(
		[]string{"exec", "--no-home", "--bind", homeMount},
		mounts...,
	)

	// Add the image path to arguments
	apptainerArgs = append(apptainerArgs, imagePath)

	// Add the user command to arguments
	apptainerArgs = append(apptainerArgs, args...)

	cmdEnv = append(os.Environ(),
		"FOO=duplicate_value", // ignored
	)

	fmt.Printf("Running command: apptainer %v\n", apptainerArgs)

	if err := utils.Run(utils.RunArgs{Command: "apptainer", Args: apptainerArgs, Env: cmdEnv, AddOsEnv: true}); err != nil {
		return 0, err
	}
	return 0, nil
}

func (f *ApptainerRuntime) Shell(env types.SharkEnv) (int, error) {
	log.Print("Shell called")
	return 0, nil
}
