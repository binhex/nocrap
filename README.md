# nocrap

Calculate CRAP (Change Risk Anti-Patterns) scores for Python, JavaScript,
TypeScript, and Go source code. A single static Go binary that works with
pre-generated coverage data.

## Quick Start

```bash
# Install
go install ./...

# Run tests with coverage (any language)
pytest --cov --cov-report=json   # Python
npm test -- --coverage           # JavaScript/TypeScript
go test -coverprofile=cover.out ./...  # Go

# Analyze
nocrap ./
```

## Supported Languages

| Language   | Coverage Format     | Source                           |
|------------|---------------------|----------------------------------|
| Python     | `.coverage.json`    | `python -m coverage json`        |
| JavaScript | `lcov.info`         | Istanbul, nyc, c8, Jest          |
| TypeScript | `lcov.info`         | Istanbul, nyc, c8, Jest          |
| Go         | `cover.out`         | `go test -coverprofile=cover.out`|

## Configuration

See `.crap.toml` in the project root:

```toml
threshold = 9
top_n = 20
exclude = ["**/test_*", "**/vendor/**", "**/node_modules/**"]

[coverage]
python = ".coverage.json"
javascript = "coverage/lcov.info"
go = "cover.out"
```

## CLI Flags

| Flag          | Description                               | Default        |
|---------------|-------------------------------------------|----------------|
| `--lang`      | Force language                            | auto-detect    |
| `--threshold` | CRAP threshold for highlighting           | 30             |
| `--top-n`     | Items per table (0 = all)                 | 20             |
| `--json`      | Output machine-readable JSON              | false          |
| `--config`    | Path to config file                       | `.crap.toml`   |
| `--exclude`   | Glob patterns to exclude (repeatable)     |                |

## Environment Variables

- `CRAP_COVERAGE_PYTHON` — Override Python coverage file path
- `CRAP_COVERAGE_JAVASCRIPT` — Override JS/TS coverage file path
- `CRAP_COVERAGE_GO` — Override Go coverage file path

## Color Coding

| CRAP Score | Color  | Meaning           |
|------------|--------|-------------------|
| ≤ 15       | Green  | Low risk          |
| 16-30      | Yellow | Moderate risk     |
| > 30       | Red    | High risk — refactor |

## Development

```bash
make test       # Run all tests
make build      # Build binary
make dogfood    # Self-analysis
```

## License

MIT
