# Agent Instructions for moat

moat is a CLI tool that runs AI coding agents inside isolated [Apptainer](https://apptainer.org/) container environments. It manages named environments (each with a fake home directory and a set of filesystem mounts) and executes arbitrary commands inside them.

- **Module**: `github.com/AaltoRSE/moat`
- **Go version**: 1.26.4
- **Build**: `$(command -v go) build -o moat`
- **Status**: Under active development — some command implementations are stubs.


---

## Best practices

- Use `go-doc`-skill when writing code comments.

---

## Repository layout

```
moat/
├── main.go                     # Entry point; calls root.CreateRootCmd().Execute()
├── go.mod
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
│   │   ├── edit.go             # CreateConfigEditCmd(); opens config file in $EDITOR via internal/config.GetConfigFile and utils.Run
│   │   └── config_test.go      # Test suite for the config commands
│   ├── env/                    # `moat env` command group
│   │   ├── env.go              # CreateEnvCmd(); assembles subcommands
│   │   ├── create.go           # CreateEnvCreateCmd(); calls internal/env.CreateEnvironment()
│   │   ├── copy.go             # CreateEnvCopyCmd(); calls internal/env.CopyEnvironment()
│   │   ├── list.go             # CreateEnvListCmd(); reads config.CmdConfig directly (pre-existing exception)
│   │   ├── show.go             # CreateEnvShowCmd(); calls internal/config.GetEnv()
│   │   ├── remove.go           # CreateEnvRemoveCmd(); calls internal/env.RemoveEnvironment()
│   │   └── env_test.go         # Test suite for the env commands
│   └── run/                    # `moat run` command
│       └── run.go              # CreateRunCmd(); resolves env + runtime, sanitizes args, calls runtime.Run()
│
├── internal/                   # Business logic; never imported by cmd/ in reverse
│   ├── types/                  # Canonical location for ALL shared types (see rules below)
│   │   ├── config.go           # Config, Defaults
│   │   ├── runtimespec.go      # RuntimeSpec (runtime config fields)
│   │   └── moatenv.go          # MoatEnv
│   ├── config/
│   │   ├── config.go           # CmdConfig, InitConfig, WriteConfig, GetEnv, GetEnvs, GetRuntimeSpec, GetVariableType, GetConfigAsString, GetConfigFile
│   │   ├── set.go              # SetConfig
│   │   ├── append.go           # AppendConfig
│   │   └── prepend.go          # PrependConfig
│   ├── env/
│   │   ├── create.go           # CreateEnvironment
│   │   ├── copy.go             # CopyEnvironment
│   │   └── remove.go           # RemoveEnvironment
│   ├── logging/
│   │   └── logging.go          # InitLogging (zerolog setup)
│   ├── runtimes/
│   │   ├── runtime.go          # Runtime interface + GetRuntime factory
│   │   └── apptainerruntime.go # ApptainerRuntime implementation
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
├── skills/
│   └── go-doc/
│       └── SKILL.md            # Go doc comment guidelines (see Best practices)
│
└── docs/
    └── code-structure.md       # Placeholder
```

---

## Two-layer architecture

The codebase is split into two strict layers. **Never reverse the dependency direction.**

| Layer | Packages | Responsibility |
|---|---|---|
| **CLI** | `cmd/root/`, `cmd/config/`, `cmd/env/`, `cmd/run/` | Cobra command construction, flag parsing, argument normalization, user-facing output. Delegates all logic to `internal/`. |
| **Logic** | `internal/...` | All business logic, I/O, config management, runtime execution. Must not import from `cmd/`. |

---

## Command construction pattern

Commands are built via **constructor functions**, not `init()` side effects. Each command group exposes a `Create{Group}Cmd()` function that returns a fully assembled `*cobra.Command` with all subcommands attached. Each subcommand has its own `Create{Group}{Sub}Cmd()` function in its own file.

- `cmd/root/root.go` — `CreateRootCmd()` builds the root command, registers persistent flags (`--config`, `--debug`), and attaches the three top-level subcommands via `CreateRunCmd()`, `CreateEnvCmd()`, `CreateConfigCmd()`.
- `cmd/config/config.go` — `CreateConfigCmd()` builds the `config` command and attaches `set`, `append`, `prepend`, `show`, `edit`.
- `cmd/env/env.go` — `CreateEnvCmd()` builds the `env` command and attaches `create`, `copy`, `list`, `show`, `remove`.
- `cmd/run/run.go` — `CreateRunCmd()` builds the `run` command (no subcommands).

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
| `MoatEnv` | `moatenv.go` | An individual named environment (home, mounts, readonlymounts, mountcwd, runtime, passenv, command) |

**Rule:** If you define a struct in `internal/runtimes`, `internal/env`, or any other package, and it is later referenced by a second package, move it to `internal/types`.

---

### `internal/config` — configuration management

- **Functions only.** Do not add type definitions here.
- `CmdConfig *viper.Viper` is a package-level variable holding the active configuration instance. It is set by `cmd/root`'s `PersistentPreRunE` after calling `InitConfig`. All subcommands and internal functions receive `*viper.Viper` as an explicit parameter; `CmdConfig` is used only by `cmd/` code that needs direct access (e.g. `cmd/env/list.go`).
- `InitConfig(cfgFile string) (*viper.Viper, error)` creates a new viper instance, registers defaults, loads the config file (named `moat-config.yaml`) from the given path or default search locations (`$HOME/.config/moat/`, `.`), unmarshals into `types.Config`, and validates. When writing tests, this can be called multiple times with different config files.
- Exported surface: `CmdConfig`, `InitConfig`, `WriteConfig`, `GetEnv`, `GetEnvs`, `GetRuntimeSpec`, `GetVariableType`, `GetConfigAsString`, `GetConfigFile`, `SetConfig`, `AppendConfig`, `PrependConfig`.
- One file per operation: `config.go` (init/lookup), `set.go`, `append.go`, `prepend.go`.
- `SetConfig`, `AppendConfig`, and `PrependConfig` all follow the same pattern: mutate the viper value, re-validate the full `types.Config`, then persist via `WriteConfig`.
- Validation uses `go-playground/validator` and operates on `types.Config`.
- `GetEnv(cfg, name, sanitized)` returns `types.MoatEnv`; when `sanitized` is true, paths are resolved via `utils.SanitizeFolderPath` / `utils.SanitizeMountsPaths`. It must not return raw `map[string]interface{}` to callers.
- `GetEnvs(cfg)` returns `map[string]types.MoatEnv` for all configured environments.
- `GetRuntimeSpec(cfg, name)` looks up a runtime spec from `defaults.runtimes.{name}` and returns `types.RuntimeSpec`.
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

### `internal/runtimes` — runtime abstraction

- `runtime.go` defines the `Runtime` interface and the `GetRuntime(cfg *viper.Viper, name string) (Runtime, error)` factory. These must remain in this file.
- The `Runtime` interface requires two methods: `Run(env types.MoatEnv, args []string, envVars []string) (int, error)` and `Shell(env types.MoatEnv) (int, error)`.
- Each runtime is implemented in its own file: `apptainerruntime.go`, and future runtimes in `{name}.go`.
- Implementation-specific helper structs that are **private to one runtime** (e.g. `ApptainerImage`) may be defined in the same file as the implementation.
- The `Runtime` interface itself must stay in `runtime.go`; if it is referenced from another package, do not duplicate it — import from `internal/runtimes`.
- New runtime: add a new `{name}.go` implementing `Runtime`, then register it in `GetRuntime`.

---

### `internal/utils` — utility functions

- **Pure helpers only.** No business logic, no user prompts, no viper access.
- Must not import other `internal/` packages (no circular risk, but avoids entanglement).
- `run.go`: `RunArgs` struct (Command, Args, Env, PassEnv), `Run`, `RunCapture`, `OutputCapture` for subprocess execution and stdout capture.
- `checks.go`: `CheckFolderExists`, `CheckEnvironmentName`, `CheckMounts`.
- `dirops.go`: `CreateMountDirs` — creates missing mount source directories.
- `sanitize.go`: `SanitizeFolderPath` (expands env vars + resolves to absolute path), `SanitizeMountsPaths` (sanitizes source paths in `source:dest` mount strings), `SanitizeArgs` (parses a single-arg command with shellwords, or validates multi-arg commands for spaces).
- New utilities must be genuinely reusable across multiple callers; one-off helpers belong in the package that uses them.

---

### `internal/tests` — shared test helpers

- Holds reusable helpers for building isolated test fixtures; it is imported by `cmd/` test files (e.g. `cmd/env/env_test.go`).
- `CreateTempConfig(name, moatEnv) (string, string, error)` creates a fresh temporary config file under `tests.MoatTestDir`, initializes it via `config.InitConfig`, creates the named environment via `env.CreateEnvironment` (with `autoCreate=true` so it never blocks on a prompt), and returns the config file path, its contents, and any error.
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

## Adding a new environment operation

1. Add any new shared types to `internal/types/`.
2. Implement the operation as a function in `internal/env/{operation}.go` accepting `*viper.Viper` and `types.MoatEnv`.
3. Add a Cobra command file `cmd/env/{operation}.go` with a `CreateEnv{Operation}Cmd()` constructor that parses flags, constructs the `types.MoatEnv`, and calls the internal function.
4. Register the subcommand in `CreateEnvCmd()` in `cmd/env/env.go` via `EnvCmd.AddCommand(CreateEnv{Operation}Cmd())`.

## Adding a new runtime

1. Add any runtime-specific config type fields to `internal/types/runtimespec.go` (alongside the existing `RuntimeSpec` fields, or as a new type if the runtime is structurally different).
2. Implement `internal/runtimes/{name}.go` with a struct that satisfies the `Runtime` interface (implementing both `Run` and `Shell`).
3. Register the new runtime type string in `GetRuntime` in `internal/runtimes/runtime.go`.
4. Add a constructor function (e.g. `New{Name}RuntimeFromSpec`) in the same file, mirroring `NewApptainerRuntimeFromSpec`.

## Adding a new test

1. Create a test suite (subtype of `suite.Suite`) that encompasses more than one command at a time.
2. Create `SetupTest()`- and `TearDownTest()`-functions that set the base starting point for each test. Use `tests.CreateTempConfig` to build an isolated config fixture.
3. Create a test method `Test{Name}()` on the suite that runs a single test. Remember that each test runs sequentially, so each one starts from the same starting point specified by `SetupTest` and `TearDownTest`.
4. To execute a command, build the root command via `root.CreateRootCmd()`, set arguments with `rootCmd.SetArgs(...)`, and call `rootCmd.Execute()`. Capture stdout with `utils.OutputCapture`.
5. To verify configuration changes, call `config.InitConfig(configFile)` to get a fresh `*viper.Viper` and inspect it.

## Test creation best practices

- Try to reuse the same test suite if possible. If there would be conflicts in `SetupTest` and `TearDownTest` among different tests, create a new suite.
- When testing command line commands with flags, create individual tests for each flag combination.
- One test can contain multiple assert-statements.
- When testing commands that change configuration values, use `config.InitConfig` to get a `*viper.Viper` object and test that object for changes.
- Use `suite.T().Cleanup(...)` to remove any temporary directories created during a test.

## Verifying additions

- Run go tests with `go test ./...`. All tests must pass.
- pre-commit hooks should be run after additions to verify that everything works. This can be done with `pre-commit run --all-files`.
- Check whether new additions should be added to `AGENTS.md`.
