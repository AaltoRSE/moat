package runtimes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/AaltoRSE/shark-tank/internal/utils"
	"github.com/spf13/viper"
)

func NewApptainerRuntimeSpec(imageUrl string, cacheDir string) *types.ApptainerRuntimeSpec {
	return &types.ApptainerRuntimeSpec{
		Type:     "apptainer",
		ImageUrl: imageUrl,
		CacheDir: cacheDir,
	}
}

func NewApptainerRuntimeFromSpec(spec *types.ApptainerRuntimeSpec) *ApptainerRuntime {
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

	runtimeConfig := viper.GetStringMapString("defaults.runtimeconfig")
	cacheDir, err := filepath.Abs(os.ExpandEnv(runtimeConfig["cachedir"]))
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

	imageUrl := runtimeConfig["imageUrl"]

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

func (f *ApptainerRuntime) Pull() error {

	var (
		args []string
	)
	image, err := f.GetImage()
	if err != nil {
		return err
	}

	args = append(args, "pull", image.Path, image.Url)

	err = utils.Run(utils.RunArgs{Command: "apptainer", Args: args, Env: []string{}, AddOsEnv: true})
	return err
}

func (f *ApptainerRuntime) Run(env types.SharkEnv, args []string) (int, error) {
	log.Print("Exec called")
	cmd := "echo"
	var (
		cmdArgs []string
		cmdEnv  []string
	)

	if err := f.Pull(); err != nil {
		return 0, err
	}

	cmdArgs = append([]string{"echo"}, args...)
	cmdEnv = append(os.Environ(),
		"FOO=duplicate_value", // ignored
	)
	if err := utils.Run(utils.RunArgs{Command: cmd, Args: cmdArgs, Env: cmdEnv, AddOsEnv: true}); err != nil {
		return 0, err
	}
	return 0, nil
}

func (f *ApptainerRuntime) Shell(env types.SharkEnv) (int, error) {
	log.Print("Shell called")
	return 0, nil
}
