#!/usr/bin/env bash
# Check that a branch is a valid release branch: named release/vX.Y.Z, with X.Y.Z set in every
# version location, a CHANGELOG section, no existing tag, and newer than the latest release.
# Runs in CI on every pull request to master; run it locally before opening one.
#
#   scripts/check-release.sh [branch]      (default: $GITHUB_HEAD_REF, else the current branch)
set -euo pipefail

branch="${1:-${GITHUB_HEAD_REF:-$(git branch --show-current)}}"
cd "$(git rev-parse --show-toplevel)"
fail=0
err() { echo "::error::$*"; fail=1; }

if [[ ! "$branch" =~ ^release/v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
  echo "::error::Branch '$branch' does not follow the release naming convention."
  echo "Pull requests to master must come from a branch named release/vX.Y.Z, the version that"
  echo "merging it will publish. Start one with: scripts/release.sh X.Y.Z (see docs/RELEASING.md)."
  echo "Contributors: open your PR against the current release/v* branch instead of master."
  exit 1
fi
version="${BASH_REMATCH[1]}"

pkg_version="$(node -p 'require("./npm/package.json").version')"
[ "$pkg_version" = "$version" ] || err "npm/package.json version is $pkg_version, expected $version"

for key in version bffgen_version; do
  got="$(sed -n "s/^  $key: \"\(.*\)\"/\1/p" internal/registry/versions.yaml | head -n 1)"
  [ "$got" = "$version" ] || err "internal/registry/versions.yaml metadata.$key is '$got', expected $version"
done

if ! grep -Eq "^## \[$version\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$" CHANGELOG.md; then
  err "CHANGELOG.md has no '## [$version] - YYYY-MM-DD' section"
elif grep -q "TODO: describe the user-facing change" CHANGELOG.md; then
  err "CHANGELOG.md still contains the TODO placeholder from scripts/release.sh"
fi

git fetch origin --tags --quiet 2>/dev/null || true
if git rev-parse --quiet --verify "refs/tags/v$version" >/dev/null; then
  err "tag v$version already exists; pick the next version"
fi
latest="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | grep -vx "v$version" | head -n 1 || true)"
if [ -n "$latest" ] && [ "$(printf '%s\n%s\n' "${latest#v}" "$version" | sort -V | tail -n 1)" != "$version" ]; then
  err "v$version is not newer than the latest release $latest"
fi

[ "$fail" -eq 0 ] && echo "release/v$version is ready to release (latest: ${latest:-none})"
exit "$fail"
