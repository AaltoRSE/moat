FROM ubuntu:24.04

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

ARG VSCODIUM_VERSION=1.126.04524

# Install VSCodium
RUN wget -q https://github.com/VSCodium/vscodium/releases/download/${VSCODIUM_VERSION}/codium_${VSCODIUM_VERSION}_amd64.deb && \
  apt-get update && \
  apt-get install -y ./codium_${VSCODIUM_VERSION}_amd64.deb && \
  rm -f codium_${VSCODIUM_VERSION}_amd64.deb && \
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
RUN apt-get update && \
  cd /tmp && \
  wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
  apt-get install -y ./google-chrome-stable_current_amd64.deb && \
  rm -f ./google-chrome-stable_current_amd64.deb && \
  rm -rf /var/lib/apt/lists/*

# Install basic command line tools
RUN apt-get update && \
  apt-get install -y git vim nano bash zsh jq ripgrep && \
  rm -rf /var/lib/apt/lists/*

ARG CODEX_VERSION=0.148.0

# Install codex
RUN wget -q https://github.com/openai/codex/releases/download/rust-v${CODEX_VERSION}/codex-x86_64-unknown-linux-musl.tar.gz && \
  tar -xzf codex-x86_64-unknown-linux-musl.tar.gz -C /usr/local/bin && \
  mv /usr/local/bin/codex-x86_64-unknown-linux-musl /usr/local/bin/codex && \
  chmod +x /usr/local/bin/codex && \
  rm -f codex-x86_64-unknown-linux-musl.tar.gz

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

ARG NVM_VERSION=0.40.7
ARG NODE_VERSION=24

ENV NVM_DIR=/opt/nvm

# Install nvm
RUN mkdir $NVM_DIR && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh | bash

# Install node
RUN . /opt/nvm/nvm.sh && \
    nvm install ${NODE_VERSION}

# Install pi.dev
RUN . /opt/nvm/nvm.sh && npm install -g --ignore-scripts @earendil-works/pi-coding-agent

# Install cline
RUN . /opt/nvm/nvm.sh && npm install -g cline

# Create typical mount points
RUN mkdir -p /project /scratch /local /l /m /u

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
