# nocrap — Design Spec

**Date:** 2026-07-03
**Status:** Approved

## Problem

The existing `pytest-crap` tool calculates CRAP scores (Change Risk Analysis and Predictions)
only for Python. The CRAP formula `CC^2 * (1 - cov)^3 + CC` is language-agnostic — any language
with branching constructs (`if`, `while`, `for`, etc.) and coverage tooling can be scored. This
project creates a single static Go binary called `nocrap` that calculates CRAP scores for Python,
JavaScript, TypeScript, and Go source code.

## Architecture

```
nocrap                   Single Go binary
├── CLI layer            Cobra: flags, paths, language detection
├── Engine               Orchestrator: detects language, routes to driver, applies CRAP
├── Drivers              One Go package per language
│   ├── python/           tree-sitter-python + coverage.py JSON parser
│   ├── javascript/       tree-sitter-javascript + LCOV parser
│   ├── typescript/       tree-sitter-typescript + LCOV parser
│   └── go/               tree-sitter-go + go cover profile parser
├── Calculator            Shared CRAP formula (language-agnostic)
├── Coverage              Parsers for coverage.py JSON, LCOV, and go cover formats
├── Reporter              Rich terminal tables matching current pytest-crap look
└── Tree-sitter           Embedded grammars, CST walker for CC calculation
```

The tool never runs tests or collects coverage — the user runs their normal test+coverage
workflow first, then points `nocrap` at the coverage output. The tool is purely analytical.

## Driver Interface

Every language driver implements:

```go
type Driver interface {
    Name() string
    Extensions() []string
    FindFunctions(source []byte) ([]Function, error)
    CalcComplexity(source []byte, fn Function) (int, error)
}

type Function struct {
    Name      string
    File      string
    StartLine int
    EndLine   int
    Package   string  // class, module, namespace, or "" for top-level
}
```

- **FindFunctions** — parses source with tree-sitter, walks the CST to find all
  function/method nodes (including nested, class methods, arrow functions, async, generators).
  Returns name and line range. Docstrings and decorators are excluded from the executable
  range (matching the existing pytest-crap fork behavior).
- **CalcComplexity** — walks the CST for a specific function's subtree, increments CC by 1
  for each branching construct in that language.

### Branching Constructs by Language

| Language | +1 CC for |
|----------|-----------|
| Python | `if`, `elif`, `while`, `for`, `except`, `with`, `match/case`, `and`/`or` in conditions |
| JavaScript | `if`, `else if`, `while`, `for`, `do/while`, `for...in`, `for...of`, `try/catch`, `switch/case`, `&&`/`\|\|`/`??`/`?.` in conditions, ternary `?:` |
| TypeScript | Same as JavaScript |
| Go | `if`, `else`, `for`, `range`, `switch/case`, `select/case`, `&&`/`\|\|` in conditions, `type switch` |

## Coverage Parsing

Coverage parsing is separate from the driver interface. A shared coverage module reads
each language's standard coverage format and produces `map[string]CoverageData` keyed
by absolute or project-relative filename.

| Language | Format | Source |
|----------|--------|--------|
| Python | `.coverage.json` | `python -m coverage json` |
| JavaScript | `lcov.info` | Istanbul, nyc, c8 |
| TypeScript | `lcov.info` | Istanbul, nyc, c8 |
| Go | `cover.out` | `go test -coverprofile=cover.out` |

The coverage module resolves filenames between the coverage data and the source tree.
For LCOV, path resolution follows the SF (source file) field relative to the project root.

## CRAP Calculator

Language-agnostic. Inputs: cyclomatic complexity (int), coverage percentage (float 0-100),
executable line count, covered line count. Output: CRAP score (float).

```
crap = cc^2 * (1 - coverage_percent / 100)^3 + cc
```

Implements the same executable-line filtering as the existing pytest-crap fork:
blank lines and comment-only lines are excluded from the total when computing
coverage percentage. Docstrings are excluded from the function range by the driver's
FindFunctions.

## Reporter

Rich terminal tables matching the current pytest-crap output:

1. **CRAP by Function** — individual functions ranked by CRAP score, with CC, coverage %,
   function name, and file path. Color-coded: green (<=15), yellow (15 < x <=30), red (>30).
2. **CRAP by File** — files ranked by max CRAP score, with count of functions at or above
   threshold.
3. **CRAP by Folder** — directories ranked by max CRAP score.

Long file paths are truncated in the middle (e.g., `very/long/.../file.py`). Terminal
width is detected and used for column sizing.

If no `--threshold` is passed, the default is 30 (matching pytest-crap upstream), but
the standard tech-debt workflow uses threshold 9.

## CLI

```
nocrap [flags] <path...>>

Flags:
  --lang <name>        Force a language (python, javascript, typescript, go)
                       If omitted, auto-detect from file extensions.
  --threshold <n>      CRAP threshold for highlighting (default: 30)
  --top-n <n>          Number of items per table. 0 = show all (default: 20)
  --json               Output machine-readable JSON instead of tables
  --config <path>      Path to config file (default: .crap.toml)
  --exclude <pattern>  Glob patterns to exclude (repeatable, appends to config)

Arguments:
  <path...>            One or more directories or files to analyze.
                       If a directory, walks recursively.
```

**Coverage discovery:** The tool looks for coverage data in standard locations:
- `.coverage.json` (Python, in project root)
- `coverage/lcov.info` (JS/TS, common default)
- `cover.out` (Go, in project root)

Each can be overridden with `CRAP_COVERAGE_PYTHON`, `CRAP_COVERAGE_JAVASCRIPT`,
`CRAP_COVERAGE_GO` environment variables, or a `[coverage]` section in `.crap.toml`.

## Configuration

Optional `.crap.toml` at the project root:

```toml
threshold = 9
top_n = 20
exclude = ["**/test_*", "**/*_test.go", "**/vendor/**", "**/node_modules/**"]

[coverage]
python = ".coverage.json"
javascript = "coverage/lcov.info"
go = "cover.out"
```

CLI flags override config file values. Environment variables override both.

## Error Handling

The tool never crashes on bad input. All errors are handled gracefully:

| Scenario | Behavior |
|----------|----------|
| Unparseable source file | Skip file, warn to stderr, continue |
| Missing coverage data | Skip file, note "no coverage data" to stderr |
| Unknown file extension | Skip silently |
| Coverage parse error | Warn to stderr, treat as 0% coverage |
| No functions found in a file | Skip file silently |

A summary line appears at the end:
```
⚠ 3 files skipped (2 parse errors, 1 missing coverage)
```

In JSON mode, skipped files appear in the output with an `"error"` field.

## Cross-Validation with pytest-crap

**The Go tool MUST produce identical CRAP scores to the existing pytest-crap Python
module for all Python source files.** This is a non-negotiable correctness requirement.

### Validation approach

1. **Generate a comprehensive test corpus** — a set of Python files covering every
   branching construct (`if`, `elif`, `while`, `for`, `except`, `with`, `match/case`,
   `and`/`or`), decorators, nested functions, class methods, async functions, docstrings,
   blank lines, comments, and edge cases (empty body, abstract stubs, single-line
   functions).

2. **Run both tools on the same corpus with the same coverage data:**
   ```bash
   pytest --cov=corpus --cov-report=json --crap --crap-top-n=0
   nocrap --lang python --top-n 0 corpus/
   ```

3. **Diff the output** — every function must have the same CC, same coverage %, and
   same CRAP score (within floating-point tolerance of 0.01). Any discrepancy means
   the Go driver is wrong and must be fixed before the tool is usable.

4. **Automate as a CI gate** — a test suite that runs both tools and asserts
   score equality. This gate runs on every commit. If a tree-sitter grammar update
   or CC walker change introduces a score divergence, the gate catches it immediately.

### Expected sources of divergence (to watch for)

| Area | Risk |
|------|------|
| CC counting | tree-sitter CST may expose different node types than Python's ast module. `elif` vs `else if`, implicit `else` blocks, comprehension scoping. |
| Executable line filtering | Comment detection (`#` vs inline comments). Blank line handling. Docstring boundary detection. |
| Coverage line mapping | coverage.py reports physical lines; the Go tool must map them identically to the function range. |
| Function boundary detection | Decorators spanning multiple lines. Functions with only a docstring body. Nested functions in class methods. |

## Testing Strategy

| Layer | Tests |
|-------|-------|
| Calculator | Unit tests: known CC + coverage -> expected CRAP. Edge cases: 0% cov, 100% cov, CC=1, CC=20. |
| Tree-sitter walker | Snapshot tests per language: given source, assert correct CC per function. Covers all branching constructs. |
| Coverage parsers | Parse real sample `.coverage.json`, `lcov.info`, `cover.out`. Assert line sets per file. |
| Drivers | Integration per language: real multi-function source file, verify all functions found with correct CC. |
| Reporter | Table output matches expected snapshot. Green/yellow/red color thresholds. Truncation. |
| CLI | Golden file tests: run `nocrap` on fixture projects, capture stdout + stderr, compare to expected. |

Fixture projects live in `testdata/` — each is a tiny project with known complexity and
coverage, with pre-generated coverage data committed alongside.

## Dependencies (Go Modules)

```
github.com/smacker/go-tree-sitter    tree-sitter Go bindings
github.com/spf13/cobra               CLI framework
github.com/pelletier/go-toml/v2      TOML config parsing
golang.org/x/term                    Terminal width detection
```

Table rendering is implemented directly (no external dependency) using ANSI escape codes
and term width detection, keeping the dependency footprint minimal.

## Dogfooding

After the initial build, `nocrap` analyzes its own source code:

```
go test -coverprofile=cover.out ./...
nocrap --lang go --threshold 9 ./
```

This ensures the tool's own codebase meets the same CRAP standards it enforces.

## Open Questions / Future Work

- **Additional languages:** Rust, Java, C#, Ruby, PHP — adding a language means adding one
  tree-sitter grammar + one coverage format parser + one driver package (~200-400 lines each).
  Each new language should be validated against a trusted reference tool for that language
  (e.g., validate Go driver against `gocyclo` complexity scores).
- **CI integration:** A GitHub Action or GitLab CI template that runs nocrap and fails the
  build if any function exceeds the threshold.
- **HTML report:** A `nocrap --html` flag that generates a browsable report with collapsible
  sections per function.
- **Diff mode:** `nocrap --diff main` — only show functions whose CRAP score changed vs a
  baseline branch.
