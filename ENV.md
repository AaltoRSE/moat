# moat env — managing environments

An **environment** is a named, self-contained container setup. It consists of:

- a **fake home** directory (used as `$HOME` inside the container),
- a list of **mounts** (read-write directories mounted into the container),
- a list of **read-only mounts**, and
- an optional **command** that is run by default when using the environment.

The `moat env` command group lets you create, inspect, copy, modify, and remove environments. All changes are persisted to your config file (`$HOME/.config/moat/moat-config.yaml`, or `moat-config.yaml` in the current directory).

## Subcommands

| Command | Purpose |
|---|---|
| `moat env create` | Create a new environment |
| `moat env copy` | Copy an existing environment (optionally overriding fields) |
| `moat env list` | List environment names |
| `moat env show` | Show the full configuration of one environment |
| `moat env set` | Change one or more variables of an existing environment |
| `moat env remove` | Remove an environment from the configuration |

The global flags `-c/--config` and `-d/--debug` are available on every command.

## Common concepts

### Environment names

Names must be non-empty and contain only **alphanumeric characters and underscores** (`^[a-zA-Z0-9_]+$`), e.g. `my_env`, `work`, `opencode`.

### Fake home (`-H/--home`)

The directory on the host that will be used as the fake `$HOME` inside the container. The path may start with `~` (expanded to the home directory) or contain environment variables (e.g. `$HOME`), which are expanded, and relative paths are resolved to absolute ones.

### Mounts (`-m/--mount`, `-r/--ro-mount`)

Mounts are given as `source` or `source:dest`:

- `--mount /home/user/project` — mount the source directory at the same path inside the container.
- `--mount /home/user/project:/workspace/project` — mount the source at a different destination path.

Both flags can be specified multiple times. Only the **source** path is expanded and resolved; the destination is used as-is. Mount sources must be existing directories (see `-y/--yes` below).

### Command (`-C/--command`)

An optional command that is started by default when running the environment. Use quotes for multi-word commands:

```shell
-C "code --wait ."
```

### Non-interactive mode (`-y/--yes`)

Without `-y`, `create` (and `copy`) will **prompt** you to create a missing fake home directory, and will **abort** if a mount source path is missing or invalid. With `-y`, missing directories (fake home and mount source paths) are created automatically without prompting. Use `-y` for scripting.

---

## `moat env create`

Create a new environment and write it to the config file.

```
moat env create -n NAME -H HOME [-m MOUNT]... [-r ROMOUNT]... [-C COMMAND] [-y]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment |
| `-H, --home` | yes | Fake home directory |
| `-m, --mount` | no | Mounted directory, `source` or `source:dest` (repeatable) |
| `-r, --ro-mount` | no | Read-only mounted directory, `source` or `source:dest` (repeatable) |
| `-C, --command` | no | Default command to run in the environment (quote multi-word commands) |
| `-y, --yes` | no | Create missing directories (fake home and mount paths) without prompting |

Behavior:

- If an environment with the same name already exists, nothing is done (`Environment already exists.`).
- If the fake home directory does not exist, you are prompted to create it (or it is created automatically with `-y`); answering no aborts creation.
- If a mount source path is missing or invalid, creation is aborted (or the directory is created automatically with `-y`).

Examples:

```shell
# Minimal environment with only a fake home
moat env create -n myenv -H ~/moat-home

# Full environment: project mount, read-only mount, default command
moat env create -n myenv -H ~/moat-home \
    -m ~/projects/myproject:/workspace \
    -r /run/dbus \
    -C "code --wait ."

# Non-interactive: create missing directories automatically
moat env create -n myenv -H ~/moat-home -m ~/projects/newproj -y
```

## `moat env copy`

Create a new environment as a copy of an existing one. The copy inherits all settings from the source environment; any flags you give override the corresponding field of the source.

```
moat env copy -s SOURCE -n NAME [-H HOME] [-m MOUNT]... [-r ROMOUNT]... [-C COMMAND] [-y]
```

| Flag | Required | Description |
|---|---|---|
| `-s, --source` | yes | Name of the source environment to copy from |
| `-n, --name` | yes | Name of the new environment |
| `-H, --home` | no | Override the fake home directory |
| `-m, --mount` | no | Override the mounts (repeatable) |
| `-r, --ro-mount` | no | Override the read-only mounts (repeatable) |
| `-C, --command` | no | Override the default command |
| `-y, --yes` | no | Create missing directories (fake home and mount paths) without prompting |

Behavior:

- The source environment must exist.
- Given mount flags **replace** the source's mount lists entirely (they are not appended).
- Directory creation/prompting behaves exactly as with `create` (useful when overriding `-H` or `-m` with new paths).
- If an environment with the target name already exists, nothing is done.

Examples:

```shell
# Straight copy
moat env copy -s myenv -n myenv-copy

# Copy but point at a different project
moat env copy -s myenv -n myenv2 -m ~/projects/otherproj

# Copy with a new fake home, created automatically if missing
moat env copy -s myenv -n myenv3 -H ~/moat-home-3 -y
```

## `moat env list`

List the names of all configured environments, one per line, sorted alphabetically.

```
moat env list [-n NAME]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | no | List only the environment with this name (errors if it does not exist) |

If no environments are configured, prints `No environments configured.`

Examples:

```shell
moat env list
moat env list -n myenv
```

## `moat env show`

Show the full configuration of a single environment as YAML (fake home, mounts, read-only mounts, and all other fields such as `mountcwd`, `runtime`, `passenv`, `command`).

```
moat env show -n NAME
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment to show |

Example:

```shell
moat env show -n vscode
```

## `moat env set`

Change one or more variables of an **existing** environment. Only the flags you give are modified; every other variable of the environment (and all other environments) keeps its current value. The change is written to the config file.

```
moat env set -n NAME ( [-H HOME] | [-m MOUNT]... | [-r ROMOUNT]... | [-C COMMAND] )...
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment to modify |
| `-H, --home` | see note | New fake home directory |
| `-m, --mount` | see note | New mounts, replacing the existing list (repeatable) |
| `-r, --ro-mount` | see note | New read-only mounts, replacing the existing list (repeatable) |
| `-C, --command` | see note | New default command |

At least one of `-H`, `-m`, `-r`, or `-C` must be given. The environment must already exist. For `-m`/`-r`, the given list replaces the previous list entirely.

Examples:

```shell
moat env set -n myenv -C "code --wait ."
moat env set -n myenv -H ~/moat-home-new
moat env set -n myenv -m ~/projects/myproject --ro-mount /run/dbus
```

## `moat env remove`

Remove an environment from the configuration.

```
moat env remove -n NAME
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment to remove |

Behavior:

- Only the configuration entry is removed. The fake home directory and any mounted directories on disk are **not** deleted.
- If the environment does not exist, an error is reported and nothing is done.

Example:

```shell
moat env remove -n myenv-copy
```

---

## Typical workflow

```shell
# 1. Create an environment for a project
moat env create -n myproj -H ~/moat-home \
    -m ~/projects/myproj -r /run/dbus -C "code --wait ." -y

# 2. See what was created
moat env show -n myproj

# 3. List all environments
moat env list

# 4. Make a variant for another project
moat env copy -s myproj -n otherproj -m ~/projects/otherproj

# 5. Tweak the default command
moat env set -n myproj -C "code --wait ."

# 6. Clean up
moat env remove -n otherproj
```
