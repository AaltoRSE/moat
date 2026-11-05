# VSCode / vscodium in moat

This example runs VSCode (`code`) or VSCodium (`codium`) inside a moat
environment. Both environments share one fake home directory, while each tool
gets its own shared project folder and runtime directory mounted into the
container. The full configuration is in [moat-config.yaml](moat-config.yaml).

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

### 2. Set up the configuration

Create the moat configuration file and open it in your editor:

```shell
moat init
moat config edit
```

`moat init` writes the default configuration to
`~/.config/moat/moat-config.yaml`, and `moat config edit` opens that file in
the editor given by the `$EDITOR` environment variable. Copy the example
configuration below into the file, replacing its contents, and save it:

```yaml
envs:
    vscode:
        command: code --wait $PWD
        home: $HOME/moat/agents-home
        mounts:
            - $HOME/moat/vscode-projects
            - /run/user/$UID/moat/vscode:/run/user/$UID
            - /run/dbus
            - /tmp/moat:/tmp
    vscodium:
        command: codium --wait $PWD
        home: $HOME/moat/agents-home
        mounts:
            - $HOME/moat/vscodium-projects
            - /run/user/$UID/moat/vscodium:/run/user/$UID
            - /run/dbus
            - /tmp/moat:/tmp
```

### 3. Open a project in the container

Create a project folder inside the shared project folder, move into it, and
run moat there:

```shell
mkdir -p ~/moat/vscode-projects/my-project
cd ~/moat/vscode-projects/my-project
moat run -n vscode
```

On the first run, moat asks whether to create the directories the
environments need but that do not exist yet — the fake home, the VSCodium
project folder, the per-editor runtime directories, and the shared
`/tmp/moat` — confirm with `y` (or start moat with `moat run -y` to create
missing directories without asking).

This runs `code --wait $PWD` (the environment's configured command): moat
expands `$PWD` to the absolute path of `my-project`, so VSCode opens that
folder inside the container, and `moat run` returns only when the editor is
closed. For VSCodium, use the `vscodium-projects` folder and
`moat run -n vscodium` instead.

## Implementation details

This section explains the choices made in [moat-config.yaml](moat-config.yaml): the shared fake home, the mounts, and the default command.

### Shared fake home

Both environments use the same fake home, `~/moat/agents-home`. VSCode and VSCodium store their settings, plugins, and extensions under the home directory, so sharing one home makes every instance load the same setup, just like when running the editor without containers.

### Mounts

moat expands environment variables in the configured mounts and command before
it starts the container, so `$HOME` is your home directory and `$UID` your
numeric user ID (the exact rules are documented in
[ENV.md](../../ENV.md#mounts--m--mount--r--ro-mount)).

- `~/moat/vscode-projects` (VSCode) and `~/moat/vscodium-projects` (VSCodium) — one shared project directory per editor, mounted at the same path in every instance. VSCode is single-instance by design: the first launched copy keeps running as a background process, and later copies in the same environment do not start their own process but ask the existing one to open a new window. Because that first process runs in the first container, it can only see that container's mounts, and the workspaces of the other instances are not visible. Keeping all projects in one directory that is mounted in every instance makes every workspace visible to every instance, and new projects can be added without changing the configuration. Each editor gets its own directory, so VSCode and VSCodium instances do not see each other's projects.
- `/run/user/$UID/moat/vscode:/run/user/$UID` (VSCode) and `/run/user/$UID/moat/vscodium:/run/user/$UID` (VSCodium) — the container's runtime directory, backed by one host directory per editor. The editors keep the sockets they use to find each other in the runtime directory: a newly started instance looks for the socket of the already-running instance there and hands it the "open this folder" request. Because every instance of an editor mounts the same host directory as its runtime directory, the socket is visible to all of them. Each editor gets its own directory, so VSCode and VSCodium instances do not share sockets. The mount sources live under your login session's runtime directory on the host, so moat can create them without administrator rights. Because `/run/user/$UID` is emptied when you log out, the two directories inside it disappear with it, and the next `moat run` asks whether to create them again.
- `/run/dbus` — the host's D-Bus system socket.
- `/tmp/moat:/tmp` — the container's `/tmp`, backed by a single host directory. The apptainer runtime otherwise gives each container a private, empty `/tmp`; this mount makes `/tmp` writable and shared by every instance of both editors, so temporary files and sockets created by one instance are visible to the others.

D-Bus is the message bus of the Linux desktop, and graphical applications use it to talk to system services such as desktop notifications and accessibility support. The editors' graphical side needs to reach these services, so the host's D-Bus system socket is mounted into the container. The container's runtime directory is a dedicated directory rather than your real one, so the sockets of your desktop session (Wayland, sound, session D-Bus) stay out of the container.

> [!NOTE]
> The shared directories limit the containment a bit: every instance of an editor can see every project in its shared project directory, and the instances share `/tmp` (and, within each editor, the runtime directory) with each other.

### Default command

Each environment's `command` is `code --wait $PWD` (or `codium --wait $PWD`); the `code` and `codium` executables are installed in the default moat container image, so no editor installation is needed. `moat run` without arguments executes this command: moat expands `$PWD` to the absolute path of the directory you run it from, so the editor opens exactly that folder, and `--wait` keeps `moat run` running until the editor is closed.

The absolute path matters because of the single-instance behavior described above: the "open this folder" request is handled by the already-running instance, which runs in the container where the editor was first started. A relative path like `.` would be resolved against that container's working directory and open the wrong folder, while the absolute `$PWD` path points to the right folder from every instance.
