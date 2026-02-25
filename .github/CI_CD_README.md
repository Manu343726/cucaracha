# Cucaracha CI/CD Setup

This directory contains GitHub Actions workflows and Docker configuration for building and testing Cucaracha.

## Dockerfile

The `Dockerfile` uses a multi-stage build process:

1. **Stage 1 (llvm-builder)**: Builds LLVM/Clang from the `cucaracha-backend` branch
   - Clones the llvm-project repository
   - Builds only the Clang compiler (optimized for size and speed)
   - Installs Clang to `/opt/clang`

2. **Stage 2 (runtime)**: Creates the final Cucaracha image
   - Sets up minimal Ubuntu 22.04 runtime environment
   - Installs Go 1.24.0
   - Copies Clang installation from builder stage
   - Builds Cucaracha binary
   - Sets Clang as the default C/C++ compiler

## Building Locally

### Build the Docker image:

```bash
docker build -t cucaracha:latest .
```

### Run Cucaracha in Docker:

```bash
# Show help
docker run --rm cucaracha:latest --help

# Run debug REPL
docker run --rm -it cucaracha:latest debug

# Run CPU emulator
docker run --rm cucaracha:latest cpu --help

# Run tools
docker run --rm cucaracha:latest tools --help
```

### Run tests in Docker:

```bash
docker run --rm cucaracha:latest sh -c 'go test ./... -v --skip TestIntegration'
```

### Mount local source for development:

```bash
docker run --rm -it \
  -v $(pwd):/workspace \
  -w /workspace \
  cucaracha:latest \
  bash
```

## GitHub Actions Workflows

### build-and-test.yaml

The main CI/CD workflow that:

1. **Builds Docker Image**: Constructs the container with LLVM and Go dependencies
2. **Tests CLI Commands**: Verifies all subcommands have working help
   - `cucaracha --help`
   - `cucaracha debug --help`
   - `cucaracha cpu --help`
   - `cucaracha tools --help`
   - `cucaracha tui --help`

3. **Runs Unit Tests**: Executes Go test suite (excluding slow integration tests)
4. **Tests REPL**: Validates REPL command parsing with sample commands
5. **Quality Checks**:
   - Code formatting with gofmt
   - Static analysis with golangci-lint
   - Security scanning with gosec
   - Go vet analysis

### Triggers

The workflow runs on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

### Output

- Test results are captured and displayed
- Docker image is pushed to GitHub Container Registry on successful builds (for non-PR runs)
- Quality check reports are generated

## Environment Variables

- `GO_VERSION`: Go version to install (default: 1.24.0)
- `CMAKE_BUILD_PARALLEL_LEVEL`: Number of parallel build jobs for LLVM
- `CC`: Set to `clang` for C compilation
- `CXX`: Set to `clang++` for C++ compilation
- `CGO_ENABLED`: Enabled for Go CGO support

## Optimization Tips

### Faster Local Builds

1. **Use BuildKit for better caching**:
   ```bash
   DOCKER_BUILDKIT=1 docker build -t cucaracha:latest .
   ```

2. **Skip LLVM rebuild if you have a cached image**:
   ```bash
   docker build -t cucaracha:latest .
   ```

### Debugging Build Issues

```bash
# Build with verbose output
DOCKER_BUILDKIT=1 docker build -t cucaracha:latest --progress=plain .

# Run interactive shell in image
docker run --rm -it cucaracha:latest bash

# Check clang installation
docker run --rm cucaracha:latest clang -v
```

## Troubleshooting

### Build Timeout

The LLVM build can take 30-60 minutes depending on system resources. GitHub Actions provides 6 hours of build time.

**Solution**: Use GitHub Actions cache to speed up rebuilds:

```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Out of Disk Space

The Docker image can be 3-5GB depending on LLVM build artifacts.

**Solution**: 
- Increase available disk space in runners
- Consider using GitHub-hosted runners with more storage
- Remove build artifacts after installing

### LLVM Branch Not Found

If `cucaracha-backend` branch doesn't exist, the build will fall back to cloning the main branch.

**Solution**: Ensure the branch exists in the llvm-project repository, or update the Dockerfile to use a specific commit.

## Testing

### Local Test Commands

```bash
# Build image
docker build -t cucaracha:latest .

# Run unit tests
docker run --rm cucaracha:latest sh -c 'go test ./... -v --skip TestIntegration -timeout 30s'

# Run REPL test
docker run --rm -i cucaracha:latest debug << 'EOF'
load-system default
load-runtime interpreter
help
exit
EOF

# Check all command helps
for cmd in "" "debug" "cpu" "tools" "tui"; do
  echo "Testing: cucaracha $cmd --help"
  docker run --rm cucaracha:latest $cmd --help || true
done
```

## Future Improvements

- [ ] Add ability to cache LLVM builds across CI runs
- [ ] Create separate lightweight image without LLVM for quick tests
- [ ] Add coverage reports
- [ ] Add performance benchmarks
- [ ] Integration tests with actual C program compilation
- [ ] Multi-platform builds (ARM64, etc.)
