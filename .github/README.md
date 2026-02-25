# Cucaracha GitHub Actions Setup Summary

This setup provides comprehensive CI/CD automation for the Cucaracha project with Docker-based environment management.

## Files Created

### Docker Configuration
- **`Dockerfile`** - Multi-stage build for Cucaracha with LLVM/Clang
  - Stage 1: Builds Clang from llvm-project (cucaracha-backend branch)
  - Stage 2: Minimal runtime with Go, Clang, and Cucaracha binary
  - Optimized for CI/CD with layer caching

- **`.dockerignore`** - Excludes unnecessary files from Docker build context

### GitHub Actions Workflows

- **`.github/workflows/quick-tests.yaml`** - Fast validation on every commit
  - Linting, formatting, basic tests
  - Duration: 5-10 minutes
  - Runs on: push to main/develop, pull requests

- **`.github/workflows/build-and-test.yaml`** - Complete Docker build and comprehensive testing
  - Builds LLVM/Clang from source
  - Full unit tests and CLI validation
  - REPL functionality testing
  - Publishes to GitHub Container Registry
  - Duration: 30-60 min (first), 10-15 min (cached)
  - Runs on: push to main/develop, pull requests, weekly schedule

### Documentation

- **`.github/WORKFLOWS.md`** - Detailed workflow documentation
- **`.github/CI_CD_README.md`** - Docker and CI/CD usage guide
- **`.github/workflows/test-repl.sh`** - Reusable REPL test script

## What It Does

### 1. Quick Tests (every commit)
```
Code changes → Formatting check → Lint → Build → Tests → Result
```
- Tests locally without Docker overhead
- ~5-10 minute feedback loop
- Perfect for rapid development

### 2. Docker Build (on relevant changes)
```
Docker build → LLVM compile → Clang install → Go setup → 
Build Cucaracha → Run all tests → Push to registry
```
- Tests in production-like environment
- Validates LLVM/Clang integration
- Publishes container to GitHub Container Registry
- ~30-60 minutes first time, ~10-15 minutes with cache

### 3. CLI Testing
All these commands are automatically tested:
- `cucaracha --help`
- `cucaracha debug --help`
- `cucaracha cpu --help`
- `cucaracha tools --help`
- `cucaracha tui --help`

### 4. REPL Testing
Tests basic debugger REPL commands:
- `load-system default`
- `load-runtime interpreter`
- `help`

### 5. Unit Tests
- Runs full test suite (excluding long-running integration tests)
- 30-second timeout
- Coverage tracking

## Using the Docker Image

### Local Development

Build the image:
```bash
docker build -t cucaracha:latest .
```

Run Cucaracha in Docker:
```bash
# Show help
docker run --rm cucaracha:latest --help

# Run debugger REPL
docker run --rm -it cucaracha:latest debug

# Run CPU emulator
docker run --rm cucaracha:latest cpu --help

# Run tests
docker run --rm cucaracha:latest sh -c 'go test ./... -v'
```

Mount local source for development:
```bash
docker run --rm -it \
  -v $(pwd):/workspace \
  -w /workspace \
  cucaracha:latest \
  bash
```

### CI/CD Environment

The workflows use the Docker image to:
1. Ensure consistent build environment
2. Test with actual LLVM/Clang toolchain
3. Verify all dependencies are included
4. Build portable artifacts

### Publishing Images

After successful builds on main branch, images are published to:
```
ghcr.io/Manu343726/cucaracha:latest
ghcr.io/Manu343726/cucaracha:main
ghcr.io/Manu343726/cucaracha:sha-<commit>
```

## Build Pipeline

```
┌──────────────────┐
│  Code Push       │
└────────┬─────────┘
         │
    ┌────▼────────────────────────────┐
    │   Quick Tests (5-10 min)        │
    │  - Linting & Formatting         │
    │  - Build Binary                 │
    │  - Unit Tests                   │
    │  - CLI Validation               │
    └────┬─────────────────┬──────────┘
         │ Pass            │ Fail
         │                 └─► Notify (Stop)
         │
    ┌────▼──────────────────────────────┐
    │   Docker Build (30-60 min first)  │
    │  - LLVM/Clang Compilation        │
    │  - Docker Image Build             │
    │  - Full Integration Tests        │
    │  - REPL Testing                  │
    └────┬─────────────────┬───────────┘
         │ Pass            │ Fail
         │                 └─► Notify (Stop)
         │
    ┌────▼──────────────────────┐
    │  Push to Registry         │
    │  (main branch only)       │
    └───────────────────────────┘
```

## Performance Tips

### First Build
- Takes 45-60 minutes due to LLVM compilation
- This is normal and expected
- Subsequent builds use caching

### Cached Builds
- Takes 10-15 minutes
- LLVM layers are cached across runs
- Go dependency layers are cached
- Only changed code is recompiled

### Local Development
- Use `quick-tests.yaml` commands for rapid feedback
- Skip Docker for most development
- Use Docker for final validation

## Monitoring & Debugging

### View Workflow Status
1. Go to GitHub repository → Actions tab
2. Click on workflow name
3. View run history and details

### Check Test Results
- Click on failed run
- Expand failed step
- Read full error output
- Check logs for specific failures

### Debug Docker Build
```bash
# Build locally with verbose output
DOCKER_BUILDKIT=1 docker build --progress=plain -t cucaracha:test .

# Run interactive shell in failed image
docker run --rm -it cucaracha:test bash

# Check clang installation
docker run --rm cucaracha:test clang++ --version
```

## Customization

### Add New Tests
1. Edit `.github/workflows/quick-tests.yaml` or `build-and-test.yaml`
2. Add new step under appropriate job
3. Commit and watch run in Actions tab

### Change Go Version
Edit both workflow files:
```yaml
- uses: actions/setup-go@v4
  with:
    go-version: 1.25.0  # Change here
```

And Dockerfile:
```dockerfile
GO_VERSION=1.25.0
```

### Configuration for Your Fork
If you fork this repository:

1. **Container Registry**:
   - Update `IMAGE_NAME` in `build-and-test.yaml`
   - Or disable registry push if not needed

2. **Branches**:
   - Update branch names in workflows if different
   - These workflows reference main/develop branches

3. **Notifications**:
   - Configure branch protection rules in Settings > Branches
   - Set workflows as required checks

## Future Enhancements

Potential improvements to add:
- [ ] Code coverage tracking (Codecov)
- [ ] Performance benchmarks
- [ ] Multi-platform builds (ARM64)
- [ ] Artifact uploads for releases
- [ ] Integration test automation
- [ ] Dependency scanning
- [ ] SLSA provenance generation
- [ ] Container image signing

## Troubleshooting

### Build Times Out (>6 hours)
- GitHub Actions timeout on large builds
- LLVM compilation can be very long
- Solution: Check if LLVM build error, not timeout
- Or split into separate jobs

### Out of Disk Space
- LLVM build can use 5GB+
- Solution: Use GitHub-hosted runners with more storage
- Or optimize LLVM build (fewer targets, release only)

### Tests Fail in Docker but Pass Locally
- Environment difference issue
- Check: Go version, dependencies, paths
- Solution: Use same versions as Dockerfile

## Support

For issues or questions:
1. Check the detailed docs:
   - `.github/WORKFLOWS.md` - Workflow details
   - `.github/CI_CD_README.md` - Docker & CI specifics

2. View GitHub Actions logs for specific errors

3. Test locally with Docker before pushing

---

**Created**: February 25, 2026
**GitHub Actions Version**: v4
**Go Version**: 1.24.0
**Base OS**: Ubuntu 22.04
