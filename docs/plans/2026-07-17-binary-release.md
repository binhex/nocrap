# Binary Release CI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use sub-agents (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a GitHub Actions release workflow that builds nocrap binaries
(linux/amd64, linux/arm64) via GoReleaser and attaches them to GitHub Releases
when a semver tag is pushed.

**Architecture:** Two new files — `.goreleaser.yml` configures GoReleaser to
cross-compile a static CGO-disabled binary for two Linux architectures.
`.github/workflows/release.yml` defines a two-job workflow: the `test` job
runs the test suite as a gate, and the `goreleaser` job (dependent on test
passing) invokes the goreleaser-action to build, checksum, changelog, and
publish. The existing `test.yml` is untouched.

**Tech Stack:** GoReleaser (`goreleaser/goreleaser-action@v6`), GitHub Actions,
Go 1.22.

---

### Task 1: Create `.goreleaser.yml`

**Files:**
- Create: `.goreleaser.yml`

- [ ] **Step 1: Write `.goreleaser.yml`**

```yaml
# nocrap GoReleaser config — builds static linux/amd64 and linux/arm64 binaries
# on tag push (v*), triggered by .github/workflows/release.yml

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

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.goreleaser.yml'))" && echo "OK"
```

- [ ] **Step 3: Dry-run GoReleaser (check config parsing)**

```bash
# Skip if goreleaser binary is not available — not a blocker
goreleaser check 2>&1 || echo "(goreleaser not installed locally — expected, config is valid YAML)"
```

- [ ] **Step 4: Commit**

```bash
git add .goreleaser.yml
git commit -m "feat: add GoReleaser config for linux amd64/arm64 binaries"
```

---

### Task 2: Create `.github/workflows/release.yml`

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install dependencies
        run: sudo apt-get install -y gcc
      - name: Test
        run: go test ./... -v -count=1 -race

  goreleaser:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Validate YAML syntax**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "OK"
```

- [ ] **Step 3: Verify GoReleaser config is discoverable**

```bash
test -f .goreleaser.yml && echo "Config present" || echo "MISSING"
test -f .github/workflows/release.yml && echo "Workflow present" || echo "MISSING"
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "feat: add release workflow — GoReleaser on tag push"
```

---

### Task 3: Final validation

**Files:** None — read-only check.

- [ ] **Step 1: Confirm both files exist and are valid YAML**

```bash
cd /data/nocrap
echo "=== .goreleaser.yml ===" && python3 -c "import yaml; yaml.safe_load(open('.goreleaser.yml'))" && echo "VALID"
echo "=== .github/workflows/release.yml ===" && python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "VALID"
```

- [ ] **Step 2: Confirm existing test.yml is untouched**

```bash
git diff HEAD~2..HEAD -- .github/workflows/test.yml || echo "(no changes to test.yml — good)"
```

- [ ] **Step 3: Dry-run the full test suite one last time**

```bash
go test ./... -count=1 -race
```
Expected: all tests pass.

---

## Post-Implementation Verification

After both commits are pushed to `main`, test the release flow:

```bash
# 1. Create a test tag (use a pre-release pattern to keep things clean)
git tag v0.4.0-rc1
git push origin v0.4.0-rc1

# 2. Watch the Release workflow in GitHub Actions UI
# Expected: test job passes → goreleaser job creates release + uploads binaries

# 3. Verify the release at https://github.com/<owner>/nocrap/releases
# Expected: nocrap_linux_amd64, nocrap_linux_arm64, checksums.txt, changelog

# 4. If something goes wrong, delete the tag and the draft release:
git tag -d v0.4.0-rc1
git push origin :v0.4.0-rc1
# Delete the release via GitHub UI
```
