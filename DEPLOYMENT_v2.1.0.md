# Deployment Guide - bffgen v2.1.0

## 📋 Pre-Deployment Checklist

✅ All tests pass (189 tests)
✅ Build successful for all platforms
✅ Version tagged: `v2.1.0`
✅ Commit created: `bb13b8b` and `2a1a09d`
✅ Release notes created: `RELEASE_NOTES_v2.1.0.md`
✅ npm package version updated to `2.1.0`
✅ Binaries built in `dist/` directory

## 🚀 Step 1: Push to GitHub

### Push commits to GitHub

```bash
cd /Users/richgodusen/Documents/MSc\ Programme/THESIS/bffgen

# Push the master branch
git push origin master

# Push the v2.1.0 tag
git push origin v2.1.0
```

**Expected output:**

```
Enumerating objects: X, done.
Counting objects: 100% (X/X), done.
Delta compression using up to N threads
Compressing objects: 100% (X/X), done.
Writing objects: 100% (X/X), X KiB | X MiB/s, done.
Total X (delta X), reused X (delta X), pack-reused 0
To https://github.com/RichGod93/bffgen.git
   d703a00..2a1a09d  master -> master
 * [new tag]         v2.1.0 -> v2.1.0
```

### Verify on GitHub

1. Visit: https://github.com/RichGod93/bffgen
2. Confirm commit `2a1a09d` is visible
3. Confirm tag `v2.1.0` appears in releases

## 📦 Step 2: Create GitHub Release

### Option A: Via GitHub Web UI

1. Go to: https://github.com/RichGod93/bffgen/releases/new
2. **Tag version**: Select `v2.1.0`
3. **Release title**: `v2.1.0 - Enhanced Testing & Code Decomposition`
4. **Description**: Copy from `RELEASE_NOTES_v2.1.0.md`
5. **Attach binaries**: Upload files from `dist/` directory:
   - `bffgen-darwin-amd64`
   - `bffgen-darwin-arm64`
   - `bffgen-linux-amd64`
   - `bffgen-linux-arm64`
   - `bffgen-windows-amd64.exe`
   - `checksums.txt`
6. Click **"Publish release"**

### Option B: Via GitHub CLI (if installed)

```bash
cd /Users/richgodusen/Documents/MSc\ Programme/THESIS/bffgen

# Create release with binaries
gh release create v2.1.0 \
  dist/bffgen-darwin-amd64 \
  dist/bffgen-darwin-arm64 \
  dist/bffgen-linux-amd64 \
  dist/bffgen-linux-arm64 \
  dist/bffgen-windows-amd64.exe \
  dist/checksums.txt \
  --title "v2.1.0 - Enhanced Testing & Code Decomposition" \
  --notes-file RELEASE_NOTES_v2.1.0.md
```

## 📦 Step 3: Publish to npm

### Prerequisites

Ensure you have:

- npm account with publish access to `bffgen` package
- Logged in via `npm login`

### Verify npm package

```bash
cd /Users/richgodusen/Documents/MSc\ Programme/THESIS/bffgen

# Check package.json version
cat npm/package.json | grep version

# Expected output:
# "version": "2.1.0",

# Test package locally
cd npm
npm pack

# This creates: bffgen-2.1.0.tgz
```

### Publish to npm

```bash
cd /Users/richgodusen/Documents/MSc\ Programme/THESIS/bffgen/npm

# Dry run (recommended first)
npm publish --dry-run

# Review the output, then publish for real
npm publish
```

**Expected output:**

```
npm notice
npm notice 📦  bffgen@2.1.0
npm notice === Tarball Contents ===
npm notice 1.1kB  LICENSE
npm notice 2.5kB  README.md
npm notice 647B   bin/bffgen.js
npm notice 1.2kB  lib/index.js
npm notice 3.1kB  package.json
npm notice 1.5kB  scripts/install.js
npm notice 892B   scripts/platform.js
npm notice === Tarball Details ===
npm notice name:          bffgen
npm notice version:       2.1.0
npm notice filename:      bffgen-2.1.0.tgz
npm notice package size:  X.X kB
npm notice unpacked size: X.X kB
npm notice shasum:        XXXXXXXXXXXXXXXXXXXXX
npm notice integrity:     XXXXXXXXXXXXXXXXXXXXX
npm notice total files:   7
npm notice
npm notice Publishing to https://registry.npmjs.org/
+ bffgen@2.1.0
```

### Verify npm publication

```bash
# Check on npm
npm view bffgen

# Should show version 2.1.0 as latest

# Test installation
npm install -g bffgen@2.1.0
bffgen --version
# Should output: bffgen version v2.1.0
```

## 🔗 Step 4: Update Documentation Links

### Update README badges (if applicable)

If your README has version badges, update them to point to v2.1.0:

```bash
# Edit README.md to update any version references
```

### Announce the release

Consider announcing on:

- GitHub Discussions
- Twitter/X
- LinkedIn
- Dev.to
- Project website

## ✅ Post-Deployment Verification

### GitHub Verification

- [ ] Commits pushed to master
- [ ] Tag v2.1.0 visible in tags
- [ ] Release v2.1.0 created with binaries
- [ ] Release notes displayed correctly

### npm Verification

- [ ] Package published: https://www.npmjs.com/package/bffgen
- [ ] Version 2.1.0 marked as latest
- [ ] Installation works: `npm install -g bffgen`
- [ ] Command works: `bffgen --version` outputs `v2.1.0`

### Download/Install Test

**From GitHub:**

```bash
# Download binary
curl -L -o bffgen https://github.com/RichGod93/bffgen/releases/download/v2.1.0/bffgen-darwin-amd64
chmod +x bffgen
./bffgen --version
```

**From npm:**

```bash
npx bffgen@2.1.0 --version
```

## 📊 Monitoring

After deployment, monitor:

1. **npm downloads**: https://www.npmjs.com/package/bffgen
2. **GitHub stars/forks**: https://github.com/RichGod93/bffgen
3. **Issues**: https://github.com/RichGod93/bffgen/issues
4. **Discussions**: https://github.com/RichGod93/bffgen/discussions

## 🐛 Rollback Plan (if needed)

If issues are discovered post-release:

### Rollback npm

```bash
npm deprecate bffgen@2.1.0 "Known issues, use v2.0.1 instead"
```

### Create hotfix

```bash
# Fix the issue
git checkout -b hotfix/v2.1.1
# Make fixes
git commit -m "fix: Critical issue in v2.1.0"
git push origin hotfix/v2.1.1

# Create v2.1.1 release
git tag -a v2.1.1 -m "Hotfix for v2.1.0"
git push origin v2.1.1
```

## 📝 Release Artifacts Summary

**Git:**

- Commits: `bb13b8b`, `2a1a09d`
- Tag: `v2.1.0`
- Branch: `master`

**Binaries (dist/):**

- `bffgen-darwin-amd64` (16MB)
- `bffgen-darwin-arm64` (15MB)
- `bffgen-linux-amd64` (16MB)
- `bffgen-linux-arm64` (15MB)
- `bffgen-windows-amd64.exe` (16MB)
- `checksums.txt` (SHA256)

**npm Package:**

- Name: `bffgen`
- Version: `2.1.0`
- Registry: https://www.npmjs.com/package/bffgen

**Documentation:**

- `RELEASE_NOTES_v2.1.0.md`
- `DEPLOYMENT_v2.1.0.md` (this file)

## 🎉 Success Criteria

Release is successful when:

✅ GitHub shows v2.1.0 release with all binaries
✅ npm shows bffgen@2.1.0 as latest version
✅ `npm install -g bffgen` installs v2.1.0
✅ `bffgen --version` outputs `v2.1.0`
✅ No critical issues reported in first 24 hours

---

**Next Release:** Plan for v2.2.0 with additional framework support and features
