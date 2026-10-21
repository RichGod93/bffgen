# Releasing bffgen

Branches are named after the release they will become, and merging that branch is the release.

```
release/v3.6.0  ──PR──▶  master  ──▶  tag v3.6.0 + GitHub release (binaries) + npm bffgen@3.6.0
```

## Branch naming

| Branch | Purpose |
|---|---|
| `master` | Released code only. Every merge into it is a release. |
| `release/vX.Y.Z` | All work for the next release, `X.Y.Z` being the version it will publish. One branch per release; delete it after merging (the merge does this). |
| `dependabot/*` | Automated dependency PRs. Exempt from the naming check; they ship with the next release. |

Pick the version with [Semantic Versioning](https://semver.org/): a **patch** (`3.5.1`) for fixes,
including dependency and security updates in generated projects; a **minor** (`3.6.0`) for new
features; a **major** (`4.0.0`) only for breaking changes to `bffgen.config.json` or the CLI.

Contributors: open your pull request against the current `release/v*` branch. If none exists,
open it against `master` and a maintainer will retarget it.

## Cutting a release

```bash
scripts/release.sh 3.6.0          # creates release/v3.6.0 from origin/master and sets the version
# ...commit the work, write the CHANGELOG.md section...
scripts/check-release.sh          # the same check CI runs on the PR
git push -u origin release/v3.6.0
gh pr create --base master --title "Release v3.6.0"
```

`scripts/release.sh` sets the version in every place it lives:

- `npm/package.json` (`version`)
- `internal/registry/versions.yaml` (`metadata.version`, `metadata.bffgen_version`, `metadata.last_updated`)
- `CHANGELOG.md` (renames `## [Unreleased]`, or adds a `## [X.Y.Z] - date` section to fill in)

CI (`.github/workflows/ci.yml`) then checks the pull request:

- **release-branch**: the branch is `release/vX.Y.Z`; the version is the same in all three files;
  the CHANGELOG section exists; tag `vX.Y.Z` does not exist yet; `X.Y.Z` is newer than the latest release.
- **test**: Go tests (race detector) on the minimum and newest Go, gofmt, golangci-lint, gosec, govulncheck.
- **generated-projects**: for Express, Fastify and FastAPI, `test/e2e/run.sh` generates a project, lints it,
  runs its own tests, audits its dependencies (`npm audit --omit=dev --audit-level=high` / `pip-audit`)
  and runs the HTTP and fault-tolerance checks in `test/e2e/checks.js` against mock backends.

## What merging does

`.github/workflows/release.yml` runs on every push to `master`. It reads the version from
`npm/package.json` and does whatever is still missing for it, so a push without a version bump is a no-op:

1. **github-release**: tests, builds six binaries (linux/darwin/windows × amd64/arm64) with the newest
   stable Go, writes `checksums.txt`, tags `vX.Y.Z` (annotated) at the merge commit, creates a draft
   release with the notes from `scripts/release-notes.sh` (the CHANGELOG section plus install
   instructions), uploads the assets, then publishes the release.
2. **npm-publish** (if trusted publishing is enabled, see below): `npm publish --provenance` from
   `npm/`. This runs only after the GitHub release is public, because the npm package's postinstall
   downloads the binary from it. Otherwise **npm-manual** puts the publish commands in the run summary.
3. **smoke**: on Linux, macOS and Windows, `npx bffgen@X.Y.Z --version` and
   `go install github.com/RichGod93/bffgen/v3/cmd/bffgen@vX.Y.Z` must report `vX.Y.Z`. After a manual
   npm publish, re-run the workflow (*Actions → Release → Run workflow*) to run it.

If a run fails part-way, fix the cause and re-run it (or run the workflow by hand). Finished steps
are skipped. Tags and releases are never deleted, moved or recreated; a broken release is fixed by
releasing the next patch version.

## One-time setup

- **npm trusted publishing** (optional; until it's set up, publish by hand): on npmjs.com, open the
  `bffgen` package → *Settings* → *Trusted publishing* → GitHub Actions, repository
  `RichGod93/bffgen`, workflow `release.yml`. Then set the repository variable
  `NPM_TRUSTED_PUBLISHING` to `true` (*Settings → Secrets and variables → Actions → Variables*).
  The workflow never uses a long-lived npm token. While the variable is unset, the run summary
  shows the manual publish commands.
- **GitHub**: enable *Settings → General → Automatically delete head branches*, so merged release
  branches disappear.

## Publishing to npm by hand

Until trusted publishing is enabled (or if the npm job fails), the GitHub release is still complete.
Publish from a checkout of the tag:

```bash
git switch --detach vX.Y.Z
cd npm && npm publish --access public    # prompts for 2FA in the browser
npm view bffgen dist-tags.latest          # may take a minute or two to update
```
