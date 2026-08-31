package runtimes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

func NewApptainerRuntimeSpec(imageUrl string, cacheDir string, passEnv bool) *types.RuntimeSpec {
	return &types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: imageUrl,
		CacheDir: cacheDir,
		PassEnv:  passEnv,
	}
}

func NewApptainerRuntimeFromSpec(spec *types.RuntimeSpec) *ApptainerRuntime {
	return &ApptainerRuntime{
		ImageUrl: spec.ImageUrl,
		CacheDir: spec.CacheDir,
		PassEnv:  spec.PassEnv,
	}
}

type ApptainerRuntime struct {
	ImageUrl string
	CacheDir string
	PassEnv  bool
}

type ApptainerImage struct {
	Name string
	Path string
	Url  string
}

func (f *ApptainerRuntime) GetImage() (ApptainerImage, error) {

	cacheDir, err := utils.SanitizeFolderPath(f.CacheDir)
	if err != nil {
		log.Error().Str("cacheDir", cacheDir).Msg("Invalid cache directory")
		return ApptainerImage{}, err
	}

	if !utils.CheckFolderExists(cacheDir) {
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			log.Error().Str("cacheDir", cacheDir).Msg("Could not create cache directory")
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

func (f *ApptainerRuntime) Pull(passEnv bool) (string, error) {

	var (
		args []string
	)
	image, err := f.GetImage()
	if err != nil {
		return "", err
	}

	args = append(args, "pull", image.Path, image.Url)

	output, err := utils.RunCapture(utils.RunArgs{Command: "apptainer", Args: args, Env: []string{}, PassEnv: passEnv})

	if err != nil && strings.Contains(output, "Image file already exists") {
		log.Debug().Str("imagePath", image.Path).Msg("Image already exists")
		return image.Path, nil
	}

	if err != nil {
		log.Error().Err(err).Str("output", output).Msg("Error pulling image")
		return "", err
	}

	return image.Path, err
}

func (f *ApptainerRuntime) Run(env types.MoatEnv, args []string, envVars []string) (int, error) {
	log.Debug().Msg("Run called")
	var (
		apptainerArgs     []string
		cmdEnv            []string
		source            string
		workingDir        string
		passEnv           bool
		workingDirMounted bool
	)

	if len(args) < 1 {
		return 1, fmt.Errorf("no command provided to run")
	}

	// Check if PassEnv is set in the environment
	if env.PassEnv == nil {
		passEnv = f.PassEnv
	} else {
		passEnv = *env.PassEnv
	}

	imagePath, err := f.Pull(passEnv)

	if err != nil {
		return 1, err
	}

	mounts := []string{}

	workingDir, err = os.Getwd()
	if err != nil {
		return 1, err
	}

	for _, mount := range env.Mounts {
		source = strings.Split(mount, ":")[0]
		if source == workingDir {
			workingDirMounted = true
		}
		mounts = append(mounts, "--bind", mount)
	}

	homeMount := fmt.Sprintf("%s:%s", env.Home, os.Getenv("HOME"))

	// Set base apptainerArgs
	apptainerArgs = []string{"exec", "--no-home", "--no-mount", "cwd"}

	// Add home mounts to apptainer arguments
	apptainerArgs = append(
		apptainerArgs, []string{"--bind", homeMount}...,
	)

	// Add other mounts to the apptainer arguments
	apptainerArgs = append(
		apptainerArgs, mounts...,
	)

	// Mount working directory if requested and not already mounted
	if !workingDirMounted && env.MountWorkingDirectory != nil && *env.MountWorkingDirectory {
		apptainerArgs = append(apptainerArgs, []string{"--bind", workingDir}...)
		// Set working directory to current directory
		apptainerArgs = append(apptainerArgs, []string{"--pwd", workingDir}...)
	}

	// Add the image path to arguments
	apptainerArgs = append(apptainerArgs, imagePath)

	log.Debug().Strs("userArgs", args).Msg("User arguments")
	// Add the user command to arguments
	apptainerArgs = append(apptainerArgs, args...)

	cmdEnv = envVars

	log.Debug().Bool("passEnv", passEnv).Msg("PassEnv")

	if err := utils.Run(utils.RunArgs{Command: "apptainer", Args: apptainerArgs, Env: cmdEnv, PassEnv: passEnv}); err != nil {
		return 1, err
	}
	return 0, nil
}

func (f *ApptainerRuntime) Shell(env types.MoatEnv) (int, error) {
	log.Debug().Msg("Shell called")
	return 0, nil
}
