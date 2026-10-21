package engines

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// NewDockerRuntimeFromSpec returns a new DockerRuntime built from the
// given runtime spec. The executable is taken from the spec's Executable
// setting, falling back to "docker" when the spec does not set it.
func NewDockerRuntimeFromSpec(spec *types.RuntimeSpec) *DockerRuntime {
	executable := spec.Executable
	if executable == "" {
		executable = "docker"
	}
	return &DockerRuntime{
		Executable: executable,
		ImageUrl:   spec.ImageUrl,
		CacheDir:   spec.CacheDir,
		PassEnv:    spec.PassEnv,
		MountCWD:   spec.MountCWD,
	}
}

// DockerRuntime is a runtime engine that executes commands inside Docker
// containers. The container image is pulled from its registry with docker
// before every execution — Docker caches pulled images itself, so
// CacheDir is not an image cache but the scratch directory for the
// temporary passwd file that provides the container with the current
// user's account information (see [DockerRuntime.Run]). CacheDir must
// exist when the runtime is used.
type DockerRuntime struct {
	// Executable is the command used to invoke docker.
	Executable string
	// ImageUrl is the reference of the container image that is pulled and
	// executed.
	ImageUrl string
	// CacheDir is the scratch directory in which the temporary passwd file
	// of each execution is created.
	CacheDir string
	// PassEnv is the default of the pass-env setting: when true, the
	// caller's environment variables are passed through to the executed
	// command. An environment's own pass-env setting overrides it.
	PassEnv bool
	// MountCWD is the default of the mount-cwd setting: when true, the
	// caller's current working directory is bind-mounted into the container
	// and used as the working directory. An environment's own mount-cwd
	// setting overrides it.
	MountCWD bool
}

// Pull pulls the runtime image with docker and returns the image
// reference. Unlike the Apptainer runtime, the image is pulled on every
// call; Docker keeps its own image cache, so repeated pulls are cheap. It
// returns an error if the pull fails.
func (f *DockerRuntime) Pull(passEnv bool) (string, error) {
	if err := utils.Run(utils.RunArgs{
		Command: f.Executable,
		Args:    []string{"pull", f.ImageUrl},
		Env:     []string{},
		PassEnv: passEnv,
	}); err != nil {
		log.Error().Err(err).Str("imageUrl", f.ImageUrl).Msg("Error pulling image")
		return "", err
	}

	return f.ImageUrl, nil
}

func (f *DockerRuntime) getEnvArgs() []string {
	var (
		envVars []string
		envArgs []string
	)
	envVars = os.Environ()
	for _, envVar := range envVars {
		envArgs = append(envArgs, "--env", envVar)
	}

	return envArgs
}

func (f *DockerRuntime) getUserArgs() []string {
	return []string{"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())}
}

// Run executes args as a command inside the container environment env
// and returns the exit code of the execution. The envVars are set for the
// executed command, and the caller's environment is passed through to the
// container when pass-env is enabled. It returns an error if no command is
// given, the image cannot be pulled, or the command fails to execute.
//
// The command runs in a fresh, read-only container that is named
// moat-<random> and removed afterwards, running as the current user. The
// environment's home directory is bind-mounted at the caller's home
// directory path, followed by the environment's mounts and read-only
// mounts. The /etc/passwd file of the image is copied into a temporary
// file in CacheDir, augmented with the current user's account line, and
// mounted back into the container so that tools inside it can resolve the
// current user; the temporary file is removed when Run returns.
func (f *DockerRuntime) Run(env types.MoatEnv, args []string, envVars []string) (int, error) {
	log.Debug().Msg("Run called")
	var (
		dockerArgs        []string
		cmdEnv            []string
		workingDir        string
		passEnv           bool
		mountCWD          bool
		workingDirMounted bool
		envArgs           []string
	)

	// Check if PassEnv is set in the environment
	if env.PassEnv == nil {
		passEnv = f.PassEnv
	} else {
		passEnv = *env.PassEnv
	}

	if passEnv {
		envArgs = f.getEnvArgs()
	} else {
		envArgs = []string{}
	}

	if len(args) < 1 {
		return 1, fmt.Errorf("no command provided to run")
	}

	// Run as the current user
	userArgs := f.getUserArgs()

	cacheDir, err := utils.SanitizeFolderPath(f.CacheDir)

	if err != nil {
		return 1, err
	}

	// The cache directory is mounted into the container so that the
	// passwd copy can be written into it.
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
		Command: f.Executable,
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
	passwdMount := fmt.Sprintf("type=bind,source=%s,target=/etc/passwd,readonly", passWdTemp.Name())

	groupsTemp, err := os.CreateTemp(cacheDir, "groups.*")
	if err != nil {
		return 1, err
	}
	defer func() {
		if err := os.Remove(groupsTemp.Name()); err != nil {
			log.Error().Err(err).Msg("Error removing temporary groups file")
		}
	}()

	groupsCmd := []string{"run", "--rm", "--tty", "--interactive"}
	groupsCmd = append(groupsCmd, userArgs...)
	groupsCmd = append(groupsCmd, "--mount", cacheMount)
	groupsCmd = append(groupsCmd, f.ImageUrl)

	// Add copy command
	groupsCmd = append(groupsCmd, "cp", "/etc/group", groupsTemp.Name())

	// Obtain groups from the image
	if err := utils.Run(utils.RunArgs{
		Command: f.Executable,
		Args:    groupsCmd,
		Env:     []string{},
		PassEnv: passEnv,
	}); err != nil {
		return 1, err
	}

	// Add user to the container's groups file
	file, err = os.OpenFile(groupsTemp.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 1, err
	}

	if _, err = fmt.Fprintf(file, "%s:x:%d:%s\n", os.Getenv("USER"), os.Getgid(), os.Getenv("USER")); err != nil {
		return 1, err
	}

	if err = file.Close(); err != nil {
		return 1, err
	}

	// Create groups mount
	groupsMount := fmt.Sprintf("type=bind,source=%s,target=/etc/group,readonly", groupsTemp.Name())

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

	// The fake home directory is mounted at the caller's home directory
	// path.
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

	// Add groups mount to ensure the container has the correct group information
	dockerArgs = append(dockerArgs, []string{"--mount", groupsMount}...)

	// Mount working directory if requested and not already mounted
	if !workingDirMounted && mountCWD {
		log.Debug().Str("workingDir", workingDir).Msg("Mounting working directory")
		dockerArgs = append(dockerArgs, []string{"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", workingDir, workingDir)}...)
		// Set working directory to current directory
		log.Debug().Str("workingDir", workingDir).Msg("Setting working directory for current directory")
		dockerArgs = append(dockerArgs, []string{"--workdir", workingDir}...)
	}

	// Add env arguments to the docker arguments
	dockerArgs = append(dockerArgs, envArgs...)

	// Add the image path to arguments
	dockerArgs = append(dockerArgs, imageUrl)

	log.Debug().Strs("args", args).Msg("User arguments")
	// Add the user command to arguments
	dockerArgs = append(dockerArgs, args...)

	cmdEnv = envVars

	log.Debug().Bool("passEnv", passEnv).Msg("PassEnv")

	if err := utils.Run(utils.RunArgs{Command: f.Executable, Args: dockerArgs, Env: cmdEnv, PassEnv: passEnv}); err != nil {
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
