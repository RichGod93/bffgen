#!/usr/bin/env bash
# Start (or prepare) a release branch.
#
#   scripts/release.sh 3.6.0
#
# Creates release/v3.6.0 from origin/master (or reuses it if it is the current branch) and sets
# the version in every place it lives: npm/package.json, internal/registry/versions.yaml and a
# CHANGELOG.md section (an existing "## [Unreleased]" section is renamed). Commit your work on
# that branch and open a PR to master; merging it publishes the release (see docs/RELEASING.md).
set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

version="${1:-}"
version="${version#v}"
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "usage: $0 X.Y.Z"
branch="release/v$version"

cd "$(git rev-parse --show-toplevel)"
git fetch origin --tags --quiet
git rev-parse --quiet --verify "refs/tags/v$version" >/dev/null && die "tag v$version already exists"

latest="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [ -n "$latest" ] && [ "$(printf '%s\n%s\n' "${latest#v}" "$version" | sort -V | tail -n 1)" != "$version" ]; then
  die "v$version is not newer than the latest release $latest"
fi

if [ "$(git branch --show-current)" != "$branch" ]; then
  if ! git diff --quiet || ! git diff --cached --quiet; then die "working tree has uncommitted changes"; fi
  git switch --create "$branch" origin/master
fi

today="$(date -u +%Y-%m-%d)"

# npm/package.json
node -e '
  const fs = require("fs"); const f = "npm/package.json";
  const pkg = JSON.parse(fs.readFileSync(f, "utf8")); pkg.version = process.argv[1];
  fs.writeFileSync(f, JSON.stringify(pkg, null, 2) + "\n");
' "$version"

# internal/registry/versions.yaml metadata
perl -0pi -e "s/(metadata:\n  version: )\"[^\"]*\"/\${1}\"$version\"/; s/(  last_updated: )\"[^\"]*\"/\${1}\"${today}T00:00:00Z\"/; s/(  bffgen_version: )\"[^\"]*\"/\${1}\"$version\"/" internal/registry/versions.yaml

# CHANGELOG.md section
if grep -q "^## \[$version\]" CHANGELOG.md; then
  :
elif grep -q '^## \[Unreleased\]' CHANGELOG.md; then
  perl -pi -e "s/^## \[Unreleased\].*/## [$version] - $today/" CHANGELOG.md
else
  perl -0pi -e "s/\n## \[/\n## [$version] - $today\n\n### Changed\n- TODO: describe the user-facing change\n\n## [/" CHANGELOG.md
fi

echo "On $branch, version set to $version. Next:"
echo "  1. edit the CHANGELOG.md section for $version"
echo "  2. scripts/check-release.sh        # what CI checks on the PR"
echo "  3. git push -u origin $branch && gh pr create --base master"
