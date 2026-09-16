# moat config — managing configuration

moat stores all of its settings in a single YAML file. The `moat config` command group lets you view and change that configuration from the command line, without opening an editor.

Every change made with `set`, `append`, or `prepend` is **validated before it is written to disk**: if the key is unknown, the value has the wrong type, or the resulting configuration is invalid, the change is rejected and the config file is left untouched.

## The config file

moat looks for a file named `moat-config.yaml` in:

1. `$HOME/.config/moat/` (global), then
2. the current directory (local override).

You can point moat at an explicit file with the global `-c/--config` flag:

```shell
moat -c /path/to/moat-config.yaml config show
```

> [!NOTE]
> The config file must already exist — moat does not create it for you. If none is found, all commands fail with `config file not found`. See [First-time setup](#first-time-setup) for a minimal starting file.

### Structure

The configuration has three top-level sections:

- `defaults` — settings that apply to every environment unless overridden:
  - `defaults.runtime` — the runtime name used when an environment does not specify one,
  - `defaults.runtimes.{name}` — the runtime specification (image, cache dir, …) for each runtime.
- `envs` — one entry per environment (fake home, mounts, command, …). See [ENV.md](ENV.md) for how environments work.
- `runtimes` (optional) — per-runtime specifications that override `defaults.runtimes` for a runtime name.

A complete example:

```yaml
defaults:
    runtime: apptainer
    runtimes:
        apptainer:
            type: apptainer
            imageurl: ghcr.io/aaltorse/moat:latest
            cachedir: $HOME/.cache/moat/images
            passenv: true
            mountcwd: false
envs:
    vscode:
        home: /home/user/.local/share/code-app/home
        mounts:
            - /home/user/projects
            - /run/dbus
        readonlymounts:
            - /usr/local
        mountcwd: true
        passenv: true
        command: code --wait .
    opencode:
        home: /home/user/.local/share/opencode/home
        runtime: apptainer
        command: opencode
```

### First-time setup

Create the global config file, e.g. `$HOME/.config/moat/moat-config.yaml`, with at least a `defaults` section and an empty `envs` map:

```yaml
defaults:
    runtime: apptainer
    runtimes:
        apptainer:
            type: apptainer
            imageurl: ghcr.io/aaltorse/moat:latest
            cachedir: $HOME/.cache/moat/images
            passenv: true
envs: {}
```

Then create your environments with [`moat env create`](ENV.md#moat-env-create).

## Subcommands

| Command | Purpose |
|---|---|
| `moat config set` | Set a configuration variable to a new value |
| `moat config append` | Append a value to a list configuration variable |
| `moat config prepend` | Prepend a value to a list configuration variable |
| `moat config show` | Print the whole configuration as YAML (aliases: `list`, `view`) |
| `moat config edit` | Open the config file in `$EDITOR` |

The global flags `-c/--config` and `-d/--debug` are available on every command.

## Keys

Variables are addressed with a **dot-separated path** through the config tree: `section.subsection…field`. Where a section contains a named map, the name is part of the path (e.g. `envs.vscode.home`).

The **type of the key** decides how many values a command accepts:

| Key type | Accepts | Examples of keys |
|---|---|---|
| string | exactly one value | `envs.vscode.command`, `defaults.runtime` |
| bool | exactly one value (`true`/`false`) | `envs.vscode.mountcwd`, `defaults.runtimes.apptainer.passenv` |
| list | one or more values (for `set`) / exactly one (for `append`/`prepend`) | `envs.vscode.mounts`, `envs.vscode.readonlymounts` |

Unknown keys are rejected with an error such as `key "envs.vscode.bogus" not found in configuration`.

### Settable keys

| Key | Type | Description |
|---|---|---|
| `defaults.runtime` | string | Default runtime used when an environment has no `runtime` of its own |
| `defaults.runtimes.{name}.type` | string | Runtime type, e.g. `apptainer` |
| `defaults.runtimes.{name}.imageurl` | string | Container image pulled for this runtime |
| `defaults.runtimes.{name}.cachedir` | string | Directory where container images are cached |
| `defaults.runtimes.{name}.passenv` | bool | Pass the host environment variables into the container |
| `defaults.runtimes.{name}.mountcwd` | bool | Bind-mount the directory you run `moat run` from, and use it as the working directory, for every environment that uses this runtime (default: `false`) |
| `runtimes.{name}.*` | as above | Runtime specs here override the matching `defaults.runtimes.{name}` entry |
| `envs.{name}.home` | string | Fake home directory (`$HOME` inside the container) |
| `envs.{name}.mounts` | list | Read-write mounts, `source` or `source:dest` |
| `envs.{name}.readonlymounts` | list | Read-only mounts, `source` or `source:dest` |
| `envs.{name}.mountcwd` | bool | Bind-mount the directory you run `moat run` from, and use it as the working directory (overrides `defaults.runtimes.{name}.mountcwd`) |
| `envs.{name}.runtime` | string | Runtime name for this environment (overrides `defaults.runtime`) |
| `envs.{name}.passenv` | bool | Pass host environment variables (overrides the runtime default) |
| `envs.{name}.command` | string | Default command run when `moat run -n {name}` gets no arguments |

Mounts use the same syntax as the `moat env` flags: `source` mounts at the same path, `source:dest` mounts at a different destination. Paths may contain environment variables (e.g. `$HOME`), which are expanded when used.

---

## `moat config set`

Set a configuration variable to a new value.

```
moat config set <key> <value> [<value> ...]
```

Behavior:

- The key must exist in the configuration structure, and its type decides how many values you may give: string and bool keys take exactly one value, list keys take one or more.
- For **list keys, all given values become the new list** — the previous list is replaced, not extended. Use `append`/`prepend` to keep the existing entries.
- Boolean values are parsed with Go's boolean rules, so `true`, `false`, `1`, `0`, `t`, `f` all work.
- The resulting configuration is re-validated and only then written to the config file. On failure the file is not modified.
- On success, prints e.g. `Set envs.vscode.mountcwd = [true]`.

Examples:

```shell
# Change the container image for the default apptainer runtime
moat config set defaults.runtimes.apptainer.imageurl ghcr.io/aaltorse/moat:v0.1.0

# Change where container images are cached
moat config set defaults.runtimes.apptainer.cachedir /home/user/.cache/moat/images

# Set the default command of an environment (quote multi-word commands)
moat config set envs.vscode.command "code --wait ."

# Mount the current working directory into the container
moat config set envs.vscode.mountcwd true

# Stop passing host environment variables through
moat config set envs.opencode.passenv false

# Point an environment at a different runtime
moat config set envs.opencode.runtime apptainer

# Replace the entire mount list at once
moat config set envs.vscode.mounts /home/user/projects /run/dbus
```

## `moat config append`

Append one value to a **list** configuration variable. The existing entries are kept; the new value is added at the end.

```
moat config append <key> <value>
```

Examples:

```shell
# Add a read-write mount to an existing environment
moat config append envs.vscode.mounts /home/user/projects/newproj

# Mount a directory at a different destination inside the container
moat config append envs.vscode.mounts /data/project:/workspace/project

# Add a read-only mount
moat config append envs.vscode.readonlymounts /usr/local:/usr/local
```

On success, prints e.g. `Appended /home/user/projects/newproj to envs.vscode.mounts`.

## `moat config prepend`

Prepend one value to a **list** configuration variable. The new value becomes the first entry; the existing entries keep their order.

```
moat config prepend <key> <value>
```

Examples:

```shell
# Make a mount the first entry of the list
moat config prepend envs.vscode.mounts /run/dbus

# Prepend a read-only mount
moat config prepend envs.vscode.readonlymounts /opt/tools:/opt/tools
```

On success, prints e.g. `Prepended /run/dbus to envs.vscode.mounts`.

## `moat config show`

Print the complete current configuration as YAML, including all defaults.

```
moat config show
```

Aliases: `list`, `view`.

Example:

```shell
$ moat config show
defaults:
    runtime: apptainer
    runtimes:
        apptainer:
            cachedir: $HOME/.cache/moat/images
            imageurl: ghcr.io/aaltorse/moat:latest
            passenv: true
            type: apptainer
envs:
    vscode:
        command: code --wait .
        home: /home/user/.local/share/code-app/home
        mounts:
            - /home/user/projects
        mountcwd: true
```

## `moat config edit`

Open the currently active config file in the editor given by the `$EDITOR` environment variable. This is the way to go for large or structural changes (new runtimes, reorganizing environments):

```shell
moat config edit
```

Behavior:

- Fails with an error if `$EDITOR` is not set (`export EDITOR=vim`, …).
- The file is **not** validated while editing — the configuration is validated the next time a moat command runs, so a syntax error or invalid value in the edited file will surface then.

---

## Notes

- `moat config` works at the raw key level. For creating, copying, or removing **environments**, prefer the [`moat env`](ENV.md) commands — they check environment names, verify that fake home and mount directories exist (and can create them with `-y`), and remove config entries cleanly. Use `config` on `envs.{name}.*` keys for fine-tuning an existing environment.
- A failed change never modifies the config file: the key, the value types, and the full resulting configuration are all checked first.
- `config set` on a list key **replaces** the list; to grow a list without losing entries, use `append` or `prepend`.
- `envs.{name}.mountcwd`, `envs.{name}.runtime`, and `envs.{name}.passenv` are overrides: when unset, the environment falls back to `defaults.runtime`, the runtime's `mountcwd` setting, and the runtime's own defaults.

## Typical workflow

```shell
# 1. See the current configuration
moat config show

# 2. Create an environment (see ENV.md)
moat env create -n myproj -H ~/moat-home -m ~/projects/myproj -y

# 3. Give it a default command
moat config set envs.myproj.command "code --wait ."

# 4. Mount the working directory automatically
moat config set envs.myproj.mountcwd true

# 5. Add another mount later
moat config append envs.myproj.mounts /run/dbus

# 6. Switch every environment to a newer image via the default runtime
moat config set defaults.runtimes.apptainer.imageurl ghcr.io/aaltorse/moat:latest

# 7. Make larger changes by hand
moat config edit
```
