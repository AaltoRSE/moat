# moat

> [!IMPORTANT]
> moat is currently under development. Some features might change in the upcoming versions.

*Dig a moat between your agent and the rest of the world. It might not be much, but at least it is something.*

moat is a small command line client that makes it easy to execute AI tools in containerized environments.

## Installation

> [!IMPORTANT]
> Currently only Linux is supported, because Apptainer needs to be installed

1. Install [Apptainer](https://apptainer.org/).
2. Download the `moat`-binary and run it.

Alternatively, you can build the project yourself with `go`:
```shell
git clone https://github.com/AaltoRSE/moat.git
cd moat
go build -o moat
```

## Configuring moat

moat looks for a config file named `moat-config.yaml` in `$HOME/.config/moat/` or the current directory.

To initialize a default configuration, run:

```shell
moat init
```

For more on configuring, see [CONFIG.md](./CONFIG.md)

## Environments

moat creates separate environments for your coding tools and only mounts relevant directories into the environment.

### Creating an environment

An environment has a fake home folder and optional mounts that mount directories when the environment is being run.

You can create an environment with:

```shell
moat env create --name example-env --home ./home
```

With various flags you can specify the environment to:
1. Mount files or directories to the environment
2. Mount files or directories to the environment in read-only mode
3. Mount current directory to the environment

For more on environments, see [ENV.md](./ENV.md)

### Running your program in the environment

You can run a program in the environment with
```shell
moat run --name example-env my_program
```

## Runtimes

Currently moat uses Apptainer as a runtime. In the future other runtimes might be added.

For more on runtimes, see [RUNTIME.md](./RUNTIME.md)

## Examples

See [EXAMPLES.md](./EXAMPLES.md) for ready-to-use example configurations, such as running VSCode or VSCodium in moat.

## AI usage in the project

This project has been created using AI assistance, but all contributions by the AI have been reviewed by a human.
