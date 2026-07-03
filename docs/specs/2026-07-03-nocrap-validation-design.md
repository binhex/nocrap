# nocrap Validation Suite — Design Spec

**Date:** 2026-07-03
**Status:** Approved

## Problem

Python CRAP scores are validated (via radon subprocess and `.coverage` DB parsing, cross-checked against binhex/pytest-crap). JavaScript, TypeScript, and Go scores have **no independent validation** — there is no way to know whether the tree-sitter CC counters or the LCOV/cover.out parsers produce correct results.

## Goal

Add an automated validation suite that runs under `go test` and verifies:

1. **CC consistency across all 4 languages** — the same McCabe decision points produce the same CC regardless of language
2. **Language-specific CC correctness** — constructs unique to Go and JS/TS produce industry-standard CC values
3. **Coverage parsing correctness** — synthetic LCOV and cover.out data produce the expected coverage percentages

Zero runtime external tool dependencies. All tests self-contained.

## File Layout

```
nocrap/
├── validate/                              # NEW top-level validation package
│   ├── validate.go                        # shared helpers (float tolerance, etc.)
│   ├── cc_corpus/                         # cross-language equivalence
│   │   ├── corpus.go                      # test runner: loads fixtures, runs driver, asserts CC
│   │   ├── fixtures/
│   │   │   ├── equivalence.py             # 12 functions (Python, gold standard)
│   │   │   ├── equivalence.js             # same 12 functions, JS syntax
│   │   │   ├── equivalence.ts             # same 12 functions, TS syntax
│   │   │   └── equivalence.go             # same 12 functions, Go syntax
│   │   └── expected.json                  # pre-computed expected CC per function
│   ├── cc_ref/                            # language-specific reference checks
│   │   ├── ref_test.go                    # test runner: compares nocrap CC vs reference JSON
│   │   ├── fixtures/
│   │   │   ├── ref_js.ts                  # JS/TS constructs without Python equivalents
│   │   │   └── ref_go.go                  # Go constructs without Python equivalents
│   │   ├── expected_js.json               # pre-computed via ESLint (committed)
│   │   └── expected_go.json               # pre-computed via gocyclo (committed)
│   └── coverage/                          # synthetic coverage validation
│       ├── coverage_test.go               # test runner: synthetic LCOV + cover.out
│       ├── fixtures/
│       │   ├── lcov.info                  # hand-crafted: known covered/total lines
│       │   ├── cover.out                  # hand-crafted: known covered/total lines
│       │   ├── source.js                  # source for JS coverage test
│       │   └── source.go                  # source for Go coverage test
│       └── expected.json                  # hand-computed expected coverage percents
```

Existing `crossval/` remains unchanged (Python cross-validation against pytest-crap).

## Design

### 1. Cross-Language CC Corpus (`validate/cc_corpus/`)

Twelve functions covering all standard McCabe decision points. Each function is written identically (same name, same structure) in Python, JS, TS, and Go.

**Corpus:**

| # | Function | What it tests | Expected CC |
|---|----------|---------------|-------------|
| 1 | `no_branches` | Straight-line code, no conditions | 1 |
| 2 | `single_if` | One `if/else` | 2 |
| 3 | `if_else_if` | `if`/`else if`/`else` chain (3 branches) | 3 |
| 4 | `nested_if` | `if` inside `if` | 3 |
| 5 | `for_loop` | Simple `for` iteration | 2 |
| 6 | `for_with_if` | `for` containing `if` | 3 |
| 7 | `while_loop` | `while` with condition | 2 |
| 8 | `try_catch` | `try`/`catch`/`finally` | 3 |
| 9 | `boolean_ops` | `if (a && b \|\| c)` | 4 |
| 10 | `early_return` | Guard clauses (multiple returns) | 3 |
| 11 | `ternary` | `x ? a : b` / `a if x else b` | 2 |
| 12 | `switch_case` | Switch with 3 cases + default | 4 |

**Expected CC values** are validated against Python's radon output (already confirmed correct via pytest-crap cross-validation). The `expected.json` file stores `{"function_name": cc_value}` for reference.

**Test assertion:** For each function in each language, `nocrap.calcComplexity() == expected[function_name]`. Forty-eight assertions total (12 functions × 4 languages).

### 2. Language-Specific CC References (`validate/cc_ref/`)

For constructs that don't exist in Python and can't be tested via the cross-language corpus.

**Go-specific** (`ref_go.go`):
- `select` statement (channel select)
- `defer` statement
- `go` statement (goroutine launch)
- `type switch` (switch on type assertion)

**JS/TS-specific** (`ref_js.ts`):
- `switch` with `break`-less fallthrough cases
- `?.` optional chaining
- `??` nullish coalescing operator
- `for...in` and `for...of` loops

**Expected CC values** are generated once manually using external tools (`gocyclo` for Go, ESLint `complexity` rule for JS/TS) and committed as JSON files. The JSON files are the source of truth for tests — external tools are NOT invoked at test time.

Regeneration (manual, not part of CI): run `gocyclo` and `npx eslint` against the fixture files, parse their output into `{"function_name": cc}` JSON. Simple one-off Python scripts can do the parsing. The exact commands are documented in the implementation plan.

**Test assertion:** For each fixture function, `nocrap.calcComplexity() == expected[function_name]`. Tests do NOT invoke eslint or gocyclo — they compare against committed JSON.

### 3. Synthetic Coverage Validation (`validate/coverage/`)

Hand-craft source fixtures with known executable line counts, then create matching LCOV and cover.out files that cover exact subsets of those lines.

**JS/TS (LCOV format):**

| Fixture | Lines | Covered | Expected % |
|---------|-------|---------|------------|
| `source.js` + `lcov.info` (full) | 10 | 10 | 100% |
| `source.js` + `lcov.info` (half) | 10 | 5 | 50% |
| `source.js` + `lcov.info` (none) | 10 | 0 | 0% |

**Go (cover.out format):**

| Fixture | Lines | Covered | Expected % |
|---------|-------|---------|------------|
| `source.go` + `cover.out` (full) | 8 | 8 | 100% |
| `source.go` + `cover.out` (half) | 8 | 4 | 50% |
| `source.go` + `cover.out` (partial) | 8 | 1 | 12.5% |

**Expected coverage values** are hand-computed (lines covered / total executable lines × 100) and stored in `expected.json`.

**Test assertion:** For each fixture, `nocrap.coverage_percent == expected_percent ± 0.5`.

TypeScript reuses the same LCOV parser as JavaScript, so no separate TS coverage fixture is needed.

### 4. Test Integration

All tests run under `go test ./validate/...` with zero external dependencies. The `crossval/` package continues to handle Python cross-validation against pytest-crap (unchanged).

```
go test ./...                          # everything, including validation
go test ./validate/cc_corpus/          # cross-language CC only
go test ./validate/cc_ref/             # language-specific CC only
go test ./validate/coverage/           # coverage parsing only
```

## Future Maintenance

- **Tree-sitter grammar updates:** if CC values change, corpus and reference tests will catch the regression. Investigate whether the change is a bug fix or a regression.
- **New language constructs:** add a function to the corpus (if cross-language) or the appropriate ref fixture (if language-specific), update expected JSONs.
- **Regenerating reference JSONs:** use the `validate/cc_ref/Makefile` (manual, not CI). After regenerating, commit the updated JSONs.
