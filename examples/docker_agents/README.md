# Docker agents in moat

This example runs CLI AI coding agents inside a moat environment that runs
its containers with **Docker**. The environment has a fake home directory
that holds the agents' configurations and state, and it mounts the current
working directory into the container, so you can run an agent on whatever
project you are standing in. The agents (`codex`, `claude`, `opencode`,
`cline`, `pi`, and `omp`) are preinstalled in the container image, which is
pulled with `docker pull`. The full configuration is in
[moat-config.yaml](moat-config.yaml).

## Motivation

When you use a coding agent without containers, it reads its configuration
and state from your home directory and works on the project in your current
working directory.

This example gives you the same experience inside moat: one fake home holds
the configuration and state of every agent, and because the directory you run
`moat` from is mounted into the container, `moat run -n agents <agent>` works
from any project folder without per-project configuration. The agents' state
stays in the fake home instead of your real home, and the rest of the host
stays outside the container.

Use this example when you have a working Docker installation and want to run
a CLI coding agent on the project in your current directory, with the agent's
configuration and state isolated in a fake home.

## Trying out the configuration

### 1. Check that Docker works

The Docker runtime needs a working `docker` CLI and daemon, and permission to
run containers as your user (e.g. membership in the `docker` group or a
rootless Docker installation):

```shell
docker run --rm hello-world
```

If this prints a hello message, you are set.

### 2. Install the latest moat release

```shell
curl -LO https://github.com/AaltoRSE/moat/releases/latest/download/moat_$(uname -s)_$(uname -m).tar.gz
tar -xzf moat_$(uname -s)_$(uname -m).tar.gz
mkdir -p ~/.local/bin
install moat ~/.local/bin/moat
```

Make sure `~/.local/bin` is in your `PATH`.

### 3. Install the example configuration

```shell
mkdir -p ~/.config/moat
curl -fsSL -o ~/.config/moat/moat-config.yaml https://raw.githubusercontent.com/AaltoRSE/moat/main/examples/docker_agents/moat-config.yaml
```

### 4. Create the fake home and the scratch directory

```shell
mkdir -p ~/moat/agents-home ~/.cache/moat
```

The second directory is the Docker runtime's scratch space for the temporary
user-identification files it writes on every run; it must exist before
`moat run`.

### 5. Enter the container

From a project folder, `moat run` without a command runs the environment's
default command, `bash`, so you get an interactive shell inside the
container:

```shell
mkdir -p ~/projects/my-project
cd ~/projects/my-project
moat run -n agents
```

The project directory is mounted at the same path and is your working
directory. From the shell you can run arbitrary commands: inspect the
project or look at an agent's configuration in the fake home (e.g. `ls
~/.codex`). The container's root filesystem is read-only, so anything you
want to keep must go into the fake home (e.g. `pip install --user ...`) or
the project directory. Exit the container with `exit`.

> [!NOTE]
> The first `moat run` downloads the container image with `docker pull`, so
> it takes a while. Later runs re-check the registry and start from Docker's
> local image cache, which is quick.

### 6. Run codex

```shell
moat run -n agents "codex --no-daemon"
```

`--no-daemon` is needed inside the container; see [Implementation
details](#implementation-details).

### 7. Run the other agents

The general form is `moat run -n agents HARNESS_NAME`, where `HARNESS_NAME`
is the name of the agent's binary in the container image:

```shell
moat run -n agents pi
moat run -n agents opencode
moat run -n agents cline
moat run -n agents claude
moat run -n agents omp
```

## Implementation details

This section explains the choices made in [moat-config.yaml](moat-config.yaml)
and how the environment is used:

```yaml
envs:
    agents:
        command: bash
        home: $HOME/moat/agents-home
        mountcwd: true
        runtime: docker
```

### The Docker runtime

The environment sets `runtime: docker`, which selects moat's built-in
default `docker` runtime: containers are started with the `docker` CLI,
which talks to the Docker daemon. In practice:

- **A working Docker installation is required.** The `docker` CLI and daemon
  must be installed, and you must have permission to run containers as your
  user (see step 1).
- **The image is pulled before every run.** `docker pull` runs on each `moat
  run`; Docker's own image cache makes repeated runs cheap. The image is
  `ghcr.io/aaltorse/moat` with the tag matching your moat release.
- **A fresh container per run.** Every run starts a new container named
  `moat-xxxx` (`xxxx` is random) that is removed when the command exits. The
  container's root filesystem is read-only; the writable places are the fake
  home and the mounted directories.
- **You are you inside the container.** The container runs with your user and
  group ID (and a matching account entry, so tools can resolve your user
  name). Files the agents create in the project directory are owned by you,
  just like on the host.
- **A scratch directory must exist.** The default Docker runtime uses
  `~/.cache/moat/` as the scratch directory for the temporary
  user-identification files of each run, and it must exist before `moat run`
  — hence the `mkdir` in step 4.

### Fake home

The agents store their configuration, session history, and state under the
home directory (e.g. `~/.codex/` or `~/.claude/`). Pointing `home` at
`~/moat/agents-home` keeps the agents away from your real home directory and
collects everything they write in one known place.

### Working directory mount

The configuration has no explicit mounts. Instead, `mountcwd: true`
bind-mounts the directory you run `moat run` from into the container and uses
it as the working directory. That is what lets you work on any project folder
without changing the configuration, and the agent sees the project at the
same absolute path it has on the host.

> [!NOTE]
> `mountcwd` limits the containment a bit: the container can see and modify
> everything in the current project directory. That is the point of this
> example — the agent has to work on the project — but the rest of the host
> filesystem (apart from the fake home) stays outside the container.

### Default command

`command: bash` means that `moat run -n agents` without arguments drops you
into an interactive shell inside the container. That is the quickest way to
do ad-hoc work: inspect the agents' configuration in the fake home or test a
command. Remember that the container's root filesystem is read-only, so a
tool you want to keep has to be installed into the fake home (e.g. `pip
install --user ...`) rather than system-wide.

### Agents in the container image

The default container image ships the agents preinstalled: `codex`, `claude`,
`opencode`, `cline`, `pi`, and `omp`. You do not need to install
anything in the fake home; each agent creates its own configuration files
there on first run. The default runtime passes the host environment variables
into the container, so API keys you have set on the host are available to the
agent as well (see [Passing environment
variables](#passing-environment-variables)).

### Passing environment variables

By default, the Docker runtime passes your host environment variables into
the container, so credentials such as API keys that you have set on the host
reach the agents. If you would rather start the container without them, set
the environment's `passenv` field to `false`:

```shell
moat config set envs.agents.passenv false
```

The next `moat run` then starts the container with only the environment that
the image itself defines — none of your host variables are visible inside.
Tools that need credentials must then get them another way, for example from
their own configuration in the fake home. To pass the host environment
variables again, set the field back to `true`:

```shell
moat config set envs.agents.passenv true
```

The environment-level `passenv` overrides the Docker runtime's default, and
it is the supported way to turn the passing off: the runtime's own `passenv`
setting must stay enabled.

### Running codex

The codex CLI can run its sessions through a background daemon process.
Inside a moat container, the container lives only as long as the `moat run`
invocation, so a daemon started inside it would be orphaned as soon as you
exit. Always run codex with `--no-daemon` so the agent runs entirely in the
foreground of the container:

```shell
moat run -n agents "codex --no-daemon"
```

The other agents (`pi`, `opencode`, `cline`, `claude`, and `omp`) do not need
any special flags.
