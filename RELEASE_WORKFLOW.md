# Release Workflow Documentation

## 🚀 GitHub Release Workflow for bffgen

This document outlines the automated release process for bffgen using GitHub Actions and manual release procedures.

## 📋 Release Process Overview

### Automated Release (Recommended)

1. **Create Release Tag**: Push a version tag (e.g., `v0.1.0`)
2. **GitHub Actions**: Automatically builds, tests, and publishes release
3. **Release Assets**: Multi-platform binaries and checksums uploaded
4. **Release Notes**: Auto-generated from git commits

### Manual Release (Fallback)

1. **Local Build**: Use Makefile or release script
2. **Manual Upload**: Upload assets to GitHub release
3. **Manual Notes**: Use release notes template

## 🔧 Release Workflow Components

### 1. GitHub Actions Workflow (`.github/workflows/release.yml`)

**Triggers:**

- Push tags matching `v*` pattern
- Manual workflow dispatch with version input

**Process:**

- ✅ Checkout code with full history
- ✅ Set up Go 1.21 environment
- ✅ Validate version format
- ✅ Build for 5 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- ✅ Run tests and linter
- ✅ Generate release notes from git commits
- ✅ Create GitHub release
- ✅ Upload binaries and checksums

### 2. Version Management

**Build Variables:**

- `version`: Git tag (e.g., v0.1.0)
- `buildTime`: UTC timestamp
- `commit`: Short commit SHA

**Version Format:** `v0.1.0` (semantic versioning)

### 3. Build Artifacts

**Platforms:**

- `bffgen-linux-amd64`
- `bffgen-linux-arm64`
- `bffgen-darwin-amd64`
- `bffgen-darwin-arm64`
- `bffgen-windows-amd64.exe`

**Checksums:** `checksums.txt` with SHA256 hashes

## 🛠️ Release Commands

### Using Makefile

```bash
# Build for current platform
make build VERSION=v0.1.0

# Build for all platforms
make build-all VERSION=v0.1.0

# Run tests
make test

# Run linter
make lint

# Create and push tag
make tag VERSION=v0.1.0

# Full release preparation
make release-prep VERSION=v0.1.0
```

### Using Release Script

```bash
# Automated release process
./scripts/release.sh v0.1.0
```

### Manual Git Commands

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push tag (triggers GitHub Actions)
git push origin v0.1.0
```

## 📝 Release Notes Generation

### Automated (GitHub Actions)

- Generated from git commits since last tag
- Includes installation instructions
- Includes usage examples
- Includes security features summary

### Manual Template

- Use `RELEASE_NOTES.md` as template
- Customize for specific release
- Include breaking changes
- Include migration guides

## 🔍 Release Checklist

### Pre-Release

- [ ] All tests passing (`make test`)
- [ ] Code formatted (`make lint`)
- [ ] Version updated in code
- [ ] CHANGELOG.md updated
- [ ] README.md up to date
- [ ] Documentation current

### Release Process

- [ ] Create release tag (`v0.1.0`)
- [ ] Push tag to trigger GitHub Actions
- [ ] Verify GitHub Actions build success
- [ ] Check release assets uploaded
- [ ] Verify checksums generated
- [ ] Test release notes accuracy

### Post-Release

- [ ] Announce release on social media
- [ ] Update installation instructions
- [ ] Monitor for issues
- [ ] Plan next release

## 🚨 Troubleshooting

### GitHub Actions Failures

- Check Go version compatibility
- Verify all tests pass locally
- Check linter errors
- Verify tag format

### Build Failures

- Check Go module dependencies
- Verify platform-specific issues
- Check build flags and ldflags

### Release Asset Issues

- Verify file permissions
- Check file sizes
- Verify checksum generation

## 📊 Release Metrics

### v0.1.0 Target Metrics

- **Binary Size**: ~14MB per platform
- **Build Time**: <5 minutes
- **Test Coverage**: 100% of critical paths
- **Platform Support**: 5 platforms
- **Dependencies**: Minimal external deps

## 🔗 Release Links

- **GitHub Releases**: https://github.com/RichGod93/bffgen/releases
- **GitHub Actions**: https://github.com/RichGod93/bffgen/actions
- **Installation**: `go install github.com/RichGod93/bffgen/cmd/bffgen@latest`

## 📚 Additional Resources

- [Semantic Versioning](https://semver.org/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Release Process](https://golang.org/doc/devel/release.html)
- [Makefile Best Practices](https://makefiletutorial.com/)
