FROM ubuntu:22.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install base utilities and dependencies
RUN apt-get update && apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    htop \
    lsb-release \
    software-properties-common \
    sudo \
    wget \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install common development tools
RUN apt-get update && apt-get install -y \
    build-essential \
    ninja-build \
    ccache \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install latest GCC (GCC 13 - latest available for Ubuntu 22.04)
RUN add-apt-repository ppa:ubuntu-toolchain-r/test && \
    apt-get update && apt-get install -y \
    gcc-13 \
    g++-13 \
    && update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-13 100 \
    && update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-13 100 \
    && rm -rf /var/lib/apt/lists/*

# Install latest Clang/LLVM (Clang 19)
RUN curl -fsSL https://apt.llvm.org/llvm-snapshot.gpg.key | gpg --dearmor -o /usr/share/keyrings/llvm-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/llvm-archive-keyring.gpg] https://apt.llvm.org/jammy/ llvm-toolchain-jammy-19 main" > /etc/apt/sources.list.d/llvm.list && \
    apt-get update && apt-get install -y \
    clang-19 \
    lld-19 \
    lldb-19 \
    clang-tools-19 \
    clang-format-19 \
    && update-alternatives --install /usr/bin/clang clang /usr/bin/clang-19 100 \
    && update-alternatives --install /usr/bin/clang++ clang++ /usr/bin/clang++-19 100 \
    && update-alternatives --install /usr/bin/lld lld /usr/bin/lld-19 100 \
    && update-alternatives --install /usr/bin/lldb lldb /usr/bin/lldb-19 100 \
    && update-alternatives --install /usr/bin/clang-format clang-format /usr/bin/clang-format-19 100 \
    && rm -rf /var/lib/apt/lists/*

# Install CMake 3.31 (latest 3.x version, not 4.x)
ARG CMAKE_VERSION=3.31.4
RUN curl -fsSL https://github.com/Kitware/CMake/releases/download/v${CMAKE_VERSION}/cmake-${CMAKE_VERSION}-linux-x86_64.tar.gz \
    | tar -xzf - -C /usr/local --strip-components=1

# Install Python 3.11
RUN add-apt-repository ppa:deadsnakes/ppa && \
    apt-get update && apt-get install -y \
    python3.11 \
    python3.11-venv \
    python3.11-dev \
    python3-pip \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.11 1 \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3.11 1 \
    && rm -rf /var/lib/apt/lists/*

# Install Go (latest stable)
ARG GO_VERSION=1.22.0
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt

ENV GOPATH=/home/vscode/go
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"

# Install zsh and oh-my-zsh
RUN apt-get update && apt-get install -y zsh \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user 'vscode' with sudo access
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Set up zsh as default shell for vscode user
RUN chsh -s /usr/bin/zsh $USERNAME

# Switch to non-root user
USER $USERNAME
WORKDIR /home/$USERNAME

# Install oh-my-zsh for vscode user
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# Create Go workspace directory
RUN mkdir -p $GOPATH/bin $GOPATH/src $GOPATH/pkg

# Set default shell to zsh
ENV SHELL=/usr/bin/zsh

# Reset to interactive frontend
ENV DEBIAN_FRONTEND=dialog
