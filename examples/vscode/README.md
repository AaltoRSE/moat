# VSCode / vscodium in moat

This example runs VSCode (`code`) or VSCodium (`codium`) inside a moat
environment. Both environments share one fake home directory, while each tool
gets its own shared project folder mounted into the container. The full
configuration is in [moat-config.yaml](moat-config.yaml).

## Motivation

When you use the editor without containers, running it in a project folder opens a window with your usual settings, plugins, and extensions, and running it in another folder opens another window with the same setup but a different workspace.

This example gives you the same experience inside moat: you can launch as many instances of VSCode or VSCodium as you like with `moat run`, all sharing one set of settings and extensions from the shared fake home while each instance loads a different workspace. To make workspaces work across instances, all projects must live in the shared project directory that is mounted in every instance; see [Implementation details](#implementation-details) for the reasoning and the containment trade-off.

Use this example when you want the full graphical editor, with your normal settings, plugins, and extensions, running inside a moat environment.

## Trying out the configuration

### 1. Install the latest moat release

```shell
curl -LO https://github.com/AaltoRSE/moat/releases/latest/download/moat_$(uname -s)_$(uname -m).tar.gz
tar -xzf moat_$(uname -s)_$(uname -m).tar.gz
mkdir -p ~/.local/bin
install moat ~/.local/bin/moat
```

Make sure `~/.local/bin` is in your `PATH`.

### 2. Install the example configuration

```shell
mkdir -p ~/.config/moat
curl -fsSL -o ~/.config/moat/moat-config.yaml https://raw.githubusercontent.com/AaltoRSE/moat/main/examples/vscode/moat-config.yaml
```

### 3. Create the folders

The environments expect a shared fake home and one shared project folder per
tool:

```shell
mkdir -p ~/moat/agents-home ~/moat/vscode-projects ~/moat/vscodium-projects
```

### 4. Open a project in the container

Create a project folder inside the shared project folder, move into it, and
run moat there:

```shell
cd ~/moat/vscode-projects
mkdir my-project
cd my-project
moat run -n vscode
```

This starts `code --wait .` (the environment's configured command), so VSCode
opens the current directory inside the container and `moat run` returns only
when the editor is closed. For VSCodium, use the `vscodium-projects` folder and
`moat run -n vscodium` instead.

## Implementation details

This section explains the choices made in [moat-config.yaml](moat-config.yaml): the shared fake home, the mounts, and the default command.

### Shared fake home

Both environments use the same fake home, `~/moat/agents-home`. VSCode and VSCodium store their settings, plugins, and extensions under the home directory, so sharing one home makes every instance load the same setup, just like when running the editor without containers.

### Mounts

- `~/moat/vscode-projects` (VSCode) and `~/moat/vscodium-projects` (VSCodium) — one shared project directory per editor, mounted at the same path in every instance. VSCode is single-instance by design: the first launched copy keeps running as a background process, and later copies in the same environment do not start their own process but ask the existing one to open a new window. Because that first process runs in the first container, it can only see that container's mounts, and the workspaces of the other instances are not visible. Keeping all projects in one directory that is mounted in every instance makes every workspace visible to every instance, and new projects can be added without changing the configuration. Each editor gets its own directory, so VSCode and VSCodium instances do not see each other's projects.
- `/run/user` — the host's per-user runtime directory, which contains the D-Bus session socket.
- `/run/dbus` — the host's D-Bus system socket.

D-Bus is the message bus of the Linux desktop, and graphical applications use it to talk to system services such as desktop notifications and accessibility support. VSCode's graphical side needs to reach these services, so the host's session and system D-Bus sockets are mounted into the container.

> [!NOTE]
> The shared project directory limits the containment a bit: every instance of an editor can see every project in its shared project directory.

### Default command

Each environment's `command` is `code --wait .` (or `codium --wait .`). `moat run` without arguments executes this command, so running moat in a project folder opens that folder in the editor, and `--wait` keeps `moat run` running until the editor is closed.
