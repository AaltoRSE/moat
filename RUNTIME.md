# moat runtime — managing runtimes

A **runtime** describes how moat starts a container: which container image is used, where the image is cached, and a couple of behavior flags. Every runtime has a **type** (currently only `apptainer`).

A runtime specification consists of:

- `type` — the runtime type, e.g. `apptainer`,
- `imageurl` — the container image that is pulled (required for `apptainer`),
- `cachedir` — the directory where pulled images are cached (required for `apptainer`),
- `passenv` — whether the host environment variables are passed into the container, and
- `mountcwd` — whether the directory you run `moat` from is bind-mounted into the container and used as the working directory.

There are two sources of runtimes:

- **Default runtimes** live under `defaults.runtimes.{name}` in the config file. moat registers one itself (`apptainer`), and you can edit them with [`moat config`](CONFIG.md) or the editor.
- **User runtimes** live under the top-level `runtimes.{name}` section. The `moat runtime create` command adds runtimes here. A user runtime with the same name as a default runtime **overwrites** it, so all lookups (including `moat runtime list` and environment resolution) see the user's version.

The `moat runtime` command group lets you create and inspect runtimes. All changes are persisted to your config file (`$HOME/.config/moat/moat-config.yaml`, or `moat-config.yaml` in the current directory).

## Subcommands

| Command | Purpose |
|---|---|
| `moat runtime create` | Create a new runtime |
| `moat runtime list` | List runtime names |

The global flags `-c/--config` and `-d/--debug` are available on every command.

## Common concepts

### Runtime names

Names must be non-empty and contain only **alphanumeric characters and underscores** (`^[a-zA-Z0-9_]+$`), e.g. `myrt`, `apptainer`, `moat_v2`. The same rules as for [environment names](ENV.md#environment-names) apply.

### Runtime specifications

| Field | Description |
|---|---|
| `type` | Runtime type, e.g. `apptainer`. Required. |
| `imageurl` | URL of the container image that is pulled (e.g. `ghcr.io/aaltorse/moat:latest`). Required for the `apptainer` type. |
| `cachedir` | Directory where pulled images are cached. Required for the `apptainer` type. The directory is created automatically on first use if it does not exist. The path may start with `~` (expanded to the home directory) or contain environment variables (e.g. `$HOME`), which are expanded, and relative paths are resolved to absolute ones before the runtime is stored. |
| `passenv` | Pass the host environment variables into the container (default: `true`). |
| `mountcwd` | Bind-mount the directory you run `moat run` from, and use it as the working directory (default: `false`). |

### Default and user runtimes

- `moat runtime create` only ever adds to the **user** section (`runtimes.{name}`); it never modifies the defaults.
- When a runtime is looked up by name, a user runtime takes precedence over a default runtime with the same name.
- Creating a runtime with the name of a default runtime (e.g. `apptainer`) is therefore a valid way to **override the default** — for example, to point every environment at a different image without touching `defaults`.

### Using runtimes

An environment uses the runtime named by `envs.{name}.runtime`, falling back to `defaults.runtime` (default: `apptainer`) when the environment does not specify one. You can change that with [`moat config set`](CONFIG.md#moat-config-set):

```shell
moat config set envs.myproj.runtime myrt
```

At run time, the runtime's `imageurl` is pulled into its `cachedir`, and its `passenv` and `mountcwd` settings apply unless the environment overrides them with its own `passenv` and `mountcwd` fields (see [ENV.md](ENV.md) and [CONFIG.md](CONFIG.md)).

---

## `moat runtime create`

Create a new runtime and write it to the config file.

```
moat runtime create -n NAME --type TYPE [--imageurl URL] [--cachedir DIR] [--passenv] [--mountcwd]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the runtime |
| `--type` | yes | Type of the runtime (e.g. `apptainer`) |
| `--imageurl` | see note | URL of the container image (required for the `apptainer` type) |
| `--cachedir` | see note | Directory for caching container images (required for the `apptainer` type) |
| `--passenv` | no | Pass the environment variables to the container (default: `true`) |
| `--mountcwd` | no | Bind-mount the caller's current working directory into the container (default: `false`) |

Behavior:

- The runtime specification is **validated before it is written to disk**: `type` must be given, and for the `apptainer` type so must `--imageurl` and `--cachedir`. The full resulting configuration is re-validated as well. On failure the config file is left untouched.
- If a user runtime with the same name already exists, nothing is done (`Runtime already exists.`).
- Creating a runtime with the name of a **default** runtime is allowed and overwrites that default in all lookups (see [Default and user runtimes](#default-and-user-runtimes)).
- The `--cachedir` path is expanded and resolved to an absolute path before it is stored.
- The configuration currently requires `passenv` to be `true`; creating a runtime with `--passenv=false` is rejected by validation.
- On success, prints `Creating runtime 'NAME'` and `Runtime created successfully.`

Examples:

```shell
# A runtime that pins a specific image version
moat runtime create -n moat-v0-1-0 --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:v0.1.0 \
    --cachedir $HOME/.cache/moat/images

# A runtime that mounts the working directory by default
moat runtime create -n myrt --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat/images \
    --mountcwd

# Override the default apptainer runtime for all environments
moat runtime create -n apptainer --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat/images
```

## `moat runtime list`

List the names of all available runtimes — the default runtimes plus all user runtimes — one per line, sorted alphabetically. When a user runtime has the same name as a default runtime, the name appears only once.

```
moat runtime list [-n NAME]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | no | List only the runtime with this name (an error is reported and nothing is printed if it does not exist) |

If no runtimes are configured at all, prints `No runtimes configured.`

Examples:

```shell
# List all available runtimes
moat runtime list
#   apptainer
#   myrt

# Check that a single runtime exists
moat runtime list -n myrt
```

---

## Notes

- `moat runtime create` works at the runtime level. To fine-tune an individual field of an existing runtime (or a default runtime), use [`moat config set`](CONFIG.md#moat-config-set) on `runtimes.{name}.*` / `defaults.runtimes.{name}.*` keys, or open the file with [`moat config edit`](CONFIG.md#moat-config-edit).
- There is no `runtime remove` yet: to get rid of a user runtime, delete its `runtimes.{name}` entry from the config file (or reset it with `moat config edit`).
- `mountcwd` set here is the runtime-level default; an environment's own `mountcwd` field takes precedence when set (see [ENV.md](ENV.md)).

## Typical workflow

```shell
# 1. See which runtimes are available
moat runtime list

# 2. Create a runtime that pins an image version
moat runtime create -n moat-v0-1-0 --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:v0.1.0 \
    --cachedir $HOME/.cache/moat/images

# 3. Point an environment at it (see ENV.md)
moat config set envs.myproj.runtime moat-v0-1-0

# 4. Verify it is in the list
moat runtime list -n moat-v0-1-0
```
