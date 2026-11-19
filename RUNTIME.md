# moat runtime — managing runtimes

A **runtime** describes how moat starts a container: which container image is used, where the image is cached, and a couple of behavior flags. Every runtime has a **type** (currently `apptainer` and `docker`).

A runtime specification consists of:

- `type` — the runtime type, e.g. `apptainer` or `docker`,
- `executable` — the command that runs the container (optional; `apptainer` for the `apptainer` type and `docker` for the `docker` type by default),
- `imageurl` — the container image that is pulled (required for `apptainer`),
- `cachedir` — the directory where pulled images are cached (required for `apptainer`; used as a scratch directory by `docker`),
- `passenv` — whether the host environment variables are passed into the container,
- `mountcwd` — whether the directory you run `moat` from is bind-mounted into the container and used as the working directory, and
- `overlayowner`, `overlaydriver`, `overlayfssnapshot`, `overlaydir` — how the [overlay mounts](ENV.md#overlay-mounts--o--overlay-mount) of the environments that use this runtime are provided (apptainer only).

There are two sources of runtimes:

- **Default runtimes** live under `defaults.runtimes.{name}` in the config file. moat registers two itself (`apptainer` and `docker`), and you can edit them with [`moat config`](CONFIG.md) or the editor.
- **User runtimes** live under the top-level `runtimes.{name}` section. The `moat runtime create` command adds runtimes here. A user runtime with the same name as a default runtime **overwrites** it, so all lookups (including `moat runtime list` and environment resolution) see the user's version.

The `moat runtime` command group lets you create, inspect and modify runtimes. All changes are persisted to your config file (`$HOME/.config/moat/moat-config.yaml`, or `moat-config.yaml` in the current directory).

## Contents

- [Subcommands](#subcommands)
- [Common concepts](#common-concepts)
  - [Runtime names](#runtime-names)
  - [Runtime specifications](#runtime-specifications)
  - [Runtime types](#runtime-types)
  - [Overlay mounts](#overlay-mounts)
  - [Default and user runtimes](#default-and-user-runtimes)
  - [Using runtimes](#using-runtimes)
- [`moat runtime create`](#moat-runtime-create)
- [`moat runtime set`](#moat-runtime-set)
- [`moat runtime list`](#moat-runtime-list)
- [`moat runtime cache list`](#moat-runtime-cache-list)
- [Notes](#notes)
- [Typical workflow](#typical-workflow)

## Subcommands

| Command | Purpose |
|---|---|
| `moat runtime create` | Create a new runtime |
| `moat runtime list` | List runtime names |
| `moat runtime set` | Change one or more variables of an existing runtime |
| `moat runtime cache list` | List the images cached by a runtime |

The global flags `-c/--config` and `-d/--debug` are available on every command.

## Common concepts

### Runtime names

Names must be non-empty and contain only **alphanumeric characters, underscores, and hyphens** (`^[a-zA-Z0-9_-]+$`), e.g. `myrt`, `my-rt`, `apptainer`, `moat_v2`. The same rules as for [environment names](ENV.md#environment-names) apply.

### Runtime specifications

| Field | Description |
|---|---|
| `type` | Runtime type, `apptainer` or `docker`. Required. |
| `executable` | Command used to run the container, e.g. `apptainer`, `docker`, or a wrapper script. Optional: when it is not set, the `apptainer` type uses `apptainer` and the `docker` type uses `docker`. |
| `imageurl` | URL of the container image that is pulled (e.g. `ghcr.io/aaltorse/moat:latest`). Required for the `apptainer` type. The `docker` type also runs this image — moat does not enforce it, but `docker pull` needs a valid image reference at run time. |
| `cachedir` | Directory where pulled images are cached. Required for the `apptainer` type. The directory is created automatically on first use if it does not exist. The path may start with `~` (expanded to the home directory) or contain environment variables (e.g. `$HOME`), which are expanded, and relative paths are resolved to absolute ones before the runtime is stored. For the `docker` type the directory is not an image cache (Docker keeps its own); it is the scratch directory for the temporary user-identification file written for every run, and it is also created automatically on first use if it does not exist. |
| `passenv` | Pass the host environment variables into the container (default: `true`). |
| `mountcwd` | Bind-mount the directory you run `moat run` from, and use it as the working directory (default: `false`). |
| `overlayowner` | Who provides the overlay mounts of the environments using this runtime: `internal` (the container engine creates the overlays at run time) or `external` (moat pre-creates them on the host). Apptainer only. Default: `internal`. Internal storage supports `fuse-overlayfs`, `copy`, and `reflink`. |
| `overlaydriver` | Overlay driver: `fuse-overlayfs`, `kernel-overlayfs`, `btrfs`, `copy`, or `reflink`. Apptainer only. Default: `fuse-overlayfs`. |
| `overlayfssnapshot` | Snapshot method for an OverlayFS host lower. Values are `auto` (default), `nativefs`, `reflink`, `copy`, and `none`. |
| `overlaydir` | Managed root directory for externally owned overlays: per-environment overlay directories are created under `<overlaydir>/<env name>/`. Apptainer only. Default: `$HOME/.cache/moat/overlays`. The path may start with `~` or contain environment variables, and is resolved to an absolute path before the runtime is stored. |

### Runtime types

The runtime type decides which container technology `moat run` uses:

- **`apptainer`** (the default) — the image is pulled with `apptainer pull` into `cachedir` as a `.sif` file and reused from there while it exists. Apptainer needs no daemon, and no root privileges. A differently named or wrapped apptainer binary can be used with the `executable` field.
- **`docker`** — requires a working `docker` CLI (or another compatible executable, see `executable`) and daemon, and permission to run containers as your user (e.g. membership in the `docker` group). The image is pulled with `docker pull` before every run; Docker's own image cache makes repeated runs cheap. Every run creates a fresh container named `moat-xxxx` (`xxxx` is random) that is removed afterwards and whose root filesystem is read-only. The container runs as the invoking user: its `/etc/passwd` is taken from the image, augmented with your user's account line, and mounted back in so that tools inside the container can resolve your user. For that, `cachedir` is used as the scratch directory; it is created automatically when it does not exist.

Both types mount the environment's fake home directory at your home directory path and apply the environment's mounts and read-only mounts (see [ENV.md](ENV.md)).

### Overlay mounts

An [overlay entry](ENV.md#overlay-mounts--o--overlay-mount) exposes an image directory or the host working directory (`@cwd`) through private writable storage. Normal file permissions still apply.

- **`overlayowner: internal`** (default) uses temporary storage for a plain destination and removes it after the run. An explicit host source persists across runs. This owner supports `fuse-overlayfs` and the direct directory drivers.
- **`overlayowner: external`** keeps writes across runs. Managed storage lives under `<overlaydir>/<env name>/`. For `@cwd`, the storage key includes the resolved host directory and container destination, so different projects keep separate state. An explicit source uses the named host directory.
- **`overlaydriver: fuse-overlayfs`** (default) runs the libfuse3-based driver inside the container with Apptainer's `--fusemount`. Moat mounts a read-only view of the SIF image at a separate path for the lower layer and binds in the writable storage. This needs `/dev/fuse` on the host and `stat` inside the image. The published Moat images include both the driver and `stat`.
- **`overlaydriver: kernel-overlayfs`** mounts each overlay in a private mount namespace before running the command. It needs kernel support for unprivileged OverlayFS and user namespaces, plus `unshare` with `--map-current-user` and `--keep-caps`, `mount`, and `setpriv` in the container. The namespace maps the caller’s user and primary group IDs.
- **`overlaydriver: btrfs`** uses the same kernel mounts with a btrfs subvolume for each writable layer. Moat creates missing subvolumes and adds the `upper/` and `work/` directories. This driver additionally needs the `btrfs` tools on the host and a btrfs filesystem for storage.

- **`overlaydriver: copy`** binds an ordinary writable copy of the lower directory.
- **`overlaydriver: reflink`** binds a writable directory populated by file-level copy-on-write clones. The source and backing storage must support reflinks. An image filesystem such as SquashFS typically requires the `copy` driver.

For the three OverlayFS drivers, `overlayfssnapshot` controls how to capture an `@cwd` lower. The default `auto` tries a native btrfs subvolume snapshot, then reflinks, then a full copy. A native snapshot requires a btrfs subvolume as the source and compatible backing storage on the same filesystem. With nested subvolumes, `auto` falls back to a file-level method so their files are included. When creating a snapshot, an explicit `nativefs` or `reflink` request fails if the requested operation is unavailable. `none` uses the live project directory and emits a warning. Avoid host edits while a live OverlayFS mount is active. Image-backed OverlayFS entries already have an immutable lower and skip this setting. Copies and reflinks capture files individually at startup. Finish host edits before startup if you need a consistent point-in-time tree. Host copies and reflinks require GNU `cp` on the host. Image-backed copies use GNU `cp` inside the image.

Persistent mounts keep their original snapshot or writable tree across runs, including container edits and deletions. To start from the current project files, stop the environment and remove the associated backing directory. The `--debug` log includes the backing directory for each mount. Switching an external `@cwd` OverlayFS mount between live and snapshotted operation selects separate managed storage.

Overlay sources are exclusive during a run. Use separate sources for simultaneous runs that need writable copies of the same destination. Overlay settings are available for Apptainer runtimes.

### Default and user runtimes

- `moat runtime create` only ever adds to the **user** section (`runtimes.{name}`); it never modifies the defaults.
- `moat runtime set` modifies an **existing** runtime, user-specified or default. When the given runtime is a default runtime, a user runtime with the same name is created from the default runtime's settings (with the changes applied) and overwrites the default in all lookups; the defaults themselves are never modified.
- When a runtime is looked up by name, a user runtime takes precedence over a default runtime with the same name.
- Creating a runtime with the name of a default runtime (e.g. `apptainer`) is therefore a valid way to **override the default** — for example, to point every environment at a different image without touching `defaults`.

### Using runtimes

An environment uses the runtime named by `envs.{name}.runtime`, falling back to `defaults.runtime` (default: `apptainer`) when the environment does not specify one. You can change that with [`moat config set`](CONFIG.md#moat-config-set):

```shell
moat config set envs.myproj.runtime myrt
```

At run time, the image is pulled (for `apptainer` into `cachedir`, reusing a cached copy; for `docker` with `docker pull` before every run), and the runtime's `passenv` and `mountcwd` settings apply unless the environment overrides them with its own `passenv` and `mountcwd` fields (see [ENV.md](ENV.md) and [CONFIG.md](CONFIG.md)).

---

## `moat runtime create`

Create a new runtime and write it to the config file.

```
moat runtime create -n NAME --type TYPE [--executable EXEC] [--imageurl URL] [--cachedir DIR] [--passenv] [--mountcwd] [--overlayowner OWNER] [--overlaydriver DRIVER] [--overlayfssnapshot MODE] [--overlaydir DIR]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the runtime |
| `--type` | yes | Type of the runtime (`apptainer` or `docker`) |
| `--executable` | no | Executable that runs the container (default: `apptainer` for the `apptainer` type, `docker` for the `docker` type) |
| `--imageurl` | see note | URL of the container image (required for the `apptainer` type) |
| `--cachedir` | see note | Directory for caching container images (required for the `apptainer` type) |
| `--passenv` | no | Pass the environment variables to the container (default: `true`) |
| `--mountcwd` | no | Bind-mount the caller's current working directory into the container (default: `false`) |
| `--overlayowner` | no | Who provides the overlay mounts of the environments using this runtime: `internal` or `external` (apptainer only; default: `internal`; see [Overlay mounts](#overlay-mounts)) |
| `--overlaydriver` | no | Overlay filesystem driver: `fuse-overlayfs`, `kernel-overlayfs`, `btrfs`, `copy`, or `reflink` (apptainer only; default: `fuse-overlayfs`; `internal` supports `fuse-overlayfs`, `copy`, and `reflink`) |
| `--overlayfssnapshot` | no | Snapshot mode for an OverlayFS host lower (default: `auto`) |
| `--overlaydir` | no | Managed root directory for externally owned overlays (apptainer only; default: `$HOME/.cache/moat/overlays`) |

Behavior:

- The runtime specification is **validated before it is written to disk**: `type` must be given, and for the `apptainer` type so must `--imageurl` and `--cachedir`. For the `docker` type they are optional, but a runtime without `--imageurl` cannot pull an image at run time. The overlay flags are only allowed for the `apptainer` type, `internal` supports `fuse-overlayfs`, `copy`, and `reflink`, and the full resulting configuration is re-validated as well. On failure the config file is left untouched.
- If a user runtime with the same name already exists, nothing is done (`Runtime already exists.`).
- Creating a runtime with the name of a **default** runtime is allowed and overwrites that default in all lookups (see [Default and user runtimes](#default-and-user-runtimes)).
- The `--cachedir` and `--overlaydir` paths are expanded and resolved to absolute paths before they are stored.
- When `--executable` is not given, the runtime is stored without an `executable` value and the type's default executable (`apptainer` or `docker`) is used at run time.
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

# A runtime with persistent, externally owned overlays
moat runtime create -n myrt --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat/images \
    --overlayowner external --overlaydriver kernel-overlayfs

# A Docker-based runtime
moat runtime create -n docker-lab --type docker \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat

# Override the default apptainer runtime for all environments
moat runtime create -n apptainer --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat/images

# A runtime that uses a differently named container engine command
moat runtime create -n custom-exe --type apptainer \
    --imageurl ghcr.io/aaltorse/moat:latest \
    --cachedir $HOME/.cache/moat/images \
    --executable my-apptainer
```

## `moat runtime set`

Change one or more variables of an **existing** runtime. Only the flags you give are modified; every other variable of the runtime (and all other runtimes) keeps its current value. The change is written to the config file.

```
moat runtime set -n NAME ( [--type TYPE] | [--executable EXEC] | [--imageurl URL] | [--cachedir DIR] | [--passenv VAL] | [--mountcwd VAL] | [--overlayowner OWNER] | [--overlaydriver DRIVER] | [--overlayfssnapshot MODE] | [--overlaydir DIR] )...
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | yes | Name of the runtime to modify |
| `--type` | see note | New runtime type (`apptainer` or `docker`) |
| `--executable` | see note | New executable that runs the container |
| `--imageurl` | see note | New container image URL |
| `--cachedir` | see note | New directory for caching container images |
| `--passenv` | see note | Pass the host environment variables into the container (default: `true`) |
| `--mountcwd` | see note | Bind-mount the caller's current working directory into the container (default: `false`) |
| `--overlayowner` | see note | Who provides the overlay mounts of the environments using this runtime: `internal` or `external` (apptainer only) |
| `--overlaydriver` | see note | New overlay filesystem driver: `fuse-overlayfs`, `kernel-overlayfs`, `btrfs`, `copy`, or `reflink` (apptainer only) |
| `--overlayfssnapshot` | see note | Snapshot mode for an OverlayFS host lower |
| `--overlaydir` | see note | New managed root directory for externally owned overlays (apptainer only) |

At least one of `--type`, `--executable`, `--imageurl`, `--cachedir`, `--passenv`, `--mountcwd`, `--overlayowner`, `--overlaydriver`, `--overlayfssnapshot`, or `--overlaydir` must be given. For the boolean flags, `VAL` is `true` or `false` (the bare flag means `true`). The runtime must already exist, either as a user runtime or as a default runtime.

Behavior:

- When the given runtime is a **default** runtime (e.g. `apptainer`), a **user** runtime with the same name is created from the default runtime's settings with the given changes applied, so that it overwrites the default runtime in all lookups (see [Default and user runtimes](#default-and-user-runtimes)). Nothing under `defaults.runtimes` is modified.
- The updated specification is **validated before it is written to disk** (e.g. the `apptainer` type requires `imageurl` and `cachedir`; the overlay flags are only allowed for `apptainer`, and `internal` supports `fuse-overlayfs`, `copy`, and `reflink`), and the full resulting configuration is re-validated as well. On failure the config file is left untouched.
- The `--cachedir` and `--overlaydir` paths are expanded and resolved to absolute paths before they are stored.
- The configuration currently requires `passenv` to be `true`; setting `--passenv=false` is rejected by validation.
- On success, prints `Set NAME.field = [value]` for every changed variable and `Runtime updated successfully.`

Examples:

```shell
# Point a user runtime at a different image
moat runtime set -n myrt --imageurl ghcr.io/aaltorse/moat:v0.1.0

# Mount the working directory in the default apptainer runtime
moat runtime set -n apptainer --mountcwd

# Switch the default apptainer runtime to persistent, externally owned overlays
moat runtime set -n apptainer --overlayowner external --overlaydriver kernel-overlayfs

# Use btrfs subvolumes for the overlays, stored in a custom directory
moat runtime set -n apptainer --overlayowner external --overlaydriver btrfs --overlaydir $HOME/.cache/my-overlays

# Change several variables at once
moat runtime set -n myrt --imageurl ghcr.io/aaltorse/moat:latest --cachedir $HOME/.cache/moat/images --mountcwd

# Use a different container engine command
moat runtime set -n myrt --executable singularity
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
#   docker
#   myrt

# Check that a single runtime exists
moat runtime list -n myrt
```

## `moat runtime cache list`

List the container images that a runtime has cached locally, one per line. For the `apptainer` type these are the absolute paths of the `.sif` files in the runtime's [cache directory](#runtime-types); for the `docker` type they are the image references that `docker` reports for the repository of the runtime's image, i.e. all of its locally stored tags. Nothing is modified — the command only inspects the cache.

```
moat runtime cache list [-n NAME]
```

| Flag | Required | Description |
|---|---|---|
| `-n, --name` | no | List the images cached by this runtime (default: the default runtime, see [`defaults.runtime`](CONFIG.md#the-config-file)) |

Behavior:

- The images are listed one per line, sorted alphabetically.
- When the named runtime does not exist, an error is reported and nothing is printed.
- If the runtime has no cached images, prints `No cached images for runtime 'NAME'.`
- The command never creates the cache directory and never pulls images; it reports what is cached right now.

Examples:

```shell
# List the images cached by the default runtime
moat runtime cache list
#   /home/user/.cache/moat/images/ghcr.io_aaltorse_moat_latest.sif

# List the images cached by a specific runtime
moat runtime cache list -n moat-v0-1-0
#   /home/user/.cache/moat/images/ghcr.io_aaltorse_moat_v0.1.0.sif
```

---

## Notes

- The default runtime settings can also be overridden temporarily with environment variables (`MOAT_DEFAULTS_RUNTIMES_{NAME}_*`) — see [Environment variable overrides](CONFIG.md#environment-variable-overrides) in [CONFIG.md](CONFIG.md).
- The `docker` runtime talks to the Docker daemon; the `apptainer` runtime needs no daemon. See [Runtime types](#runtime-types) for the full differences.
- To fine-tune an individual field of an existing runtime, use [`moat runtime set`](#moat-runtime-set). To modify a field under `defaults.runtimes.{name}.*` directly (which `moat runtime set` never touches), use [`moat config set`](CONFIG.md#moat-config-set) on `defaults.runtimes.{name}.*` keys, or open the file with [`moat config edit`](CONFIG.md#moat-config-edit).
- There is no `runtime remove` yet: to get rid of a user runtime, delete its `runtimes.{name}` entry from the config file (or reset it with `moat config edit`).
- `mountcwd` set here is the runtime-level default; an environment's own `mountcwd` field takes precedence when set (see [ENV.md](ENV.md)).
- The overlay settings (`overlayowner`, `overlaydriver`, `overlayfssnapshot`, `overlaydir`) are runtime-level: environments declare **which** container paths to overlay via `envs.{name}.overlaymounts` (see [ENV.md](ENV.md#overlay-mounts--o--overlay-mount)); the runtime decides **how** the overlays are provided (see [Overlay mounts](#overlay-mounts)). With `overlayowner: external`, the persistent overlay directories live under `overlaydir` in per-environment subdirectories; delete them manually if you want to reset an environment's overlays.
- To free the disk space used by an `apptainer` runtime's cache, delete the `.sif` files listed by [`moat runtime cache list`](#moat-runtime-cache-list); the image is pulled again on the next run. The `docker` type needs no manual cache management — Docker prunes its own images.

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

# 5. Later, point the runtime at a newer image
moat runtime set -n moat-v0-1-0 --imageurl ghcr.io/aaltorse/moat:v0.1.1
```
