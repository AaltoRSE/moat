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
├── main.go                     # Entry point; imports cmd sub-packages to trigger init()
├── go.mod
│
├── cmd/                        # CLI layer — Cobra commands only, no business logic
│   ├── root.go                 # RootCmd, CmdConfig (*viper.Viper), Execute(); init() registers --config/--debug flags
│   ├── config/                 # `moat config` command group
│   │   ├── config.go           # configCmd; registers with RootCmd
│   │   ├── set.go              # setCmd; calls internal/config.SetConfig()
│   │   ├── append.go           # appendCmd; calls internal/config.AppendConfig()
│   │   ├── prepend.go          # prependCmd; calls internal/config.PrependConfig()
│   │   ├── show.go             # showCmd (aliases: list, view); calls internal/config.GetConfigAsString()
│   │   ├── edit.go             # editCmd; opens config file in $EDITOR via internal/config.GetConfigFile and utils.Run
│   │   └── config_test.go      # Test suite for the config commands
│   ├── env/                    # `moat env` command group
│   │   ├── env.go              # EnvCmd; registers with RootCmd
│   │   ├── create.go           # createCmd; calls internal/env.CreateEnvironment()
│   │   ├── copy.go             # copyCmd; calls internal/env.CopyEnvironment()
│   │   ├── list.go             # listCmd; reads viper directly (pre-existing exception)
│   │   ├── show.go             # showCmd; calls internal/config.GetEnv()
│   │   ├── remove.go           # removeCmd; calls internal/env.RemoveEnvironment()
│   │   └── env_test.go         # Test suite for the env commands
│   └── run/                    # `moat run` command
│       └── run.go              # runCmd; resolves env + runtime, then calls runtime.Run()
│
├── internal/                   # Business logic; never imported by cmd/ in reverse
│   ├── types/                  # Canonical location for ALL shared types (see rules below)
│   │   ├── config.go           # Config, Defaults
│   │   ├── runtimespec.go      # RuntimeSpec (runtime config fields)
│   │   └── moatenv.go          # MoatEnv
│   ├── config/
│   │   ├── config.go           # InitConfig, WriteConfig, GetEnv, GetRuntimeSpec, GetVariableType, GetConfigAsString
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
│   │   ├── sanitize.go         # SanitizeFolderPath, SanitizeMountsPaths
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
| **CLI** | `cmd/`, `cmd/config/`, `cmd/env/`, `cmd/run/` | Cobra command registration, flag parsing, argument normalization, user-facing output. Delegates all logic to `internal/`. |
| **Logic** | `internal/...` | All business logic, I/O, config management, runtime execution. Must not import from `cmd/`. |

---

## Package-by-package conventions

### `internal/types` — shared type definitions

**All types that are used by more than one package must be defined here.** This is the single source of truth for the domain model.

- Contains only `type` declarations (structs, interfaces, type aliases). No functions, no `init()`.
- Use `go-playground/validator` struct tags (`validate:"required"`, `validate:"filepath"`, `validate:"required_if=..."`) for any struct that passes through `internal/config.validateConfig`.
- When adding a new feature that introduces a shared struct or interface, add it here first, then use it from the relevant `internal/` package.

**Current types:**

| Type | File | Purpose |
|---|---|---|
| `Config` | `config.go` | Root config struct validated by viper unmarshal |
| `Defaults` | `config.go` | Default runtime settings |
| `RuntimeSpec` | `runtimespec.go` | Config fields for any runtime (type, imageurl, cachedir, passenv) |
| `MoatEnv` | `moatenv.go` | An individual named environment (home, mounts) |

**Rule:** If you define a struct in `internal/runtimes`, `internal/env`, or any other package, and it is later referenced by a second package, move it to `internal/types`.

---

### `internal/config` — configuration management

- **Functions only.** Do not add type definitions here.
- Configuration is initialized by `InitConfig(cfgFile string) (*viper.Viper, error)`. When writing tests, this can be done multiple times. On the `cmd`-side a single configuration instance is initialized in `cmd/root` (stored in `cmd.CmdConfig`) and this will be used by all subcommands as well. `*viper.Viper` is passed as a parameter to internal functions.
- Exported surface: `InitConfig`, `WriteConfig`, `GetEnv`, `GetRuntimeSpec`, `GetVariableType`, `GetConfigAsString`, `GetConfigFile`, `SetConfig`, `AppendConfig`, `PrependConfig`.
- One file per operation: `config.go` (init/lookup), `set.go`, `append.go`, `prepend.go`.
- `SetConfig`, `AppendConfig`, and `PrependConfig` all follow the same pattern: mutate the viper value, re-validate the full `types.Config`, then persist via `WriteConfig`.
- Validation uses `go-playground/validator` and operates on `types.Config`.
- `GetEnv` returns `types.MoatEnv`; it must not return raw `map[string]interface{}` to callers.

---

### `internal/env` — environment lifecycle

- One file per operation: `create.go`, `copy.go`, `remove.go`. Add `list.go` if list logic is moved from `cmd/env/list.go`.
- Functions accept `*viper.Viper` and `types.MoatEnv` as parameters; they do not parse flags or read `os.Args`.
- May use `promptkit` for interactive confirmation prompts (user-facing only; not in functions called programmatically).
- Must call `internal/config.WriteConfig()` after any mutation to persist changes.
- Path arguments must be resolved to absolute paths with `filepath.Abs` before storing.

---

### `internal/runtimes` — runtime abstraction

- `runtime.go` defines the `Runtime` interface and the `GetRuntime(cfg *viper.Viper, name string)` factory. These must remain in this file.
- Each runtime is implemented in its own file: `apptainerruntime.go`, and future runtimes in `{name}.go`.
- Implementation-specific helper structs that are **private to one runtime** (e.g. `ApptainerImage`) may be defined in the same file as the implementation.
- The `Runtime` interface itself must stay in `runtime.go`; if it is referenced from another package, do not duplicate it — import from `internal/runtimes`.
- New runtime: add a new `{name}.go` implementing `Runtime`, then register it in `GetRuntime`.

---

### `internal/utils` — utility functions

- **Pure helpers only.** No business logic, no user prompts, no viper access.
- Must not import other `internal/` packages (no circular risk, but avoids entanglement).
- `RunArgs` struct and `Run` function for subprocess execution live here.
- `CheckFolderExists`, `CheckEnvironmentName`, `SanitizeFolderPath` and `SanitizeMountsPaths` live here.
- New utilities must be genuinely reusable across multiple callers; one-off helpers belong in the package that uses them.

---

### `internal/tests` — shared test helpers

- Holds reusable helpers for building isolated test fixtures; it is imported by `cmd/` test files (e.g. `cmd/env/env_test.go`).
- `CreateTempConfig(name, moatEnv)` creates a fresh temporary config file, initializes it via `config.InitConfig`, creates the named environment via `env.CreateEnvironment` (with `autoCreate=true` so it never blocks on a prompt), and returns the config file path and its contents.
- Helpers must be self-contained and must not depend on test-suite state; the caller is responsible for cleaning up any temporary files they create.
- If tests create temporary files or directories, these temporary files should be situated under `/tmp/moat_tests` specified in `tests.MoatTestDir`.

---

### `cmd/{group}/` — CLI commands

- Each command group (`config`, `env`, `run`) is its own directory and Go package, all named `package cmd`.
- **One `{group}.go` file** defines the group's parent `cobra.Command` and registers it with `cmd.RootCmd` in `init()`.
- **One file per subcommand** (e.g. `create.go`, `list.go`). The file registers the subcommand to the group command in `init()`.
- Flag variables are declared at package scope (see `cmd/env/create.go` pattern) or inside `init()` as captured closure variables for command-scoped flags (see `cmd/env/remove.go` pattern).
- `cmd/` files must not contain business logic. Extract any logic beyond argument marshaling into `internal/`.
- Use `log.Fatal().Msgf(...)` for unrecoverable errors; use `fmt.Println` for normal user-facing output.

---

## Naming conventions

| Category | Pattern | Examples |
|---|---|---|
| Cobra command variables | `{verb}Cmd` | `createCmd`, `listCmd`, `runCmd`, `configCmd` |
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
| `init()` setup failure (flag registration, etc.) | `panic(err)` — acceptable only in `init()` |
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
| `go.yaml.in/yaml/v3` | `internal/config`, `cmd/env/list.go` | YAML marshal/unmarshal |
| `github.com/mattn/go-shellwords` | `cmd/run` only | Shell-style word splitting for `moat run` arguments |
| `github.com/stretchr/testify` | All test codes | Test framework for creating tests |


Do not add viper access to packages other than `internal/config` without strong justification. Prefer calling `internal/config` functions instead.

---

## Adding a new environment operation

1. Add any new shared types to `internal/types/`.
2. Implement the operation as a function in `internal/env/{operation}.go` accepting `types.MoatEnv`.
3. Add a Cobra command file `cmd/env/{operation}.go` that parses flags, constructs the `types.MoatEnv`, and calls the internal function.
4. Register the subcommand in `init()` via `EnvCmd.AddCommand(...)`.

## Adding a new runtime

1. Add any runtime-specific config type fields to `internal/types/runtimespec.go` (alongside the existing `RuntimeSpec` fields, or as a new type if the runtime is structurally different).
2. Implement `internal/runtimes/{name}.go` with a struct that satisfies the `Runtime` interface.
3. Register the new runtime type string in `GetRuntime` in `internal/runtimes/runtime.go`.
4. Add a constructor function (e.g. `New{Name}RuntimeFromSpec`) in the same file, mirroring `NewApptainerRuntimeFromSpec`.

## Adding a new test

1. Create a test suite (subtype of `suite.Suite`) that encompasses more than one command at a time.
2. Create `SetupTest()`- and `TearDownTest()`-functions that set the base starting point for each test.
3. Create a test `Test{Name}()`-function that runs a single test. Remember that each test happens sequentially, so each one will start from the same starting point specified by `SetupTest` and `TearDownTest`.

## Test creation best practices

- Try to reuse same test suite if possible. If there would be conflicts in `SetupTest` and `TearDownTest` among different tests, create a new suite.
- When testing command line commands with flags, create individual tests for each flag combination.
- One test can contain multiple assert-statements.
- When testing commands that change configuration values, use `InitConfig` to get a `*viper.Viper` object and test that object for changes.

## Verifying additions

- Run go tests with `go test ./..`. All tests must pass.
- pre-commit hooks should be run after additions to verify that everything works. This can be done with `pre-commit run --all-files`.
- Check whether new additions should be added to `AGENTS.md`.
