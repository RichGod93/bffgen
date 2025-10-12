# npm Package Testing Guide

## Local Testing Before Publishing

### 1. Build Go Binaries

First, ensure you have the binaries built:

```bash
cd ..
make build-all VERSION=v1.2.0
ls -lh dist/

# Should see:
# - bffgen-darwin-amd64
# - bffgen-darwin-arm64
# - bffgen-linux-amd64
# - bffgen-linux-arm64
# - bffgen-windows-amd64.exe
# - checksums.txt
```

### 2. Create GitHub Release (or Mock)

The install script downloads from GitHub Releases. For testing:

**Option A: Create actual pre-release**

```bash
gh release create v1.2.0-test --prerelease \
  --title "Test Release" \
  --notes "Testing npm package" \
  dist/*
```

**Option B: Test with existing release**

```bash
# Use an existing release version
cd npm
npm version 1.1.0 --no-git-tag-version
```

### 3. Test Package Creation

```bash
cd npm

# See what will be published
npm pack --dry-run

# Create actual tarball
npm pack

# Should create: bffgen-1.2.0.tgz (7-8 KB)
```

### 4. Test Local Installation

```bash
# Install from tarball
npm install -g ./bffgen-1.2.0.tgz

# Test version
bffgen version

# Test init command
bffgen init test-npm-install --lang nodejs-express

# Verify utilities created
ls test-npm-install/src/utils/
# Should show: aggregator.js cache-manager.js circuit-breaker.js ...

# Cleanup
npm uninstall -g bffgen
rm -rf test-npm-install
```

### 5. Test npx Usage

```bash
# Without global install
npx ./bffgen-1.2.0.tgz version

# Should work and show version
```

### 6. Test on Different Platforms

**macOS (Intel):**

```bash
arch -x86_64 npm install -g ./bffgen-1.2.0.tgz
bffgen version
```

**macOS (Apple Silicon):**

```bash
arch -arm64 npm install -g ./bffgen-1.2.0.tgz
bffgen version
```

**Linux (Docker):**

```bash
docker run --rm -v $(pwd):/work -w /work node:18 bash -c "
  npm install -g ./bffgen-1.2.0.tgz &&
  bffgen version
"
```

### 7. Test Install Script Directly

```bash
# Test platform detection
node scripts/platform.js

# Test install script (requires binaries on GitHub)
VERSION=1.2.0 node scripts/install.js
```

---

## Publishing Checklist

### Pre-Publish

- [ ] Binaries built: `make build-all VERSION=v1.2.0`
- [ ] Binaries uploaded to GitHub Releases
- [ ] npm package version matches binary version
- [ ] Package tested locally on at least 2 platforms
- [ ] npm pack shows correct files (no binaries included)
- [ ] README.md is clear and accurate
- [ ] LICENSE file included

### npm Account Setup

- [ ] npm account created: `npm adduser`
- [ ] Logged in: `npm whoami`
- [ ] 2FA enabled (recommended)
- [ ] Package name available: `npm search bffgen`

### First Publish

```bash
cd npm

# Dry run to see what would be published
npm publish --dry-run

# Publish (first time)
npm publish

# Or for scoped package
npm publish --access public
```

### Verify Publication

```bash
# Check on npm
npm view bffgen

# Install and test
npm install -g bffgen@1.2.0
bffgen version

# Test npx
npx bffgen@1.2.0 version

# Cleanup
npm uninstall -g bffgen
```

---

## CI/CD Testing

### GitHub Actions Testing

The workflow includes automatic testing:

```yaml
test-npm-package:
  needs: release
  strategy:
    matrix:
      os: [ubuntu-latest, macos-latest, windows-latest]
  runs-on: ${{ matrix.os }}
  steps:
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: "18"

    - name: Test npm install
      run: npm install -g bffgen@${{ steps.version.outputs.VERSION }}

    - name: Test command
      run: bffgen version

    - name: Test init
      run: bffgen init test-ci --lang nodejs-express --skip-tests
```

---

## Troubleshooting

### Binary Download Fails

**Check:**

1. GitHub Release exists for the version
2. Binaries uploaded to release
3. checksums.txt uploaded
4. Network connectivity

**Test download manually:**

```bash
curl -L https://github.com/RichGod93/bffgen/releases/download/v1.2.0/bffgen-darwin-arm64 -o test-binary
chmod +x test-binary
./test-binary version
```

### Checksum Mismatch

**Regenerate checksums:**

```bash
cd dist
sha256sum * > checksums.txt
```

### Platform Not Detected

**Check platform.js:**

```bash
node scripts/platform.js
# Should output platform info
```

**Supported:**

- darwin-x64 (macOS Intel)
- darwin-arm64 (macOS Apple Silicon)
- linux-x64 (Linux x86_64)
- linux-arm64 (Linux ARM64)
- win32-x64 (Windows x64)

---

## Version Management

### Keeping Versions in Sync

**Rule:** npm version = Go binary version (without 'v' prefix)

**Example:**

- Git tag: `v1.2.0`
- Go binary: version 1.2.0
- npm package: version 1.2.0

**Update npm version:**

```bash
cd npm
npm version 1.2.0 --no-git-tag-version
```

**Or use Makefile:**

```bash
make npm-package VERSION=v1.2.0
```

---

## Publishing Workflow

### Manual Publish

```bash
# 1. Build binaries
make build-all VERSION=v1.2.0

# 2. Create GitHub Release
gh release create v1.2.0 --title "Release v1.2.0" dist/*

# 3. Update npm package version
make npm-package VERSION=v1.2.0

# 4. Test locally
cd npm
npm pack
npm install -g ./bffgen-1.2.0.tgz
bffgen version

# 5. Publish to npm
npm publish

# 6. Verify
npm view bffgen
npm install -g bffgen@1.2.0
```

### Automated Publish (CI/CD)

```bash
# Just tag and push
git tag v1.2.0
git push origin v1.2.0

# GitHub Actions will:
# 1. Build binaries
# 2. Create GitHub Release
# 3. Publish to npm (if NPM_TOKEN secret is set)
```

---

## Success Criteria

### Installation Works

- ✅ `npm install -g bffgen` succeeds
- ✅ `npx bffgen` works without errors
- ✅ Binary downloaded to correct location
- ✅ Checksum verified
- ✅ Platform detected correctly

### Commands Work

- ✅ `bffgen version` shows correct version
- ✅ `bffgen init` creates project
- ✅ All utilities generated in Node.js projects
- ✅ Examples created

### Cross-Platform

- ✅ Works on macOS (Intel & Apple Silicon)
- ✅ Works on Linux (x64 & ARM64)
- ✅ Works on Windows (x64)

---

## Post-Publish Monitoring

### Check Package Health

```bash
# View package info
npm view bffgen

# Check download stats
npm info bffgen

# Test latest version
npm install -g bffgen
bffgen version
```

### Monitor Issues

- Watch GitHub Issues for installation problems
- Check npm package page for feedback
- Monitor download statistics

---

## Rollback Plan

### If Critical Issue Found

**Deprecate version:**

```bash
npm deprecate bffgen@1.2.0 "Critical issue found, please use v1.1.0"
```

**Publish patch:**

```bash
# Fix issue
# Bump to v1.2.1
make npm-package VERSION=v1.2.1
cd npm && npm publish
```

**Unpublish (within 72 hours only):**

```bash
npm unpublish bffgen@1.2.0
```

---

## Resources

- [npm Publishing Guide](https://docs.npmjs.com/packages-and-modules/contributing-packages-to-the-registry)
- [npm CLI Documentation](https://docs.npmjs.com/cli)
- [Semantic Versioning](https://semver.org)

---

**Testing Status:** Ready for manual testing  
**Publish Status:** Pending binaries on GitHub Releases
