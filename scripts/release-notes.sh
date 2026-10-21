#!/usr/bin/env bash
# Print the GitHub release notes for a version: its CHANGELOG.md section (relative links made
# absolute, pinned to the tag) followed by install instructions.
#
#   scripts/release-notes.sh 3.6.0 [owner/repo]
set -euo pipefail

version="${1#v}"
repo="${2:-${GITHUB_REPOSITORY:-RichGod93/bffgen}}"
cd "$(git rev-parse --show-toplevel)"

section="$(awk -v v="$version" '
  index($0, "## [" v "]") == 1 { found = 1; next }
  found && /^## \[/ { exit }
  found { print }
' CHANGELOG.md)"
[ -n "$(printf '%s' "$section" | tr -d '[:space:]')" ] || { echo "no CHANGELOG.md section for $version" >&2; exit 1; }

printf '%s\n' "$section" |
  perl -pe "s{\]\((?!https?://|mailto:|#)([^)\s]+)\)}{](https://github.com/$repo/blob/v$version/\$1)}g" |
  sed -e '/./,$!d'

cat <<EOF

---

### Install

\`\`\`bash
npm install -g bffgen@$version        # or: npx bffgen@$version init my-bff
go install github.com/$repo/v3/cmd/bffgen@v$version
\`\`\`

Prebuilt binaries for Linux, macOS and Windows (amd64/arm64) are attached below; verify them with
\`checksums.txt\` (\`sha256sum -c checksums.txt --ignore-missing\`).
EOF
