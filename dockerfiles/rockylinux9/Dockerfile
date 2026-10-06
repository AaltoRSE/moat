FROM rockylinux:9 AS slurm_build

ARG SLURM_VERSION=26.05.4
ARG SLURM_VERSION_GITHUB=26-05-4-1

ENV DEBIAN_FRONTEND=noninteractive

# Upgrade existing packages
RUN dnf update -y && \
  dnf upgrade -y && \
  dnf clean all

# Install development tools
RUN dnf update -y && \
  dnf install -y dnf-plugins-core && \
  dnf config-manager --set-enabled crb && \
  dnf group install -y development && \
  dnf group install -y rpm-development-tools && \
  dnf clean all

# Install Slurm dependencies
RUN dnf update -y && \
  dnf install -y mariadb-devel munge munge-libs munge-devel pam-devel perl-ExtUtils-MakeMaker perl-devel readline-devel bind-utils wget && \
  dnf clean all

# Obtain slurm source code
RUN mkdir -p /opt/slurm && \
  cd /opt/slurm && \
  wget -q https://github.com/SchedMD/slurm/releases/download/slurm-${SLURM_VERSION_GITHUB}/slurm-${SLURM_VERSION}.tar.bz2

# Build slurm
RUN cd /opt/slurm && \
  rpmbuild -ta slurm-${SLURM_VERSION}.tar.bz2

FROM rockylinux:9

# Arguments and software versions
ARG VSCODIUM_VERSION=1.135.06055
ARG CODEX_VERSION=0.159.0
ARG NVM_VERSION=0.40.8
ARG NODE_VERSION=24

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
  dnf install -y ca-certificates wget gpg bind-utils && \
  dnf clean all

# Install Microsoft's packaging key
RUN rpm --import https://packages.microsoft.com/keys/microsoft.asc && \
  echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\nautorefresh=1\ntype=rpm-md\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" | tee /etc/yum.repos.d/vscode.repo > /dev/null && \
  dnf check-update && \
  dnf install -y code && \
  dnf clean all

# Install VSCodium
RUN wget -q https://github.com/VSCodium/vscodium/releases/download/${VSCODIUM_VERSION}/codium-${VSCODIUM_VERSION}-el8.x86_64.rpm && \
  dnf install -y ./codium-${VSCODIUM_VERSION}-el8.x86_64.rpm && \
  rm -f codium-${VSCODIUM_VERSION}-el8.x86_64.rpm && \
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
RUN dnf update -y && \
  cd /tmp && \
  wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm && \
  dnf install -y ./google-chrome-stable_current_x86_64.rpm && \
  rm -f ./google-chrome-stable_current_x86_64.rpm && \
  dnf clean all

# Install basic command line tools
RUN dnf update -y && \
  dnf install -y git vim neovim emacs nano bash zsh jq ripgrep ddgr && \
  dnf clean all

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

# Install nvm
ENV NVM_DIR=/opt/nvm
RUN mkdir $NVM_DIR && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh | bash

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

# Create typical mount points
RUN mkdir -p /project /scratch /local /l /m /u

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Copy slurm binaries
RUN mkdir -p /opt/slurm-binaries

COPY --from=slurm_build /root/rpmbuild/RPMS/x86_64/*.rpm /opt/slurm-binaries

# Install slurm
RUN dnf update -y && \
  dnf install -y /opt/slurm-binaries/slurm-${SLURM_VERSION}*.rpm munge && \
  dnf clean all

ENTRYPOINT [ "/entrypoint.sh" ]
