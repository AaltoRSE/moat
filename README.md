# moat

> [!IMPORTANT]
> moat is currently under development. Some features might change in the upcoming versions.

*Dig a moat between your agent and the rest of the world. It might not be much, but at least it is something.*

moat is a small command line client that makes it easy to execute AI tools in containerized environments.

## Contents

- [Motivation](#motivation)
- [Installation](#installation)
- [Configuring moat](#configuring-moat)
- [Environments](#environments)
  - [Creating an environment](#creating-an-environment)
  - [Running your program in the environment](#running-your-program-in-the-environment)
- [Runtimes](#runtimes)
- [Examples](#examples)
- [AI usage in the project](#ai-usage-in-the-project)

## Motivation

> My idea behind moat was to create a tool that makes it easy for people to run agents in containers.
> I, personally, do not want to give agents access to all of the files on my computer.
> That meant that in order for me to be able to use harnesses in my day-to-day work,
> I needed to figure out how I could containerize them.
>
> Even though many harnesses nowadays come with their sandboxing solutions I do not want to trust
> the harnesses themselves to access only the files I want them to access. I wanted something
> that was independent of the harnesses or editors and would work before the harness was even
> launched.
>
> I also wanted the tool to be as usable as possible. That is why I decided on doing a
> go binary and pre-built images. If the users can install apptainer/singularity or podman/docker,
> that should be enough. Many of the existing tools require lots of knowledge on how containerization
> works and I wanted my tool to have sensible defaults so that it would be useful for people
> without that knowledge.
>
> To accomplish this I decided to dig a moat between my non-AI work and my AI work:
> whenever I use AI tools I use moat to run the agents in a container that will limit
> the blast radius of my or my agent's mistakes by mounting only certain folders into
> the container.
>
> I know that moat is not a perfect security tool: if the agents want to escape, they probably can.
> But to me the bigger risk is the accidental sending of information to the LLM endpoints
> or accidental modifications to my files by the agents. And with those cases solved I can
> be relaxed enough to use my agents without giving them constant oversight.

- Simo, moat's developer

## Installation

1. Install [Apptainer](https://apptainer.org/) (the default runtime) or [Docker](https://docs.docker.com/). If you're using Linux, you can use either. On Mac only Docker is supported.
2. Download the `moat`-binary and run it:
```sh
curl -sSL https://github.com/AaltoRSE/moat/releases/latest/download/moat_$(uname -s)_$(uname -m).tar.gz | tar -xz -C . moat
./moat
```

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

moat ships with two runtimes: **Apptainer** (the default) and **Docker**. The runtime of an environment is selected in the configuration.

For more on runtimes, see [RUNTIME.md](./RUNTIME.md)

## Examples

See [EXAMPLES.md](./EXAMPLES.md) for ready-to-use example configurations, such as running VSCode or VSCodium in moat.

## AI usage in the project

This project has been created using AI assistance, but all contributions by the AI have been reviewed by a human.
