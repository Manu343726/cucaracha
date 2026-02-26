# Multi-stage Dockerfile for Cucaracha
# Stage 1: Build LLVM/Clang from cucaracha-backend with Cucaracha target support
FROM ubuntu:22.04 as llvm-builder

ENV DEBIAN_FRONTEND=noninteractive \
    LLVM_PROJECT_DIR=/tmp/llvm-project \
    CLANG_INSTALL_DIR=/opt/clang \
    CMAKE_BUILD_PARALLEL_LEVEL=4

# Install build dependencies for LLVM
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    ninja-build \
    ccache \
    python3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Clone cucaracha-backend branch from Manu343726/llvm-project (our fork)
RUN mkdir -p ${LLVM_PROJECT_DIR} && \
    git clone --depth 1 --branch cucaracha-backend \
    https://github.com/Manu343726/llvm-project.git ${LLVM_PROJECT_DIR}

WORKDIR ${LLVM_PROJECT_DIR}

# Create build directory with cucaracha-specific LLVM configuration
# Using the same configuration as CMakePresets.json base preset with Cucaracha target
RUN mkdir -p build && cd build && \
    cmake -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX=${CLANG_INSTALL_DIR} \
    -DLLVM_ENABLE_PROJECTS="clang;clang-tools-extra" \
    -DLLVM_TARGETS_TO_BUILD="X86;Sparc" \
    -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD="LEG;Cucaracha" \
    -DCUCARACHA_GENERATE_TD=true \
    -DLLVM_BUILD_LLVM_DYLIB=ON \
    -DLLVM_LINK_LLVM_DYLIB=ON \
    -DCMAKE_CXX_FLAGS_RELEASE="-O3 -DNDEBUG" \
    ../llvm && \
    ninja -j ${CMAKE_BUILD_PARALLEL_LEVEL:-4} clang clang-tools-extra && \
    ninja install-clang install-clang-headers

# Stage 2: Minimal runtime image
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive \
    CLANG_INSTALL_DIR=/opt/clang \
    PATH=/opt/clang/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    GO_VERSION=1.25.0 \
    CGO_ENABLED=1 \
    CC=clang \
    CXX=clang++

# Install minimal runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    libc6 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    && rm -rf /var/lib/apt/lists/* && \
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz && \
    go version

# Copy Clang installation from builder
COPY --from=llvm-builder ${CLANG_INSTALL_DIR} ${CLANG_INSTALL_DIR}

# Verify clang and clang++ are available and have Cucaracha target support
RUN clang --version && clang++ --version && \
    clang --print-targets | grep -i cucaracha

# Create application directory
WORKDIR /workspace

# Copy cucaracha source (expects to be called from repo root)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build cucaracha
RUN go build -o cucaracha -ldflags="-s -w" . && \
    chmod +x cucaracha

# Install to system PATH
RUN cp cucaracha /usr/local/bin/

# Cleanup
RUN rm -rf /root/go/pkg/mod/cache

# Default entrypoint and command
ENTRYPOINT ["cucaracha"]
CMD ["--help"]
