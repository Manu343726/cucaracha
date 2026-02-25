# GitHub Actions Workflows for Cucaracha

This document explains the GitHub Actions CI/CD setup for Cucaracha.

## Workflows Overview

### 1. `quick-tests.yaml` - Fast Feedback Loop
**Triggers**: Push to main/develop, Pull requests  
**Duration**: ~5-10 minutes  
**Purpose**: Quick validation of code changes

**Steps**:
- Download Go dependencies
- Run code linting (golangci-lint)
- Format checking (gofmt)
- Static analysis (go vet)
- Build the binary
- Test all CLI commands (--help)
- Run unit tests (quick subset)

**Benefits**: 
- Fast feedback on every commit
- Catches formatting issues early
- No Docker overhead

### 2. `build-and-test.yaml` - Comprehensive Docker Build
**Triggers**: Push to main/develop, PR to main/develop, Weekly schedule  
**Duration**: 30-60 minutes (first run), 10-15 minutes (with cache)  
**Purpose**: Full Docker build with LLVM/Clang integration

**Steps**:
1. **Build Docker Image**:
   - Compiles LLVM/Clang from `cucaracha-backend` branch
   - Installs Go and dependencies
   - Builds Cucaracha binary
   - Uses layer caching for faster rebuilds

2. **CLI Testing**:
   - Validates all subcommands work
   - Tests: `--help` for each command

3. **Unit Tests**:
   - Runs full test suite (excluding integration tests)
   - Captures coverage information
   - ~30 second timeout

4. **REPL Testing**:
   - Tests basic REPL commands:
     - `load-system default`
     - `load-runtime interpreter`
     - `help`

5. **Docker Registry Push** (main branch only):
   - Publishes image to GitHub Container Registry
   - Tags: branch name, git commit, version tags

**Benefits**:
- Tests in production-like environment
- Ensures LLVM integration works
- Publishes ready-to-use Docker images

## Docker Build Process

### Multi-Stage Build

```
Stage 1: llvm-builder
├── Ubuntu 22.04 base
├── Install LLVM build tools
├── Clone llvm-project (cucaracha-backend branch)
└── Compile and install Clang to /opt/clang

Stage 2: runtime
├── Ubuntu 22.04 base
├── Install Go 1.24.0
├── Copy Clang from builder
├── Build Cucaracha binary
└── Set as container entrypoint
```

### Build Optimizations

- **Shallow clone**: Only fetches necessary history
- **Incremental builds**: Uses CMake and Ninja for faster compilation
- **GitHub Actions Cache**: Reuses build layers across runs
- **Parallel builds**: Uses 4 parallel build jobs
- **Minimal final image**: Only includes runtime dependencies

## Running Workflows Locally

### Quick Tests (without Docker)

```bash
# Install Go 1.24.0
# Then run:
go mod download
go test ./... -v --skip TestIntegration -timeout 30s
go build -o cucaracha .
./cucaracha --help
```

### Docker-based Tests

```bash
# Build image
DOCKER_BUILDKIT=1 docker build -t cucaracha:latest .

# Run tests in container
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  cucaracha:latest \
  go test ./... -v --skip TestIntegration -timeout 30s

# Test CLI
docker run --rm cucaracha:latest --help
docker run --rm cucaracha:latest debug --help
```

## Accessing Build Results

### Test Results

For pull requests:
- GitHub Actions tab shows pass/fail summary
- Click on workflow run for detailed logs
- Test output is available in "Run unit tests" step

### Docker Images

For main branch builds:
- Published to: `ghcr.io/Manu343726/cucaracha:<tag>`
- Tags: branch name, git SHA, version
- Example: 
  ```bash
  docker pull ghcr.io/Manu343726/cucaracha:main
  docker pull ghcr.io/Manu343726/cucaracha:sha-abc1234
  ```

### Cache Performance

The GitHub Actions cache significantly improves build times:

**First build** (no cache): ~45-60 minutes
- Clones and builds entire LLVM project
- Downloads all Go dependencies

**Subsequent builds** (with cache): ~5-15 minutes
- Reuses LLVM build layers
- Reuses Go dependency layers
- Only rebuilds changed files

## Configuration

### Branch Protection Rules (Recommended)

Add these status checks as required to enforce before merging:

1. `Quick Tests / quick-tests` - Must pass
2. `Build Docker / build-docker` - Should pass (can be slow)

### Secrets

Currently requires:
- Default GitHub token (GITHUB_TOKEN) - automatically provided

Optional for registry push:
- GitHub PAT with `write:packages` scope (optional, uses GITHUB_TOKEN by default)

## Customization

### Changing Go Version

Edit `.github/workflows/quick-tests.yaml`:
```yaml
- uses: actions/setup-go@v4
  with:
    go-version: 1.25.0  # Change this
```

### Changing LLVM Branch

Edit `Dockerfile`:
```dockerfile
git clone --depth 1 --branch YOUR_BRANCH \
  https://github.com/llvm/llvm-project.git ...
```

### Running Weekly Scheduled Builds

Already enabled in `build-and-test.yaml`:
```yaml
schedule:
  - cron: '0 0 * * 0'  # Every Sunday at midnight UTC
```

## Troubleshooting

### Workflow Failures

**Problem**: "Setup failed: Node 18 not found"
- **Solution**: Workflows automatically use GitHub-hosted runner with Node.js

**Problem**: "Docker image build timed out"
- **Solution**: Usually due to LLVM compilation taking >60min
- **Action**: Rerun job, or check LLVM build complexity

**Problem**: "Test failures after code change"
- **Solution**: 
  - Check quick-tests output first
  - If only in Docker: may be environment-specific
  - Common issue: missing test cleanup in failed tests

### Caching Issues

**Problem**: Stale cache causing failures
- **Solution**: Add `cache-invalidation-suffix` with date
- Or manually clear cache in GitHub Actions settings

### LLVM Branch Not Found

**Problem**: "fatal: reference is not a tree"
- **Solution**: The cucaracha-backend branch doesn't exist
- **Action**: Update Dockerfile to use different branch or commit

## Monitoring

### View Workflow Status

1. Go to Actions tab in GitHub
2. Click on workflow name
3. See list of runs with status
4. Click run number to see details

### Notifications

Default GitHub notifications:
- PR checks fail: Notifications enabled by default
- Main branch build fails: Must configure in GitHub settings

### Metrics

Monitor these metrics:
- Build time trends
- Test pass rate
- Docker image size (Registry > Packages)
- Cache hit rate

## Future Enhancements

- [ ] Automatic benchmark comparisons
- [ ] Code coverage reporting (Codecov integration)
- [ ] Multi-platform builds (ARM64)
- [ ] Artifact uploads for release builds
- [ ] Integration test automation (currently skipped)
- [ ] Performance regression detection
- [ ] Dependency vulnerability scanning (Dependabot)
- [ ] Slack/Discord notifications

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker/build-push-action](https://github.com/docker/build-push-action)
- [Golang GitHub Actions](https://github.com/actions/setup-go)
- [golangci-lint Action](https://github.com/golangci/golangci-lint-action)
