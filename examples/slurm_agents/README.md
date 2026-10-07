# Slurm agents in moat

This example runs CLI AI coding agents inside a moat environment on a
RHEL 9-based HPC cluster, so that you (or the agent) can submit and monitor
Slurm jobs from inside the container. It uses the RHEL 9 compatible
(Rocky Linux 9) container image, stores moat's image cache and the agents'
fake home in your project directory (`$WRKDIR`), and mounts the host's user
database and munge socket into the container so that the Slurm client works.
The agents (`codex`, `claude`, `opencode`, `cline`, `pi`, and `omp`) are
preinstalled in the image, like in the
[Simple agents](../simple_agents/README.md) example. The full configuration
is in [moat-config.yaml](moat-config.yaml).

## Motivation

On an HPC cluster, the workflow is: log in to a login node, move into your
project, run a coding agent, and have it (or you) submit jobs with `sbatch`
and `srun` and watch them with `squeue`. Two things make this awkward with
containers:

- The cluster home directory is usually small and quota-limited; the large
  writable space is the project directory. The multi-GB container image
  cache and the agents' state should live there, not in `$HOME`.
- A Slurm client inside a container does not know your account and cannot
  authenticate to the cluster: the image only knows its own users, and Slurm
  authenticates with munge credentials that are signed by the cluster's
  munge daemon.

This example solves both: moat's cache and the agents' fake home live under
`$WRKDIR` (your project directory), and the container is given the host's
`/etc/passwd`, `/etc/group`, and `/run/munge`, so the Slurm client inside
the container works the same as on the login node.

Use this example when you want to run a CLI coding agent on a RHEL 9-based
cluster and let it submit Slurm jobs from inside the container.

## Trying out the configuration

### 1. Install the latest moat release

```shell
curl -LO https://github.com/AaltoRSE/moat/releases/latest/download/moat_$(uname -s)_$(uname -m).tar.gz
tar -xzf moat_$(uname -s)_$(uname -m).tar.gz
mkdir -p ~/.local/bin
install moat ~/.local/bin/moat
```

Make sure `~/.local/bin` is in your `PATH`.

### 2. Point `$WRKDIR` at your project directory

This example expects an environment variable `$WRKDIR` pointing to a
writable project directory where moat stores the container image and the
agents' fake home. Many cluster login profiles set `$WRKDIR` per project;
check that it is set:

```shell
echo $WRKDIR
```

If your cluster does not provide it, point it at a writable project
directory and export it for the session:

```shell
export WRKDIR=/projects/my-project
```

### 3. Create the directories

```shell
mkdir -p $WRKDIR/issues/moat/images $WRKDIR/agents-home
```

The image directory is created automatically by moat on first use; the fake
home must exist before `moat run` can mount it.

### 4. Install the example configuration

```shell
mkdir -p ~/.config/moat
curl -fsSL -o ~/.config/moat/moat-config.yaml https://raw.githubusercontent.com/AaltoRSE/moat/main/examples/slurm_agents/moat-config.yaml
```

The configuration references `$WRKDIR`, which moat expands when it runs, so
the same file works on any cluster as long as `$WRKDIR` points at the right
place.

### 5. Enter the container and check Slurm

From a project folder, `moat run` without a command runs the environment's
default command, `bash`, so you get an interactive shell inside the
container:

```shell
cd $WRKDIR/my-project
moat run -n slurm_agents
```

The project directory is mounted at the same path and is your working
directory. Check that the Slurm client works from inside the container:

```shell
squeue
```

If you see queue information, Slurm works in the container. You can then
exit with `exit`.

> [!NOTE]
> The first `moat run` pulls the rockylinux9 image (several GB) into
> `$WRKDIR/issues/moat/images`, so it takes a while. Later runs start from
> the cached image.

### 6. Run codex

```shell
moat run -n slurm_agents "codex --no-daemon"
```

`--no-daemon` is needed inside the container; see [Implementation
details](#implementation-details).

### 7. Run the other agents

The general form is `moat run -n slurm_agents HARNESS_NAME`, where
`HARNESS_NAME` is the name of the agent's binary in the container image:

```shell
moat run -n slurm_agents pi
moat run -n slurm_agents opencode
moat run -n slurm_agents cline
moat run -n slurm_agents claude
moat run -n slurm_agents omp
```

## Implementation details

This section explains the choices made in
[moat-config.yaml](moat-config.yaml):

```yaml
defaults:
    runtime: apptainer
    runtimes:
        apptainer:
            cachedir: $WRKDIR/issues/moat/images
            imageurl: ghcr.io/aaltorse/moat:v0.2.3-rockylinux9
            mountcwd: false
            passenv: true
            type: apptainer
envs:
    slurm_agents:
        command: bash
        home: $WRKDIR/agents-home/
        mountcwd: true
        mounts:
            - /run/munge
            - /etc/passwd
            - /etc/group
```

### RHEL 9 image

The `imageurl` points at the RHEL 9 compatible image (built from Rocky
Linux 9). On a RHEL 9-based cluster, a container whose userland matches the
cluster avoids most compatibility surprises. Unlike the default image, the
rockylinux9 image is published only with version tags (there is no
`latest`), because the Slurm client it ships must match the cluster's Slurm
version; the `v0.2.3` image ships Slurm 26.05. If your cluster runs a
different Slurm version, pin the image tag that matches it. The agents
(`codex`, `claude`, `opencode`, `cline`, `pi`, and `omp`) and VSCode /
VSCodium are preinstalled, like in the default image.

### Cache and fake home under `$WRKDIR`

- `cachedir: $WRKDIR/issues/moat/images` — the Apptainer image cache lives
  in the project directory instead of the usual `~/.cache/moat/images`.
  Cluster home directories are usually small and quota-limited, while the
  project directory is the large writable space; a container image of
  several GB belongs there. Users who share the project also share the
  cached image.
- `home: $WRKDIR/agents-home/` — the same reasoning for the agents'
  configuration and state.
- moat expands `$WRKDIR` in these paths when it runs, so the same
  configuration file works on different clusters and projects.

### Mounts for Slurm

- `/run/munge` — the host's munge socket directory. Slurm authenticates
  with munge credentials. The image has the munge client installed, and
  through this mount it asks the host's munge daemon (which holds the
  cluster's key) for a credential, so Slurm commands from inside the
  container are accepted by the cluster.
- `/etc/passwd` and `/etc/group` — the host's user and group databases. The
  RHEL 9 base image knows only its own root and system users. Apptainer
  runs the container process with your host user ID, and the Slurm client
  resolves the submitting account from the local user database — without
  these mounts, the container would not know your Slurm account.
- The Slurm client also needs its configuration file (`slurm.conf`). If a
  Slurm command complains about a missing `slurm.conf`, point the
  `SLURM_CONF` environment variable at a copy of it (for example in your
  project directory); the default runtime passes the host environment
  variables into the container (`passenv: true`).

> [!NOTE]
> Because the container authenticates to the cluster through the host's
> munge socket, everything the agent does inside the container runs under
> your cluster account. Treat the environment as your own account with full
> cluster access.

### Working directory mount

`mountcwd: true` (on the environment; the runtime-level default stays
`false`) bind-mounts the directory you run `moat run` from into the
container and uses it as the working directory. That is what lets you work
on any project folder without changing the configuration, and the agent
sees the project at the same absolute path it has on the host.

> [!NOTE]
> `mountcwd` limits the containment a bit: the container can see and modify
> everything in the current project directory. That is the point of this
> example — the agent has to work on the project — but the rest of the host
> filesystem (apart from the fake home and the three Slurm mounts) stays
> outside the container.

### Default command

`command: bash` means that `moat run -n slurm_agents` without arguments
drops you into an interactive shell inside the container. That is the
quickest way to do ad-hoc work, including the Slurm check from step 5.

### Running codex

The codex CLI can run its sessions through a background daemon process.
Inside a moat container, the container lives only as long as the `moat run`
invocation, so a daemon started inside it would be orphaned as soon as you
exit. Always run codex with `--no-daemon` so the agent runs entirely in the
foreground of the container:

```shell
moat run -n slurm_agents "codex --no-daemon"
```

The other agents (`pi`, `opencode`, `cline`, `claude`, and `omp`) do not
need any special flags.
