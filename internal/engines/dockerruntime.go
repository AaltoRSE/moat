package engines

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

func NewDockerRuntimeFromSpec(spec *types.RuntimeSpec) *DockerRuntime {
	return &DockerRuntime{
		ImageUrl: spec.ImageUrl,
		CacheDir: spec.CacheDir,
		PassEnv:  spec.PassEnv,
		MountCWD: spec.MountCWD,
	}
}

type DockerRuntime struct {
	ImageUrl string
	CacheDir string
	PassEnv  bool
	MountCWD bool
}

func (f *DockerRuntime) Pull(passEnv bool) (string, error) {
	if err := utils.Run(utils.RunArgs{
		Command: "docker",
		Args:    []string{"pull", f.ImageUrl},
		Env:     []string{},
		PassEnv: passEnv,
	}); err != nil {
		log.Error().Err(err).Str("imageUrl", f.ImageUrl).Msg("Error pulling image")
		return "", err
	}

	return f.ImageUrl, nil
}

func (f *DockerRuntime) Run(env types.MoatEnv, args []string, envVars []string) (int, error) {
	log.Debug().Msg("Run called")
	var (
		dockerArgs        []string
		cmdEnv            []string
		workingDir        string
		passEnv           bool
		mountCWD          bool
		workingDirMounted bool
	)

	if len(args) < 1 {
		return 1, fmt.Errorf("no command provided to run")
	}

	// Run as the current user
	userArgs := []string{"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())}

	cacheDir, err := utils.SanitizeFolderPath(f.CacheDir)

	if err != nil {
		return 1, err
	}

	cacheMount := fmt.Sprintf("type=bind,source=%s,target=%s", cacheDir, cacheDir)

	passWdTemp, err := os.CreateTemp(cacheDir, "passwd.*")
	if err != nil {
		return 1, err
	}
	defer func() {
		if err := os.Remove(passWdTemp.Name()); err != nil {
			log.Error().Err(err).Msg("Error removing temporary passwd file")
		}
	}()

	passwdCmd := []string{"run", "--rm", "--tty", "--interactive"}
	passwdCmd = append(passwdCmd, userArgs...)
	passwdCmd = append(passwdCmd, "--mount", cacheMount)
	passwdCmd = append(passwdCmd, f.ImageUrl)

	// Add copy command
	passwdCmd = append(passwdCmd, "cp", "/etc/passwd", passWdTemp.Name())

	// Obtain passwd from the image
	if err := utils.Run(utils.RunArgs{
		Command: "docker",
		Args:    passwdCmd,
		Env:     []string{},
		PassEnv: passEnv,
	}); err != nil {
		return 1, err
	}

	// Add user to the container's passwd file
	file, err := os.OpenFile(passWdTemp.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 1, err
	}

	if _, err = fmt.Fprintf(file, "%s:x:%d:%d::%s:/bin/bash\n", os.Getenv("USER"), os.Getuid(), os.Getgid(), os.Getenv("HOME")); err != nil {
		return 1, err
	}

	if err = file.Close(); err != nil {
		return 1, err
	}

	// Create passwd mount
	passwdMount := fmt.Sprintf("type=bind,source=%s,target=/etc/passwd", passWdTemp.Name())

	// Check if PassEnv is set in the environment
	if env.PassEnv == nil {
		passEnv = f.PassEnv
	} else {
		passEnv = *env.PassEnv
	}

	// The environment's MountCWD setting overrides the runtime's MountCWD
	// setting when it is set
	mountCWD = f.MountCWD
	if env.MountCWD != nil {
		mountCWD = *env.MountCWD
	}

	imageUrl, err := f.Pull(passEnv)

	if err != nil {
		return 1, err
	}

	mounts := []string{}

	workingDir, err = os.Getwd()
	if err != nil {
		return 1, err
	}

	// Create mount binding flags
	for _, mount := range env.Mounts {
		mountSplit := strings.Split(mount, ":")
		if len(mountSplit) == 1 {
			mountSplit = append(mountSplit, mountSplit[0])
		}
		mountString := fmt.Sprintf("type=bind,source=%s,target=%s", mountSplit[0], mountSplit[1])
		if mountSplit[0] == workingDir {
			workingDirMounted = true
		}

		mounts = append(mounts, "--mount", mountString)
	}

	// Create read-only mount binding flags
	for _, roMount := range env.ReadOnlyMounts {
		roSplit := strings.Split(roMount, ":")
		if len(roSplit) == 1 {
			roSplit = append(roSplit, roSplit[0])
		}
		roString := fmt.Sprintf("type=bind,source=%s,target=%s,readonly", roSplit[0], roSplit[1])
		mounts = append(mounts, "--mount", roString)
	}

	// Set base dockerArgs:
	// - Use `run` to run the container
	dockerArgs = []string{"run", "--read-only", "--rm", "--tty", "--interactive"}

	// Generate random name
	containerName := fmt.Sprintf("moat-%s", utils.RandomString(4))
	dockerArgs = append(dockerArgs, "--name", containerName, "--hostname", containerName)

	// Run as the current user
	dockerArgs = append(dockerArgs, userArgs...)

	homeMount := fmt.Sprintf("type=bind,source=%s,target=%s", env.Home, os.Getenv("HOME"))

	// Add home mounts to docker arguments
	dockerArgs = append(
		dockerArgs, []string{"--mount", homeMount}...,
	)

	// Add other mounts to the docker arguments
	dockerArgs = append(
		dockerArgs, mounts...,
	)

	// Add passwd mount to ensure the container has the correct user information
	dockerArgs = append(dockerArgs, []string{"--mount", passwdMount}...)

	// Mount working directory if requested and not already mounted
	if !workingDirMounted && mountCWD {
		log.Debug().Str("workingDir", workingDir).Msg("Mounting working directory")
		dockerArgs = append(dockerArgs, []string{"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", workingDir, workingDir)}...)
		// Set working directory to current directory
		log.Debug().Str("workingDir", workingDir).Msg("Setting working directory for current directory")
		dockerArgs = append(dockerArgs, []string{"--workdir", workingDir}...)
	}

	// Add the image path to arguments
	dockerArgs = append(dockerArgs, imageUrl)

	log.Debug().Strs("args", args).Msg("User arguments")
	// Add the user command to arguments
	dockerArgs = append(dockerArgs, args...)

	cmdEnv = envVars

	log.Debug().Bool("passEnv", passEnv).Msg("PassEnv")

	if err := utils.Run(utils.RunArgs{Command: "docker", Args: dockerArgs, Env: cmdEnv, PassEnv: passEnv}); err != nil {
		return 1, err
	}
	return 0, nil
}

// Shell opens an interactive shell inside the container environment env
// and returns the exit code of the shell.
func (f *DockerRuntime) Shell(env types.MoatEnv) (int, error) {
	log.Debug().Msg("Shell called")
	return 0, nil
}
