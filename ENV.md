# moat env — managing environments

An **environment** is a named, self-contained container setup. It consists of:

- a **fake home** directory (used as `$HOME` inside the container),
- a list of **mounts** (read-write directories mounted into the container),
- a list of **read-only mounts**, and
- an optional **command** that is run by default when using the environment.

The `moat env` command group lets you create, inspect, copy, modify, and remove environments. All changes are persisted to your config file (`$HOME/.config/moat/moat-config.yaml`, or `moat-config.yaml` in the current directory).

## Contents

- [Subcommands](#subcommands)
- [Common concepts](#common-concepts)
  - [Environment names](#environment-names)
  - [Fake home (`-H/--home`)](#fake-home--h--home)
  - [Mounts (`-m/--mount`, `-r/--ro-mount`)](#mounts--m--mount--r--ro-mount)
  - [Working directory mounting (`mountcwd`)](#working-directory-mounting-mountcwd)
  - [Command (`-C/--command`)](#command--c--command)
  - [Non-interactive mode (`-y/--yes`)](#non-interactive-mode--y--yes)
- [`moat env create`](#moat-env-create)
- [`moat env copy`](#moat-env-copy)
- [`moat env list`](#moat-env-list)
- [`moat env show`](#moat-env-show)
- [`moat env set`](#moat-env-set)
- [`moat env remove`](#moat-env-remove)
- [Typical workflow](#typical-workflow)

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

### Working directory mounting (`mountcwd`)

The optional `mountcwd` field controls whether the directory you run `moat run` from is bind-mounted into the container (at the same path) and used as the working directory inside it. This is handy when you start `moat run` from a project directory that the environment does not list as a mount: the project is available inside the container without adding it to the config, and the container starts right in it.

- When `mountcwd` is unset, the `mountcwd` setting of the runtime the environment uses applies (the built-in runtimes default to `false` — see [RUNTIME.md](RUNTIME.md#runtime-specifications)).
- When set, it **overrides** the runtime's setting for this environment only.
- If the directory you run from is already mounted by the environment (one of the mounts or read-only mounts has it as its source), no extra mount is added.
- Set it with the `--mountcwd` flag of `moat env create`, `moat env copy`, or `moat env set`, giving `true`, `false`, or nothing (an empty value, as in `--mountcwd=` or `--mountcwd "$VAR"` with `VAR` unset, means `true`). Alternatively, use `moat config set envs.{name}.mountcwd ...` (see [CONFIG.md](CONFIG.md#moat-config-set)). Inspect the current setting with `moat env show`.

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
moat env create -n NAME -H HOME [-m MOUNT]... [-r ROMOUNT]... [-C COMMAND] [--mountcwd[=true|false]] [-y]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment |
| `-H, --home` | yes | Fake home directory |
| `-m, --mount` | no | Mounted directory, `source` or `source:dest` (repeatable) |
| `-r, --ro-mount` | no | Read-only mounted directory, `source` or `source:dest` (repeatable) |
| `-C, --command` | no | Default command to run in the environment (quote multi-word commands) |
| `--mountcwd` | no | Bind-mount the directory you run `moat run` from into the container and use it as the working directory: `true` or `false`, an empty value means `true` (see [Working directory mounting](#working-directory-mounting-mountcwd)) |
| `-y, --yes` | no | Create missing directories (fake home and mount paths) without prompting |

Behavior:

- If an environment with the same name already exists, nothing is done (`Environment already exists.`).
- If the fake home directory does not exist, you are prompted to create it (or it is created automatically with `-y`); answering no aborts creation.
- If a mount source path is missing or invalid, creation is aborted (or the directory is created automatically with `-y`).
- With `--mountcwd` or `--mountcwd=true`, the environment is created with working-directory mounting enabled; with `--mountcwd=false`, it is explicitly disabled. Without the flag, the environment has no `mountcwd` of its own and falls back to the runtime's setting.

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

# Mount the directory you run from into the container
moat env create -n myenv -H ~/moat-home --mountcwd=true
```

## `moat env copy`

Create a new environment as a copy of an existing one. The copy inherits all settings from the source environment; any flags you give override the corresponding field of the source.

```
moat env copy -s SOURCE -n NAME [-H HOME] [-m MOUNT]... [-r ROMOUNT]... [-C COMMAND] [--mountcwd[=true|false]] [-y]
```

| Flag | Required | Description |
|---|---|---|
| `-s, --source` | yes | Name of the source environment to copy from |
| `-n, --name` | yes | Name of the new environment |
| `-H, --home` | no | Override the fake home directory |
| `-m, --mount` | no | Override the mounts (repeatable) |
| `-r, --ro-mount` | no | Override the read-only mounts (repeatable) |
| `-C, --command` | no | Override the default command |
| `--mountcwd` | no | Set working-directory mounting in the copy: `true` or `false`, an empty value means `true`, overriding the source's setting (see [Working directory mounting](#working-directory-mounting-mountcwd)) |
| `-y, --yes` | no | Create missing directories (fake home and mount paths) without prompting |

Behavior:

- The source environment must exist.
- Given mount flags **replace** the source's mount lists entirely (they are not appended).
- Without `--mountcwd`, the copy inherits the source's `mountcwd` setting; with it, the copy's setting is the given value (`true`, `false`, or empty for `true`) regardless of the source's setting.
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

# Copy with working-directory mounting enabled
moat env copy -s myenv -n myenv4 --mountcwd=true
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
moat env set -n NAME ( [-H HOME] | [-m MOUNT]... | [-r ROMOUNT]... | [-C COMMAND] | [--mountcwd[=true|false]] )...
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the environment to modify |
| `-H, --home` | see note | New fake home directory |
| `-m, --mount` | see note | New mounts, replacing the existing list (repeatable) |
| `-r, --ro-mount` | see note | New read-only mounts, replacing the existing list (repeatable) |
| `-C, --command` | see note | New default command |
| `--mountcwd` | see note | Set working-directory mounting: `true` or `false`, an empty value means `true` (see [Working directory mounting](#working-directory-mounting-mountcwd)) |

At least one of `-H`, `-m`, `-r`, `-C`, or `--mountcwd` must be given. The environment must already exist. For `-m`/`-r`, the given list replaces the previous list entirely.

Examples:

```shell
moat env set -n myenv -C "code --wait ."
moat env set -n myenv -H ~/moat-home-new
moat env set -n myenv -m ~/projects/myproject --ro-mount /run/dbus
moat env set -n myenv --mountcwd=true
moat env set -n myenv --mountcwd=false
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
