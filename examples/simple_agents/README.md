# Simple agents in moat

This example runs CLI AI coding agents inside a moat environment. The
environment has a fake home directory that holds the agents' configurations
and state, and it mounts the current working directory into the container, so
you can run an agent on whatever project you are standing in. The agents
(`codex`, `claude`, `opencode`, `cline`, `pi`, and `omp`) are preinstalled in
the default container image. The full configuration is in
[moat-config.yaml](moat-config.yaml).

## Motivation

When you use a coding agent without containers, it reads its configuration and
state from your home directory and works on the project in your current
working directory.

This example gives you the same experience inside moat: one fake home holds
the configuration and state of every agent, and because the directory you run
`moat` from is mounted into the container, `moat run -n agents <agent>` works
from any project folder without per-project configuration. The agents' state
stays in the fake home instead of your real home, and the rest of the host
stays outside the container.

Use this example when you want to run a CLI coding agent on the project in
your current directory, with the agent's configuration and state isolated in
a fake home.

## Trying out the configuration

### 1. Install the latest moat release

```shell
curl -LO https://github.com/AaltoRSE/moat/releases/latest/download/moat_$(uname -s)_$(uname -m).tar.gz
tar -xzf moat_$(uname -s)_$(uname -m).tar.gz
mkdir -p ~/.local/bin
install moat ~/.local/bin/moat
```

Make sure `~/.local/bin` is in your `PATH`.

### 2. Install the example configuration

```shell
mkdir -p ~/.config/moat
curl -fsSL -o ~/.config/moat/moat-config.yaml https://raw.githubusercontent.com/AaltoRSE/moat/main/examples/simple_agents/moat-config.yaml
```

### 3. Create the fake home

```shell
mkdir -p ~/moat/agents-home
```

If you also use the [VSCode / vscodium](../vscode/README.md) example, this is
the same fake home — the two examples share it on purpose.

### 4. Enter the container

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
project, look at an agent's configuration in the fake home (e.g. `ls
~/.codex`), or install a missing tool. Exit the container with `exit`.

> [!NOTE]
> The first `moat run` pulls the container image into the local image cache,
> so it takes a while. Later runs start from the cached image.

### 5. Run codex

```shell
moat run -n agents "codex --no-daemon"
```

`--no-daemon` is needed inside the container; see [Implementation
details](#implementation-details).

### 6. Run the other agents

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
```

### Fake home

The agents store their configuration, session history, and state under the
home directory (e.g. `~/.codex/` or `~/.claude/`). Pointing `home` at
`~/moat/agents-home` keeps the agents away from your real home directory and
collects everything they write in one known place. This is the same directory
the [VSCode / vscodium](../vscode/README.md) example uses, so the two
examples share one setup when you use them together.

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
do ad-hoc work: inspect the agents' configuration in the fake home, test a
command, or install a missing tool.

### Agents in the container image

The default container image ships the agents preinstalled: `codex`,
`claude`, `opencode`, `cline`, `pi`, and `omp` (plus VSCode and VSCodium for
the [VSCode / vscodium](../vscode/README.md) example). You do not need to
install anything in the fake home; each agent creates its own configuration
files there on first run. The default runtime passes the host environment
variables into the container, so API keys you have set on the host are
available to the agent as well.

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
