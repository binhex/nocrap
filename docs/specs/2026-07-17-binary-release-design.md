# Binary Release CI Design

**Date:** 2026-07-17
**Status:** Approved

## Summary

Add a GitHub Actions release workflow that builds `nocrap` binaries (linux/amd64,
linux/arm64) and attaches them to GitHub Releases when a semver tag is pushed.

## Trigger

Push a semver tag (`v*`) to `main` triggers the release workflow.

Example:
```bash
git tag v0.4.0
git push origin v0.4.0
```

## Flow

```
git tag v0.4.0 && git push origin v0.4.0
        |
        v
[CI: release.yml]
  +-- test job -----------------------+
  | go test ./... -race               |
  | go build (sanity check)           |
  +---------|-------------------------+
            | (pass)
            v
  +-- goreleaser job -----------------+
  | cross-compile:                    |
  |   linux/amd64                     |
  |   linux/arm64                     |
  | generate checksums (SHA256)       |
  | generate changelog from commits   |
  | create GitHub Release (via API)   |
  | upload binaries + checksums       |
  +-----------------------------------+
```

Tests gate the release — if they fail, nothing is published.

## New Files

### `.goreleaser.yml` (repo root)

```yaml
before:
  hooks:
    - go mod tidy

builds:
  - id: nocrap
    binary: nocrap
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath
```

Key details:
- `CGO_ENABLED=0`: pure static binary, no libc dependency
- `-s -w`: strip debug info and symbol table for smaller binaries
- `-trimpath`: remove build machine filesystem paths from the binary

### `.github/workflows/release.yml`

Runs on `push: tags: ["v*"]`. Two jobs:

1. **test** — runs `go test ./... -race` and a build sanity check. Gates the
   release.
2. **goreleaser** — `needs: test`, sets up Go 1.22, calls
   `goreleaser/goreleaser-action@v6` with the `GITHUB_TOKEN` secret.

## Existing Files

- `.github/workflows/test.yml` — unchanged. Continues to run on every push and
  PR to `main`.

## Artifacts per Release

| Artifact | Description |
|---|---|
| `nocrap_linux_amd64` | x86_64 static binary |
| `nocrap_linux_arm64` | ARM64 static binary |
| `checksums.txt` | SHA256 hashes of both binaries |
| Changelog | Auto-generated from commit messages since previous tag |

## Versioning

- Existing tag pattern (`v0.3.3`) continues unchanged.
- GoReleaser uses the tag as the release name and version.
- Tags matching `v*.*.*-*` (e.g., `v0.4.0-rc1`) are marked as pre-releases
  automatically.

## Edge Cases

- **Tag push but tests fail:** No release is created. Delete the tag
  (`git tag -d v0.4.0 && git push origin :v0.4.0`), fix the issue, re-tag.
- **Re-tagging an existing release:** GoReleaser will not overwrite by default.
  Delete the existing GitHub Release first, then re-tag.
- **First release after setup:** Changelog covers all commits since the initial
  commit. Subsequent releases only cover commits since the previous tag.
