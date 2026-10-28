FROM ubuntu:26.04

ARG TARGETARCH

# Arguments and software versions (an empty version resolves to the latest release on GitHub)
ARG VSCODIUM_VERSION=""
ARG CODEX_VERSION=""
ARG NVM_VERSION=""
ARG NODE_VERSION=26

ENV DEBIAN_FRONTEND=noninteractive

# Upgrade existing packages
RUN apt-get update && \
  apt-get dist-upgrade -y && \
  rm -rf /var/lib/apt/lists/*

# Install build tools
RUN apt-get update && \
  apt-get install -y build-essential && \
  rm -rf /var/lib/apt/lists/*

# Install wget and GPG for repository setup
RUN apt-get update && \
  apt-get install -y curl ca-certificates wget gpg && \
  rm -rf /var/lib/apt/lists/*

COPY slurm-binaries /opt/slurm-binaries

# List Slurm binaries for debugging
RUN ls -l /opt/slurm-binaries

# Install slurm (the debs copied from the build stage match the target architecture)
RUN ARCH="${TARGETARCH:-$(dpkg --print-architecture)}" && \
  apt-get update && \
  apt-get install -y \
    /opt/slurm-binaries/slurm-smd-client_*_${ARCH}.deb \
    /opt/slurm-binaries/slurm-smd_*_${ARCH}.deb \
    munge && \
  rm -rf /var/lib/apt/lists/*

# Install Microsoft's packaging key
RUN apt-get update && \
  apt-get install -y wget gpg && \
  wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > microsoft.gpg && \
  install -D -o root -g root -m 644 microsoft.gpg /usr/share/keyrings/microsoft.gpg && \
  rm -f microsoft.gpg && \
  rm -rf /var/lib/apt/lists/*

# Install VSCode from repo
COPY vscode.sources /etc/apt/sources.list.d/vscode.sources

RUN apt-get update && \
  apt-get install -y code && \
  rm -rf /var/lib/apt/lists/*

# Install VSCodium (latest version from GitHub when VSCODIUM_VERSION is empty)
RUN ARCH="${TARGETARCH:-$(dpkg --print-architecture)}" && \
  if [ -z "${VSCODIUM_VERSION}" ]; then \
    VSCODIUM_VERSION="$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/VSCodium/vscodium/releases/latest | sed -n 's|.*/tag/||p')" && \
    [ -n "${VSCODIUM_VERSION}" ] || { echo "Failed to resolve the latest VSCodium version from GitHub" >&2; exit 1; }; \
  fi && \
  wget -q "https://github.com/VSCodium/vscodium/releases/download/${VSCODIUM_VERSION}/codium_${VSCODIUM_VERSION}_${ARCH}.deb" && \
  apt-get update && \
  apt-get install -y "./codium_${VSCODIUM_VERSION}_${ARCH}.deb" && \
  rm -f "codium_${VSCODIUM_VERSION}_${ARCH}.deb" && \
  rm -rf /var/lib/apt/lists/*

# Add locales
RUN apt-get update && \
  apt-get install -y \
  gnome-keyring \
  locales && \
  rm -rf /var/lib/apt/lists/*

RUN sed -i -e 's/# en_US.UTF-8/en_US.UTF-8/' /etc/locale.gen && \
  sed -i -e 's/# en_GB.UTF-8/en_GB.UTF-8/' /etc/locale.gen && \
  sed -i -e 's/# fi_FI.UTF-8/fi_FI.UTF-8/' /etc/locale.gen && \
  sed -i -e 's/# fi_FI ISO-8859-1/fi_FI ISO-8859-1/' /etc/locale.gen && \
  dpkg-reconfigure locales && \
  update-locale LANG=en_US.UTF-8

# Add Google Chrome for logging into Github Copilot
RUN ARCH="${TARGETARCH:-$(dpkg --print-architecture)}" && \
  apt-get update && \
  cd /tmp && \
  wget -q "https://dl.google.com/linux/direct/google-chrome-stable_current_${ARCH}.deb" && \
  apt-get install -y "./google-chrome-stable_current_${ARCH}.deb" && \
  rm -f "./google-chrome-stable_current_${ARCH}.deb" && \
  rm -rf /var/lib/apt/lists/*

# Install basic command line tools
RUN apt-get update && \
  apt-get install -y git vim neovim emacs nano bash zsh jq ripgrep ddgr && \
  rm -rf /var/lib/apt/lists/*

# Install codex (release assets use the rust target triple: x86_64 / aarch64)
RUN case "${TARGETARCH:-$(dpkg --print-architecture)}" in \
      amd64) CODEX_ARCH="x86_64" ;; \
      arm64) CODEX_ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: ${TARGETARCH:-$(dpkg --print-architecture)}" >&2; exit 1 ;; \
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
RUN curl -L micro.mamba.pm/install.sh | PREFIX_LOCATION=/usr/local/bin sh

# Create typical mount points
RUN mkdir -p /project /scratch /local /l /m /u

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Copy slurm binaries
RUN mkdir -p /opt/slurm-binaries

ENTRYPOINT [ "/entrypoint.sh" ]
