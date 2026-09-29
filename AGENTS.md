# Agent Instructions for moat

moat is a CLI tool that runs AI coding agents inside isolated [Apptainer](https://apptainer.org/) container environments. It manages named environments (each with a fake home directory and a set of filesystem mounts) and executes arbitrary commands inside them.

- **Module**: `github.com/AaltoRSE/moat`
- **Go version**: 1.26.4
- **Build**: `$(command -v go) build -o moat -ldflags "-X github.com/AaltoRSE/moat/internal/version.MoatVersion=$(git describe --tags)"`
- **Status**: Under active development — some command implementations are stubs.


---

## Best practices

- Use `go-doc`-skill when writing code comments.
- Keep the user-facing documentation (`README.md`, `ENV.md`, `CONFIG.md`, `RUNTIME.md`) in sync with code changes — see [User-facing documentation](#user-facing-documentation).

---

## Repository layout

```
moat/
├── main.go                     # Entry point; calls root.CreateRootCmd().Execute()
├── go.mod
├── AGENTS.md                   # Agent instructions (this file)
├── README.md                   # User-facing docs: overview, installation, quick start
├── ENV.md                      # User-facing docs for the `moat env` command group
├── CONFIG.md                   # User-facing docs for the `moat config` command group
├── RUNTIME.md                  # User-facing docs for the `moat runtime` command group
├── _config.yml                 # GitHub Pages config; docs are served from the repo root
│
├── cmd/                        # CLI layer — Cobra commands only, no business logic
│   ├── root/                   # Root command construction
│   │   └── root.go             # CreateRootCmd(); PersistentPreRunE inits logging + config; registers --config/--debug flags
│   ├── config/                 # `moat config` command group
│   │   ├── config.go           # CreateConfigCmd(); assembles subcommands
│   │   ├── set.go              # CreateConfigSetCmd(); calls internal/config.SetConfig()
│   │   ├── append.go           # CreateConfigAppendCmd(); calls internal/config.AppendConfig()
│   │   ├── prepend.go          # CreateConfigPrependCmd(); calls internal/config.PrependConfig()
│   │   ├── show.go             # CreateConfigShowCmd() (aliases: list, view); calls internal/config.GetConfigAsString()
│   │   ├── showdefaults.go     # CreateConfigShowDefaultsCmd(); calls internal/config.CreateDefaultConfig() + internal/config.GetConfigAsString()
│   │   ├── edit.go             # CreateConfigEditCmd(); opens config file in $EDITOR via internal/config.GetConfigFile and utils.Run
│   │   ├── config_test.go      # Shared config command test suite (ConfigTestSuite) + suite runner
│   │   ├── show_test.go        # Tests for the config show subcommand
│   │   ├── showdefaults_test.go # Tests for the config show-defaults subcommand
│   │   └── set_test.go         # Tests for the config set subcommand
│   ├── env/                    # `moat env` command group
│   │   ├── env.go              # CreateEnvCmd(); assembles subcommands; declares shared env flag groups (-n/--name, -H/--home, -m/--mount, -r/--ro-mount, -C/--command) as pflag.FlagSet factories
│   │   ├── create.go           # CreateEnvCreateCmd(); calls internal/env.CreateEnvironment()
│   │   ├── copy.go             # CreateEnvCopyCmd(); calls internal/env.CopyEnvironment()
│   │   ├── list.go             # CreateEnvListCmd(); reads config.CmdConfig directly (pre-existing exception)
│   │   ├── show.go             # CreateEnvShowCmd(); calls internal/config.GetEnv()
│   │   ├── set.go              # CreateEnvSetCmd(); calls internal/config.GetEnv() + SetConfig()
│   │   ├── remove.go           # CreateEnvRemoveCmd(); calls internal/env.RemoveEnvironment()
│   │   ├── env_test.go         # Shared env command test suite (EnvTestSuite) + suite runner
│   │   ├── list_test.go        # Tests for the env list subcommand
│   │   ├── show_test.go        # Tests for the env show subcommand
│   │   ├── create_test.go      # Tests for the env create subcommand
│   │   ├── copy_test.go        # Tests for the env copy subcommand
│   │   └── set_test.go         # Tests for the env set subcommand
│   ├── init/                   # `moat init` command
│   │   ├── init.go             # CreateInitCmd(); passes config.CmdConfig + --output/-o flag to internal/config.SetConfigPath(), then WriteConfig()
│   │   └── init_test.go        # Shared init command test suite (InitTestSuite) + suite runner
│   ├── run/                    # `moat run` command
│   │   └── run.go              # CreateRunCmd(); resolves env + runtime, sanitizes args, calls runtime.Run()
│   ├── runtime/                # `moat runtime` command group
│   │   ├── runtime.go          # CreateRuntimeCmd(); assembles subcommands
│   │   ├── create.go           # CreateRuntimeCreateCmd(); calls internal/runtime.CreateRuntime()
│   │   ├── list.go             # CreateRuntimeListCmd(); lists available runtimes via internal/config.GetRuntimes()
│   │   ├── set.go              # CreateRuntimeSetCmd(); calls internal/runtime.SetRuntime()
│   │   ├── runtime_test.go     # Shared runtime command test suite (RuntimeTestSuite) + suite runner
│   │   ├── create_test.go      # Tests for the runtime create subcommand
│   │   ├── set_test.go         # Tests for the runtime set subcommand
│   │   └── list_test.go        # Tests for the runtime list subcommand
│   └── version/                # `moat version` command
│       ├── version.go          # CreateVersionCmd(); prints internal/version.MoatVersion
│       └── version_test.go     # Shared version command test suite (VersionTestSuite) + suite runner
│
├── internal/                   # Business logic; never imported by cmd/ in reverse
│   ├── types/                  # Canonical location for ALL shared types (see rules below)
│   │   ├── config.go           # Config, Defaults
│   │   ├── runtimespec.go      # RuntimeSpec (runtime config fields), RuntimeUpdate (field/value pair used by runtime set)
│   │   └── moatenv.go          # MoatEnv
│   ├── config/
│   │   ├── config.go           # CmdConfig, InitConfig, WriteConfig, SetConfigPath, ValidateConfig, GetConfigAsString, CreateDefaultConfig, GetConfigFile, GetVariableType
│   │   ├── env.go              # GetEnv, GetEnvs
│   │   ├── runtime.go          # ValidateRuntimeSpec, GetRuntimeSpec, GetRuntimes, GetUserRuntimes
│   │   ├── set.go              # SetConfig
│   │   ├── append.go           # AppendConfig
│   │   ├── prepend.go          # PrependConfig
│   │   └── config_test.go      # Tests for InitConfig (default-config fallback)
│   ├── env/
│   │   ├── create.go           # CreateEnvironment
│   │   ├── copy.go             # CopyEnvironment
│   │   └── remove.go           # RemoveEnvironment
│   ├── logging/
│   │   └── logging.go          # InitLogging (zerolog setup)
│   ├── version/
│   │   └── version.go          # MoatVersion, set at build time via -ldflags
│   ├── runtime/                # Runtime management: internal functions for the `moat runtime` command group
│   │   ├── create.go           # CreateRuntime
│   │   └── set.go              # SetRuntime
│   ├── engines/                # Runtime engines that execute commands inside container environments
│   │   ├── runtime.go             # Runtime interface + GetRuntime factory
│   │   ├── apptainerruntime.go    # ApptainerRuntime implementation
│   │   └── apptainerruntime_test.go # ApptainerRuntime test suite (ApptainerRuntimeTestSuite, fake apptainer binary)
│   ├── utils/
│   │   ├── checks.go           # CheckFolderExists, CheckEnvironmentName, CheckMounts
│   │   ├── dirops.go           # CreateMountDirs
│   │   ├── sanitize.go         # SanitizeFolderPath, SanitizeMountsPaths, SanitizeArgs
│   │   └── run.go              # Run, RunCapture, RunArgs, OutputCapture
│   └── tests/
│       └── tests.go            # CreateTempConfig (shared test helper)
│
├── dockerfiles/
│   └── ubuntu24.04/            # Container image used by the apptainer runtime
│       ├── Dockerfile
│       └── entrypoint.sh
│
├── .github/
│   └── workflows/
│       └── deploy-image.yml    # Builds and publishes the docker image to GHCR on tag push
│
└── skills/
    └── go-doc/
        └── SKILL.md            # Go doc comment guidelines (see Best practices)
```

---

## Two-layer architecture

The codebase is split into two strict layers. **Never reverse the dependency direction.**

| Layer | Packages | Responsibility |
|---|---|---|
| **CLI** | `cmd/root/`, `cmd/config/`, `cmd/env/`, `cmd/init/`, `cmd/run/`, `cmd/runtime/`, `cmd/version/` | Cobra command construction, flag parsing, argument normalization, user-facing output. Delegates all logic to `internal/`. |
| **Logic** | `internal/...` | All business logic, I/O, config management, runtime execution. Must not import from `cmd/`. |

---

## Command construction pattern

Commands are built via **constructor functions**, not `init()` side effects. Each command group exposes a `Create{Group}Cmd()` function that returns a fully assembled `*cobra.Command` with all subcommands attached. Each subcommand has its own `Create{Group}{Sub}Cmd()` function in its own file.

- `cmd/root/root.go` — `CreateRootCmd()` builds the root command, registers persistent flags (`--config`, `--debug`), and attaches the six top-level subcommands via `CreateRunCmd()`, `CreateEnvCmd()`, `CreateConfigCmd()`, `CreateInitCmd()`, `CreateRuntimeCmd()`, `CreateVersionCmd()`.
- `cmd/config/config.go` — `CreateConfigCmd()` builds the `config` command and attaches `set`, `append`, `prepend`, `show`, `show-defaults`, `edit`.
- `cmd/env/env.go` — `CreateEnvCmd()` builds the `env` command and attaches `create`, `copy`, `list`, `show`, `set`, `remove`.
- `cmd/run/run.go` — `CreateRunCmd()` builds the `run` command (no subcommands).
- `cmd/runtime/runtime.go` — `CreateRuntimeCmd()` builds the `runtime` command and attaches `create`, `list`, and `set`.
- `cmd/init/init.go` — `CreateInitCmd()` builds the `init` command (no subcommands).
- `cmd/version/version.go` — `CreateVersionCmd()` builds the `version` command (no subcommands), which prints `internal/version.MoatVersion`.

`main.go` calls `root.CreateRootCmd().Execute()` directly. The root command's `PersistentPreRunE` hook initializes logging (`logging.InitLogging`) and configuration (`config.InitConfig`), storing the result in the package-level `config.CmdConfig` variable. All subcommands read the active configuration from `config.CmdConfig`.

---

## Package-by-package conventions

### `internal/types` — shared type definitions

**All types that are used by more than one package must be defined here.** This is the single source of truth for the domain model.

- Contains only `type` declarations (structs, interfaces, type aliases). No functions, no `init()`.
- Use `go-playground/validator` struct tags (`validate:"required"`, `validate:"dirpath"`, `validate:"filepath"`, `validate:"required_if=..."`) for any struct that passes through `internal/config.validateConfig`.
- When adding a new feature that introduces a shared struct or interface, add it here first, then use it from the relevant `internal/` package.

**Current types:**

| Type | File | Purpose |
|---|---|---|
| `Config` | `config.go` | Root config struct: `Defaults`, `Envs map[string]MoatEnv`, `Runtimes map[string]RuntimeSpec` |
| `Defaults` | `config.go` | Default runtime settings: `Runtime string`, `Runtimes map[string]RuntimeSpec` |
| `RuntimeSpec` | `runtimespec.go` | Config fields for any runtime (type, imageurl, cachedir, passenv, mountcwd) |
| `RuntimeUpdate` | `runtimespec.go` | A single field/value pair used by `moat runtime set` to change one runtime field |
| `MoatEnv` | `moatenv.go` | An individual named environment (home, mounts, readonlymounts, mountcwd, runtime, passenv, command) |

**Rule:** If you define a struct in `internal/engines`, `internal/env`, or any other package, and it is later referenced by a second package, move it to `internal/types`.

---

### `internal/config` — configuration management

- **Functions only.** Do not add type definitions here.
- `CmdConfig *viper.Viper` is a package-level variable holding the active configuration instance. It is set by `cmd/root`'s `PersistentPreRunE` after calling `InitConfig`. All subcommands and internal functions receive `*viper.Viper` as an explicit parameter; `CmdConfig` is used only by `cmd/` code that needs direct access (e.g. `cmd/env/list.go`).
- `InitConfig(cfgFile string) (*viper.Viper, error)` creates a new viper instance, registers defaults, loads the config file (named `moat-config.yaml`) from the given path or default search locations (`$HOME/.config/moat/`, `.`), unmarshals into `types.Config`, and validates. When no config file is present (neither the given path nor any search location), it does not return an error; the returned configuration simply holds the default configuration contents. When writing tests, this can be called multiple times with different config files.
- `ValidateConfig(cfg)` validates the configuration held in `cfg` using the same validation as `InitConfig` and the mutation functions. `ValidateRuntimeSpec(spec)` validates a single `types.RuntimeSpec` against its validator tags. Both return an error if validation fails and do not write to disk.
- Exported surface: `CmdConfig`, `InitConfig`, `WriteConfig`, `SetConfigPath`, `GetEnv`, `GetEnvs`, `GetRuntimeSpec`, `GetRuntimes`, `GetUserRuntimes`, `GetVariableType`, `GetConfigAsString`, `CreateDefaultConfig`, `GetConfigFile`, `SetConfig`, `AppendConfig`, `PrependConfig`, `ValidateConfig`, `ValidateRuntimeSpec`.
- One file per operation: `config.go` (init/lookup), `env.go` (environment lookup), `runtime.go` (runtime lookup and spec validation), `set.go`, `append.go`, `prepend.go`.
- `SetConfigPath(cfg, outputPath)` switches the configuration file path of `cfg` (via `viper.SetConfigFile`) to the global config file ($HOME/.config/moat/moat-config.yaml) — or to `outputPath` when it is non-empty — and validates the configuration. It does not write to disk. If a configuration file has already been found and no output path is given, the configuration is left as it is and `changed=false` is reported; it returns an error if `outputPath` is the same as the path of the found configuration. It reports whether the path was set (`changed`); callers that need to persist the configuration must call `WriteConfig` separately.
- `SetConfig`, `AppendConfig`, and `PrependConfig` all follow the same pattern: mutate the viper value and re-validate the full `types.Config`. They do not write to disk; callers that need to persist the change must call `WriteConfig` separately.
- Validation uses `go-playground/validator` and operates on `types.Config`.
- `GetEnv(cfg, name, sanitized)` returns `types.MoatEnv`; when `sanitized` is true, paths are resolved via `utils.SanitizeFolderPath` / `utils.SanitizeMountsPaths`. It must not return raw `map[string]interface{}` to callers.
- `GetEnvs(cfg)` returns `map[string]types.MoatEnv` for all configured environments.
- `GetRuntimeSpec(cfg, name)` looks up a runtime spec from `runtimes.{name}`, falling back to `defaults.runtimes.{name}` when the runtime is not user-specified, and returns `types.RuntimeSpec`.
- `GetRuntimes(cfg)` returns `map[string]types.RuntimeSpec` for all available runtimes: the default runtimes from `defaults.runtimes` merged with the user-specified runtimes from `runtimes`, where a user-specified runtime with the same name overwrites the default runtime.
- `GetUserRuntimes(cfg)` returns `map[string]types.RuntimeSpec` for the user-specified runtimes from the top-level `runtimes` key only, without merging in the default runtimes from `defaults.runtimes`.
- `GetConfigFile(cfg)` returns the path of the active config file, falling back to `$HOME/.config/moat/moat-config.yaml`.

---

### `internal/env` — environment lifecycle

- One file per operation: `create.go`, `copy.go`, `remove.go`. Add `list.go` if list logic is moved from `cmd/env/list.go`.
- Functions accept `*viper.Viper` and `types.MoatEnv` (or related parameters) as arguments; they do not parse flags or read `os.Args`.
- `CreateEnvironment(cfg, name, env, autoCreate)` creates a new environment. If `autoCreate` is true, missing directories (fake home and mount source paths) are created automatically. Otherwise, the user is prompted to create a missing fake home directory, and missing or invalid mount paths abort creation.
- `CopyEnvironment(cfg, sourceName, newName, home, mounts, roMounts, command, autoCreate)` copies an existing environment, optionally overriding home, mounts, read-only mounts, and command.
- `RemoveEnvironment(cfg, name)` removes an environment from the configuration.
- May use `promptkit` for interactive confirmation prompts (user-facing only; not in functions called programmatically).
- Must call `internal/config.WriteConfig(cfg)` after any mutation to persist changes.
- Path arguments must be resolved to absolute paths with `utils.SanitizeFolderPath` before storing.

---

### `internal/runtime` — runtime management

- Internal functions for the `moat runtime` command group (`cmd/runtime/`); they mutate the configuration and persist changes to the config file.
- One file per operation: `create.go`, `set.go`.
- `create.go` defines `CreateRuntime(cfg *viper.Viper, name string, spec types.RuntimeSpec) error`, which creates a new user-specified runtime in the top-level `runtimes` key of the configuration. It validates the name and the spec (`config.ValidateRuntimeSpec`), resolves the spec's `CacheDir` to an absolute path, skips creation when a user-specified runtime with the same name already exists, and writes the configuration to disk via `config.WriteConfig` only after `config.ValidateConfig` passes.
- `set.go` defines `SetRuntime(cfg *viper.Viper, name string, updates []types.RuntimeUpdate) error`, which changes the given fields of an existing runtime (user-specified or default) and writes the configuration to disk via `config.WriteConfig` only after `config.ValidateConfig` passes. Only the given fields are changed; all other fields keep their current values. When the named runtime is a default runtime, a user-specified runtime with the same name is created from the default runtime's specification with the given changes applied, so that it overwrites the default runtime in lookups; the default runtimes are never modified. The cache directory is resolved to an absolute path before storing.

---

### `internal/engines` — runtime engines

- Runtime engines implement the actual runtime functionality: executing commands and shells inside container environments. They are used by `moat run` through the `Runtime` interface.
- `runtime.go` defines the `Runtime` interface and the `GetRuntime(cfg *viper.Viper, name string) (Runtime, error)` factory. These must remain in this file.
- The `Runtime` interface requires two methods: `Run(env types.MoatEnv, args []string, envVars []string) (int, error)` and `Shell(env types.MoatEnv) (int, error)`.
- The runtime spec's `MountCWD` setting controls whether the runtime bind-mounts the caller's current working directory: when `ApptainerRuntime` is constructed from a spec with `MountCWD` enabled, `Run` adds `--bind <cwd> --pwd <cwd>` to the apptainer command line (unless the working directory is already given as an environment mount). The environment's `MountCWD` (`types.MoatEnv.MountCWD`) has priority over the runtime's `MountCWD`: when it is non-nil, its value overwrites the runtime's setting. The `mountcwd` default for the apptainer runtime is registered as `false` in `internal/config.CreateDefaultConfig`.
- Each runtime is implemented in its own file: `apptainerruntime.go`, and future runtimes in `{name}.go`.
- Implementation-specific helper structs that are **private to one runtime** (e.g. `ApptainerImage`) may be defined in the same file as the implementation.
- The `Runtime` interface itself must stay in `runtime.go`; if it is referenced from another package, do not duplicate it — import from `internal/engines`.
- New runtime: add a new `{name}.go` implementing `Runtime`, then register it in `GetRuntime` in `runtime.go`.

---

### `internal/utils` — utility functions

- **Pure helpers only.** No business logic, no user prompts, no viper access.
- Must not import other `internal/` packages (no circular risk, but avoids entanglement).
- `run.go`: `RunArgs` struct (Command, Args, Env, PassEnv), `Run`, `RunCapture`, `OutputCapture` for subprocess execution and stdout capture.
- `checks.go`: `CheckFolderExists`, `CheckEnvironmentName`, `CheckMounts`.
- `dirops.go`: `CreateMountDirs` — creates missing mount source directories.
- `sanitize.go`: `SanitizeFolderPath` (expands a leading `~` to the home directory, expands env vars, and resolves to an absolute path), `SanitizeMountsPaths` (sanitizes source paths in `source:dest` mount strings), `SanitizeArgs` (parses a single-arg command with shellwords, or validates multi-arg commands for spaces).
- New utilities must be genuinely reusable across multiple callers; one-off helpers belong in the package that uses them.

---

### `internal/tests` — shared test helpers

- Holds reusable helpers for building isolated test fixtures; it is imported by `cmd/` test files (e.g. `cmd/env/env_test.go`).
- `CreateTempConfig(envs, runtimes) (string, string, error)` creates a fresh temporary config file under `tests.MoatTestDir`, initializes it via `config.InitConfig`, creates the given environments via `env.CreateEnvironment` (with `autoCreate=true` so it never blocks on a prompt), sets the given runtimes under the top-level `runtimes` key (where they overwrite default runtimes with the same name), and returns the config file path, its contents, and any error.
- Helpers must be self-contained and must not depend on test-suite state; the caller is responsible for cleaning up any temporary files they create.
- If tests create temporary files or directories, these temporary files should be situated under `/tmp/moat_tests` specified in `tests.MoatTestDir`.

---

### `cmd/{group}/` — CLI commands

- Each command group (`config`, `env`, `run`) is its own directory and Go package.
- **One `{group}.go` file** defines the group's `Create{Group}Cmd()` constructor, which builds the parent `cobra.Command` and attaches all subcommands.
- **One file per subcommand** (e.g. `create.go`, `list.go`). Each file defines a `Create{Group}{Sub}Cmd()` constructor that builds and returns the subcommand with its flags and `Run` closure.
- Flags are registered on the command inside the constructor function. Required flags use `MarkFlagRequired`.
- `cmd/` files must not contain business logic. Extract any logic beyond argument marshaling into `internal/`.
- Use `log.Fatal().Msgf(...)` for unrecoverable errors; use `fmt.Println` for normal user-facing output.
- The `cmd/root/` package is special: it constructs the root command and owns the `PersistentPreRunE` hook that initializes logging and configuration.

---

## Naming conventions

| Category | Pattern | Examples |
|---|---|---|
| Command constructors | `Create{Group}{Sub}Cmd` | `CreateEnvCmd`, `CreateEnvCreateCmd`, `CreateConfigSetCmd`, `CreateRunCmd` |
| Cobra command variables (local) | `{verb}Cmd` | `createCmd`, `listCmd`, `runCmd`, `setCmd` |
| Exported functions | PascalCase, verb-first | `CreateEnvironment`, `GetRuntime`, `InitConfig` |
| Unexported functions | camelCase, verb-first | `validateConfig`, `getImage` |
| Type names | PascalCase, noun | `MoatEnv`, `Config`, `RuntimeSpec` |
| Interface names | PascalCase, noun or agent noun | `Runtime` |
| Flag variables | camelCase, descriptive | `mountString`, `absFakeHome` |
| Struct tags | lowercase validator keywords | `validate:"required"`, `validate:"required_if=Type=apptainer"` |

---

## Error handling

| Situation | Pattern |
|---|---|
| `internal/` function fails | Return `error`; caller checks `if err != nil` |
| Fatal error in `cmd/` (cannot continue) | `log.Fatal().Msgf(...)` — exits the process |
| User-facing informational message | `fmt.Println(...)` |
| Diagnostic / debug output | `log.Error().Msgf(...)` / `log.Info().Msgf(...)` / `log.Debug().Msgf(...)` (zerolog) |
| Flag registration failure in constructor | `panic(err)` — acceptable only during command construction |
| Viper / config parse error | Log full config via `GetConfigAsString()`, then return the wrapped error |

Do not use `log.Fatal` inside `internal/` packages — return errors and let `cmd/` decide how to handle them.

---

## Key dependencies

| Dependency | Used by | Purpose |
|---|---|---|
| `github.com/spf13/cobra` | All `cmd/` packages | CLI command tree, flag parsing |
| `github.com/spf13/viper` | `internal/config`, `cmd/env/list.go` | YAML config loading and global state |
| `github.com/go-playground/validator/v10` | `internal/config` only | Struct validation via tags |
| `github.com/erikgeiser/promptkit` | `internal/env` only | Interactive confirmation prompts |
| `github.com/rs/zerolog` | All packages | Structured logging |
| `go.yaml.in/yaml/v3` | `internal/config`, `cmd/env/show.go` | YAML marshal/unmarshal |
| `github.com/mattn/go-shellwords` | `internal/utils` | Shell-style word splitting for command arguments |
| `github.com/stretchr/testify` | All test codes | Test framework for creating tests |


Do not add viper access to packages other than `internal/config` without strong justification. Prefer calling `internal/config` functions instead.

---

## User-facing documentation

User-facing documentation lives at the **repository root** next to `README.md` — not in a `docs/` directory. The repository is configured as a GitHub Pages site rooted at the repo root (`_config.yml`), so the documents must stay at the root and link to each other with **relative links** (e.g. `[ENV.md](ENV.md)`, `[CONFIG.md](CONFIG.md#moat-config-set)`).

| Document | Covers |
|---|---|
| `README.md` | Project overview, installation, and quick start. The entry point for new users; keep it short and link to the per-group documents for details. |
| `ENV.md` | The `moat env` command group: what an environment is (fake home, mounts, read-only mounts, default command) and every `moat env` subcommand. |
| `CONFIG.md` | The `moat config` command group: config file location and structure, the settable keys and their types, and every `moat config` subcommand. |
| `RUNTIME.md` | The `moat runtime` command group: runtime specifications, default vs. user runtimes, and every `moat runtime` subcommand. |

### Structure

The command-group documents share a fixed skeleton; new documents and new sections must match it:

1. `# moat {group} — managing {noun}` title, followed by a short introduction that defines the core concept (in bold) and notes where changes are persisted.
2. `## Subcommands` — a `| Command | Purpose |` table listing every subcommand, followed by a note that the global `-c/--config` and `-d/--debug` flags are available on every command.
3. A concepts section (`## Common concepts`, `## The config file`, `## Keys`, …) covering rules shared by several subcommands (naming, path expansion, key types, non-interactive mode, …).
4. One `## moat {group} {sub}` section per subcommand, each containing, in order:
   - the usage synopsis in a code block (e.g. `moat env create -n NAME -H HOME [-m MOUNT]... [-y]`),
   - a `| Flag | Required | Description |` table (use `see note` where a flag is only conditionally required),
   - `Behavior:` bullets covering prompts, validation, error cases, and what the command prints,
   - `Examples:` in a `shell` code block.
5. `## Notes` — edge cases and cross-references to the sibling documents (where useful).
6. `## Typical workflow` — a numbered end-to-end example.

### Style

- Write for end users. Describe observable behavior; never mention implementation details (Go packages, internal functions, source file paths).
- Use backticks for commands, flags, keys, paths, and quoted output; quote multi-word commands in examples.
- Document every flag, its default value, and what happens on failure. Quoted output (e.g. `Environment already exists.`) must match what the command actually prints.
- Use GitHub alerts (`> [!NOTE]`, `> [!IMPORTANT]`) sparingly, only for things that would surprise the user.
- Each fact lives in exactly one document; cross-link with relative links instead of duplicating content.
- Examples must be realistic and copy-pasteable — they should work on a fresh install.

### Keeping the documentation up to date

Update the documentation **in the same change** as the code it describes; do not leave stale documents for a follow-up.

| Code change | Document(s) to update |
|---|---|
| New/removed subcommand | Its document's `## Subcommands` table and a new/removed `## moat {group} {sub}` section; `README.md` if the quick start is affected |
| New/changed/removed flag | The affected subcommand's usage synopsis, flag table, `Behavior:` bullets, and examples |
| Changed behavior (prompts, validation, defaults, error messages, printed output) | The affected `Behavior:` bullets and `## Notes`; fix any quoted output |
| New/changed config key or default | `CONFIG.md` (key table + structure examples) and any example config in the other documents |
| New runtime type or runtime spec field | `RUNTIME.md` (spec table + `create` flags) and `CONFIG.md` (`defaults.runtimes.{name}.*` / `runtimes.{name}.*` keys) |
| New environment field | `ENV.md` and the `envs.{name}.*` table in `CONFIG.md` |

Before committing, re-run the documented examples and compare the printed output against the document.

---

## Adding a new environment operation

1. Add any new shared types to `internal/types/`.
2. Implement the operation as a function in `internal/env/{operation}.go` accepting `*viper.Viper` and `types.MoatEnv`.
3. Add a Cobra command file `cmd/env/{operation}.go` with a `CreateEnv{Operation}Cmd()` constructor that parses flags, constructs the `types.MoatEnv`, and calls the internal function.
4. Register the subcommand in `CreateEnvCmd()` in `cmd/env/env.go` via `EnvCmd.AddCommand(CreateEnv{Operation}Cmd())`.
5. Update `ENV.md` (subcommand table + a `## moat env {operation}` section) and `README.md` if the quick start is affected — see User-facing documentation.

## Adding a new runtime

1. Add any runtime-specific config type fields to `internal/types/runtimespec.go` (alongside the existing `RuntimeSpec` fields, or as a new type if the runtime is structurally different).
2. Implement `internal/engines/{name}.go` with a struct that satisfies the `Runtime` interface (implementing both `Run` and `Shell`).
3. Register the new runtime type string in `GetRuntime` in `internal/engines/runtime.go`.
4. Add a constructor function (e.g. `New{Name}RuntimeFromSpec`) in the same file, mirroring `NewApptainerRuntimeFromSpec`.
5. Update `RUNTIME.md` (spec fields, `create` flags) and `CONFIG.md` if new spec fields or config keys are introduced — see User-facing documentation.

## Adding a new test

1. Create a test suite (subtype of `suite.Suite`) that encompasses more than one command at a time.
2. Create `SetupTest()`- and `TearDownTest()`-functions that set the base starting point for each test. Use `tests.CreateTempConfig` to build an isolated config fixture.
3. Create a test method `Test{Name}()` on the suite that runs a single test. Remember that each test runs sequentially, so each one starts from the same starting point specified by `SetupTest` and `TearDownTest`. Put tests for a specific subcommand in a `{subcommand}_test.go` file in the same directory (e.g. `cmd/config/set_test.go`); the shared suite type, `SetupTest`/`TearDownTest`, and the suite runner stay in `{group}_test.go`.
4. To execute a command, build the root command via `root.CreateRootCmd()`, set arguments with `rootCmd.SetArgs(...)`, and call `rootCmd.Execute()`. Capture stdout with `utils.OutputCapture`.
5. To verify configuration changes, call `config.InitConfig(configFile)` to get a fresh `*viper.Viper` and inspect it.

## Test creation best practices

- All tests must be written with testify: use `assert`/`require` for all checks and `suite.Suite` for grouping related tests. Plain `testing` idioms (`t.Error`, `t.Fatal`, `panic`, manual `if err != nil` error checks) must not be used in test code.
- Try to reuse the same test suite if possible. If there would be conflicts in `SetupTest` and `TearDownTest` among different tests, create a new suite.
- When testing command line commands with flags, create individual tests for each flag combination.
- One test can contain multiple assert-statements.
- When testing commands that change configuration values, use `config.InitConfig` to get a `*viper.Viper` object and test that object for changes.
- Use `suite.T().Cleanup(...)` to remove any temporary directories created during a test.

## Verifying additions

- Run go tests with `go test ./...`. All tests must pass.
- pre-commit hooks should be run after additions to verify that everything works. This can be done with `pre-commit run --all-files`.
- If the change affects user-facing behavior (subcommands, flags, config keys, defaults, or printed output), the relevant user-facing documentation (`README.md`, `ENV.md`, `CONFIG.md`, `RUNTIME.md`) must be updated in the same change — see User-facing documentation.
- Check whether new additions should be added to `AGENTS.md`.
