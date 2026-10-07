# Examples

Ready-to-use example configurations for moat. Each example lives in its own
directory under `examples/` and has a README with step-by-step instructions.

| Example | Description |
|---|---|
| [Simple agents](examples/simple_agents/README.md) | Run CLI coding agents (codex, claude, opencode, cline, pi, omp) in a shared fake home with the current directory mounted into the container. |
| [Slurm agents](examples/slurm_agents/README.md) | Run CLI coding agents on a RHEL 9-based cluster where the container can submit Slurm jobs, with moat's cache and the agents' home under the project directory. |
| [VSCode / vscodium](examples/vscode/README.md) | Run VSCode or VSCodium in moat with a shared fake home and one shared project directory per editor. |
