# shark-tank

> [!IMPORTANT]
> shark-tank is currently under development, features might not yet work

shark-tank is a small client that makes it easy to execute AI tools in separate environments.

## Installation

Currently shark-tank needs to be compiled manually. After installing go and downloading the repository, run:
```shell
go build -o shark-tank
```

## Configuring shark-tank

To start using shark-tank, run
```shell
shark-tank init
```

## Environments

shark-tank creates separate environments for your coding tools and only mounts relevant directories into the environment.

### Creating an environment

An environment needs to have:
1. A fake home folder
2. List of directories / files you want to mount into the container.
3. (Optionally) List of directories /files you want to mount into the container in a read-only mode.

You can create an environment with:

```shell
shark-tank env create --name example-env --home ./home --mounts src,/path/to/some/other-directory --ro-mounts /path/to/ro-file
```

### Running your program in the environment

You can run a program in the environment with
```shell
shark-tank run --name example-env my_program
```

You can also set the environment by creating a `.sharkrc` with the name of the environment.
Then you can run:
```shell
shark-tank run my_program
```

## Runtimes

Currently shark-tank uses Apptainer as a runtime. In the future other runtimes might be added.
