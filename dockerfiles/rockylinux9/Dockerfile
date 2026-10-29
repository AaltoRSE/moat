FROM rockylinux:9

# Target architecture, set automatically by BuildKit for multi-platform builds
# (falls back to the build host's architecture for a plain `docker build`)
ARG TARGETARCH

# Arguments and software versions (an empty version resolves to the latest release on GitHub)
ARG VSCODIUM_VERSION=""
ARG CODEX_VERSION=""
ARG NVM_VERSION=""
ARG NODE_VERSION=26
# Needed in this stage so the slurm install glob matches the copied RPMs
ARG SLURM_VERSION=26.05.4

ENV DEBIAN_FRONTEND=noninteractive

# Upgrade existing packages
RUN dnf update -y && \
  dnf upgrade -y && \
  dnf clean all

# Install build tools
RUN dnf update -y && \
  dnf install -y dnf-plugins-core && \
  dnf config-manager --set-enabled crb && \
  dnf install -y epel-release && \
  dnf group install -y development && \
  dnf clean all

# Install wget and GPG for repository setup
RUN dnf update -y && \
  dnf install -y libatomic ca-certificates curl-minimal wget gpg bind-utils && \
  dnf clean all

COPY slurm-binaries /opt/slurm-binaries

# List Slurm binaries for debugging
RUN ls -l /opt/slurm-binaries

# Install slurm (the copied RPMs match the target architecture)
RUN dnf update -y && \
  case "${TARGETARCH:-$(uname -m)}" in \
      amd64|x86_64) RPM_ARCH="x86_64" ;; \
      arm64|aarch64) RPM_ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: ${TARGETARCH:-$(uname -m)}" >&2; exit 1 ;; \
    esac && \
  dnf install -y /opt/slurm-binaries/slurm-${SLURM_VERSION}*${RPM_ARCH}.rpm munge && \
  dnf clean all

# Install Microsoft's packaging key
RUN rpm --import https://packages.microsoft.com/keys/microsoft.asc && \
  echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\nautorefresh=1\ntype=rpm-md\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" | tee /etc/yum.repos.d/vscode.repo > /dev/null && \
  dnf check-update && \
  dnf install -y code && \
  dnf clean all

# Install VSCodium (latest version from GitHub when VSCODIUM_VERSION is empty)
RUN case "${TARGETARCH:-$(uname -m)}" in \
      amd64|x86_64) RPM_ARCH="x86_64" ;; \
      arm64|aarch64) RPM_ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: ${TARGETARCH:-$(uname -m)}" >&2; exit 1 ;; \
    esac && \
  if [ -z "${VSCODIUM_VERSION}" ]; then \
    VSCODIUM_VERSION="$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/VSCodium/vscodium/releases/latest | sed -n 's|.*/tag/||p')" && \
    [ -n "${VSCODIUM_VERSION}" ] || { echo "Failed to resolve the latest VSCodium version from GitHub" >&2; exit 1; }; \
  fi && \
  wget -q "https://github.com/VSCodium/vscodium/releases/download/${VSCODIUM_VERSION}/codium-${VSCODIUM_VERSION}-el8.${RPM_ARCH}.rpm" && \
  dnf install -y "./codium-${VSCODIUM_VERSION}-el8.${RPM_ARCH}.rpm" && \
  rm -f "codium-${VSCODIUM_VERSION}-el8.${RPM_ARCH}.rpm" && \
  dnf clean all

# Add locales
RUN dnf update -y && \
  dnf install -y gnome-keyring glibc-langpack-en && \
  dnf clean all

RUN dnf update -y && \
  dnf install -y glibc-locale-source && \
  dnf clean all

RUN localedef -i en_US -f UTF-8 en_US.UTF-8 && \
  localedef -i en_GB -f UTF-8 en_GB.UTF-8 && \
  localedef -i fi_FI -f UTF-8 fi_FI.UTF-8 && \
  localedef -i fi_FI -f ISO-8859-1 fi_FI.ISO-8859-1 && \
  export LANG=en_US.UTF-8

# Add Google Chrome for logging into Github Copilot
RUN case "${TARGETARCH:-$(uname -m)}" in \
      amd64|x86_64) RPM_ARCH="x86_64" ;; \
      arm64|aarch64) RPM_ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: ${TARGETARCH:-$(uname -m)}" >&2; exit 1 ;; \
    esac && \
  dnf update -y && \
  cd /tmp && \
  wget -q "https://dl.google.com/linux/direct/google-chrome-stable_current_${RPM_ARCH}.rpm" && \
  dnf install -y "./google-chrome-stable_current_${RPM_ARCH}.rpm" && \
  rm -f "./google-chrome-stable_current_${RPM_ARCH}.rpm" && \
  dnf clean all

# Install basic command line tools
RUN dnf update -y && \
  dnf install -y git vim neovim emacs nano bash zsh jq ripgrep ddgr && \
  dnf clean all

# Install codex (release assets use the rust target triple: x86_64 / aarch64)
RUN case "${TARGETARCH:-$(uname -m)}" in \
      amd64|x86_64) CODEX_ARCH="x86_64" ;; \
      arm64|aarch64) CODEX_ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: ${TARGETARCH:-$(uname -m)}" >&2; exit 1 ;; \
    esac && \
  if [ -z "${CODEX_VERSION}" ]; then \
    CODEX_TAG="$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/openai/codex/releases/latest | sed -n 's|.*/tag/||p')" && \
    CODEX_VERSION="${CODEX_TAG#rust-v}" && \
    [ -n "${CODEX_VERSION}" ] || { echo "Failed to resolve the latest codex version from GitHub" >&2; exit 1; }; \
  fi && \
  wget -q "https://github.com/openai/codex/releases/download/rust-v${CODEX_VERSION}/codex-${CODEX_ARCH}-unknown-linux-musl.tar.gz" && \
  tar -xzf "codex-${CODEX_ARCH}-unknown-linux-musl.tar.gz" -C /usr/local/bin && \
  mv "/usr/local/bin/codex-${CODEX_ARCH}-unknown-linux-musl" /usr/local/bin/codex && \
  chmod +x /usr/local/bin/codex && \
  rm -f "codex-${CODEX_ARCH}-unknown-linux-musl.tar.gz"

# Install claude
RUN curl -fsSL https://claude.ai/install.sh | bash && \
  CLAUDE_TARGET=`readlink /root/.local/bin/claude | sed 's:/root/.local:/usr/local:g'` && \
  rm /root/.local/bin/claude && \
  mv /root/.local/share/claude /usr/local/share/claude && \
  ln -s $CLAUDE_TARGET /usr/local/bin/claude

# Install OpenCode
ENV OPENCODE_INSTALL_DIR=/usr/local/bin
ENV OPENCODE_DISABLE_AUTOUPDATE=1
RUN curl -fsSL https://opencode.ai/install | \
    sed 's/INSTALL_DIR=.*/INSTALL_DIR=$OPENCODE_INSTALL_DIR/' | \
    bash

# Install goose CLI
RUN curl -fsSL https://github.com/aaif-goose/goose/releases/download/stable/download_cli.sh | CONFIGURE=false GOOSE_BIN_DIR=/usr/local/bin bash

# Install nvm (latest version from GitHub when NVM_VERSION is empty)
ENV NVM_DIR=/opt/nvm
RUN mkdir $NVM_DIR && \
  if [ -z "${NVM_VERSION}" ]; then \
    NVM_TAG="$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/nvm-sh/nvm/releases/latest | sed -n 's|.*/tag/||p')" && \
    NVM_VERSION="${NVM_TAG#v}" && \
    [ -n "${NVM_VERSION}" ] || { echo "Failed to resolve the latest nvm version from GitHub" >&2; exit 1; }; \
  fi && \
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh | bash

# Install node
RUN . /opt/nvm/nvm.sh && \
    nvm install ${NODE_VERSION}

# Install pi.dev
RUN . /opt/nvm/nvm.sh && npm install -g --ignore-scripts @earendil-works/pi-coding-agent

# Install cline
RUN . /opt/nvm/nvm.sh && npm install -g cline

# Install bun
RUN . /opt/nvm/nvm.sh && npm install -g bun

# Install oh-my-pi
RUN . /opt/nvm/nvm.sh && npm install -g @oh-my-pi/pi-coding-agent

# Install uv
RUN curl -LsSf https://astral.sh/uv/install.sh | XDG_BIN_HOME=/usr/local/bin sh

# Install micromamba
RUN curl -L micro.mamba.pm/install.sh | SHELL=/bin/bash INIT_YES=no BIN_FOLDER=/usr/local/bin bash

# Create typical mount points
RUN mkdir -p /project /scratch /local /l /m /u

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Copy slurm binaries (the build stage built them for the target architecture)
RUN mkdir -p /opt/slurm-binaries

ENTRYPOINT [ "/entrypoint.sh" ]
