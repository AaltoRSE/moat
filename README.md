# moat

> [!IMPORTANT]
> moat is currently under development, features might not yet work

*Put a moat between your agent and the rest of the world*

moat is a small client that makes it easy to execute AI tools in containerized environments.

## Installation

Currently moat needs to be compiled manually. After installing go and downloading the repository, run:
```shell
go build -o moat
```

## Configuring moat

moat looks for a config file named `moat-config.yaml` in `$HOME/.config/moat/` or the current directory.

## Environments

moat creates separate environments for your coding tools and only mounts relevant directories into the environment.

### Creating an environment

An environment needs to have:
1. A fake home folder
2. List of directories / files you want to mount into the container.

You can create an environment with:

```shell
moat env create --name example-env --home ./home --mounts src,/path/to/some/other-directory
```

### Running your program in the environment

You can run a program in the environment with
```shell
moat run example-env my_program
```

## Runtimes

Currently moat uses Apptainer as a runtime. In the future other runtimes might be added.
