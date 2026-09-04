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
├── moat-config.yaml           # Default config file loaded by viper
├── go.mod
│
├── cmd/                        # CLI layer — Cobra commands only, no business logic
│   ├── root.go                 # RootCmd definition and Execute(); calls internal/config.InitConfig()
│   ├── config/                 # `moat config` command group
│   │   ├── config.go           # configCmd; registers with RootCmd
│   │   ├── set.go              # setCmd stub
│   │   └── view.go             # viewCmd; calls internal/config.GetConfigAsString()
│   ├── env/                    # `moat env` command group
│   │   ├── env.go              # envCmd; registers with RootCmd
│   │   ├── create.go           # createCmd; calls internal/env.CreateEnvironment()
│   │   ├── list.go             # listCmd; reads viper directly (pre-existing exception)
│   │   └── remove.go           # removeCmd; calls internal/env.RemoveEnvironment()
│   └── run/                    # `moat run` command
│       └── run.go              # runCmd; resolves env + runtime, then calls runtime.Run()
│
├── internal/                   # Business logic; never imported by cmd/ in reverse
│   ├── types/                  # Canonical location for ALL shared types (see rules below)
│   │   ├── config.go           # Config, Defaults
│   │   ├── runtimespec.go      # RuntimeSpec (runtime config fields)
│   │   └── moatenv.go         # MoatEnv
│   ├── config/
│   │   └── config.go           # InitConfig, WriteConfig, GetEnv, GetConfigAsString
│   ├── env/
│   │   ├── create.go           # CreateEnvironment
│   │   └── remove.go           # RemoveEnvironment
│   ├── runtimes/
│   │   ├── runtime.go          # Runtime interface + GetRuntime factory
│   │   └── apptainerruntime.go # ApptainerRuntime implementation
│   └── utils/
│       ├── checks.go           # CheckFolderExists, CheckEnvironmentName
│       ├── sanitize.go         # SanitizeFolderPath, SanitizeMountsPaths
│       └── run.go              # Run, RunArgs
│
└── docs/
    └── code-structure.md
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
- All configuration state lives in viper's global instance. Do not pass `*viper.Viper` as a parameter.
- Exported surface: `InitConfig()`, `WriteConfig()`, `GetEnv(name string)`, `GetConfigAsString()`.
- Validation uses `go-playground/validator` and operates on `types.Config`.
- `GetEnv` returns `types.MoatEnv`; it must not return raw `map[string]interface{}` to callers.

---

### `internal/env` — environment lifecycle

- One file per operation: `create.go`, `remove.go`. Add `list.go` if list logic is moved from `cmd/env/list.go`.
- Functions accept `types.MoatEnv` as parameter; they do not parse flags or read `os.Args`.
- May use `promptkit` for interactive confirmation prompts (user-facing only; not in functions called programmatically).
- Must call `internal/config.WriteConfig()` after any mutation to persist changes.
- Path arguments must be resolved to absolute paths with `filepath.Abs` before storing.

---

### `internal/runtimes` — runtime abstraction

- `runtime.go` defines the `Runtime` interface and the `GetRuntime(name string)` factory. These must remain in this file.
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

### `cmd/{group}/` — CLI commands

- Each command group (`config`, `env`, `run`) is its own directory and Go package, all named `package cmd`.
- **One `{group}.go` file** defines the group's parent `cobra.Command` and registers it with `cmd.RootCmd` in `init()`.
- **One file per subcommand** (e.g. `create.go`, `list.go`). The file registers the subcommand to the group command in `init()`.
- Flag variables are declared at package scope for persistent flags, or inside `init()` as captured closure variables for command-scoped flags (see `cmd/env/remove.go` pattern).
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

## Verifying additions

pre-commit hooks should be run after additions to verify that everything works. This can be done with `pre-commit run --all-files`.
