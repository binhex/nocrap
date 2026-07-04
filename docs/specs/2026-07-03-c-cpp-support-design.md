# nocrap C/C++ Language Support — Design Spec

**Date:** 2026-07-03
**Status:** Approved

## Problem

nocrap currently supports Python, JavaScript, TypeScript, and Go. A real-world firmware project (CrossInk, an ESP32-based e-ink reader firmware with 102 `.cpp` and 142 `.h` files) cannot be analyzed — C and C++ language drivers and coverage parsing do not exist.

## Goal

Add full C and C++ language support to nocrap, matching the rigour of the existing 4-language implementations: tree-sitter AST parsing for function discovery and cyclomatic complexity, gcov coverage parsing for CRAP scores, and automated validation (cross-language CC corpus, language-specific ref tests, synthetic coverage tests).

## File Layout

```
nocrap/
├── internal/
│   ├── config/config.go              # MODIFY: add C, Cpp fields to CoverageConfig
│   ├── driver/
│   │   ├── c/                         # NEW
│   │   │   ├── c_driver.go           # C language driver (tree-sitter-c grammar)
│   │   │   └── c_driver_test.go      # basic driver tests
│   │   └── cpp/                       # NEW
│   │       ├── cpp_driver.go         # C++ language driver (tree-sitter-cpp grammar)
│   │       └── cpp_driver_test.go
│   ├── coverage/
│   │   ├── gcov.go                   # NEW: ParseGcov(path) → CoverageMap
│   │   └── gcov_test.go             # NEW
│   └── engine/
│       └── engine.go                 # MODIFY: add drivers to registry, detect .c/.cpp/.h
├── validate/
│   ├── cc_corpus/
│   │   ├── expected.json             # MODIFY: add skip_c array
│   │   └── fixtures/
│   │       ├── equivalence.c         # NEW: 12 functions, C syntax
│   │       └── equivalence.cpp       # NEW: 12 functions, C++ syntax
│   ├── cc_ref/
│   │   ├── expected_cpp.json         # NEW: pre-computed via lizard (committed)
│   │   └── fixtures/
│   │       └── ref_cpp.cpp           # NEW: C++-only constructs
│   └── coverage/
│       ├── expected.json             # MODIFY: add gcov entries
│       └── fixtures/
│           ├── add.c                 # NEW: 2 single-line functions (add, sub)
│           ├── full.gcov             # NEW: 2/2 lines covered (100%)
│           ├── half.gcov             # NEW: 1/2 lines covered (50%)
│           └── none.gcov             # NEW: 0/2 lines covered (0%)
```

## Design

### 1. C/C++ Language Drivers

Two drivers share the same CC counter logic but use different tree-sitter grammars:

| Property | C driver | C++ driver |
|----------|----------|------------|
| Grammar | `github.com/smacker/go-tree-sitter/c` | `github.com/smacker/go-tree-sitter/cpp` |
| Extensions | `.c`, `.h` | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hh` |
| Language ID | `"c"` | `"cpp"` |
| Package | `nocrap/internal/driver/c` | `nocrap/internal/driver/cpp` |

**CC constructs counted (shared function between both drivers):**

| Tree-sitter node type | CC increment | Notes |
|------------------------|-------------|-------|
| `if_statement` | +1 | |
| `for_statement` | +1 | Includes C range-for and standard for |
| `while_statement` | +1 | |
| `do_statement` | +1 | |
| `case_statement` | +1 | Per case label (switch count = number of cases) |
| `catch_clause` | +1 | Per catch block (C++ only) |
| `&&`, `\|\|` | +1 | Boolean operators in conditions |
| `conditional_expression` | +1 | Ternary `a ? b : c` |

**Not counted:**
- `switch_statement` — not counted; only individual `case_statement` nodes add CC (matches McCabe definition: a switch's complexity is the number of cases, not the switch construct itself)
- `goto_statement` — goto doesn't create independent paths per McCabe
- `break` / `continue` — loop control doesn't add paths
- `return` — return statements don't add CC (function already has a return path)
- `default` in switch — McCabe counts cases as binary decisions (N-1 for N cases), not default

**Function discovery:** Walks tree-sitter CST for `function_definition` nodes. C++ additionally handles `struct_specifier` and `class_specifier` as containing contexts (member functions get qualified names like `MyClass::method`).

**Existing file modifications:**

`internal/config/config.go` — add to `CoverageConfig`:
```go
C   string `toml:"c"`
Cpp string `toml:"cpp"`
```

Default values in `DefaultConfig()`:
```go
C:   ".gcov",
Cpp: ".gcov",
```

`internal/engine/engine.go` — add to `detectLanguage`:
```go
case ".c", ".h":
    return "c"
case ".cpp", ".cc", ".cxx", ".hpp", ".hh":
    return "cpp"
```

`driveRegistry` — add:
```go
c_driver.New(),
cpp_driver.New(),
```

`parseCoverageByLang` — add:
```go
case "c", "cpp":
    return coverage.ParseGcov(path)
```

`CoveragePathForLang` — add:
```go
case "c":
    return c.Coverage.C
case "cpp":
    return c.Coverage.Cpp
```

### 2. gcov Coverage Parser

The `.gcov` format is generated by GCC's `gcov` tool (one file per source file, e.g., `main.cpp.gcov`).

**Format:**
```
        -:    0:Source:/path/to/main.cpp
        -:    1:#include <stdio.h>
        5:    2:int main() {
        #####:    3:    return -1;
        1:    4:    return 0;
        -:    5:}
```

**Column meanings:**
- `N:` (integer > 0) — line executed N times → covered
- `#####:` — line never executed → uncovered
- `-:` — non-executable (blank, comment, brace, declaration)

**Parsing:**
1. Read file line by line
2. Skip header line (contains `Source:` prefix)
3. For each line, extract the count prefix:
   - `-:` → non-executable → skip
   - `#####:` → line is uncovered → record line number in MissingLines
   - Any integer `N:` → line is covered → record in CoveredLines
4. Extract the source file path from the `Source:` header
5. Build `CoverageData{CoveredLines: map[int]bool, TotalLines: len(covered) + len(uncovered)}`

**Default coverage path:** `".gcov"` for both C and C++ (configurable via `.crap.toml` or env vars `CRAP_COVERAGE_C` / `CRAP_COVERAGE_CPP`).

**Filtering:** The parser should only process lines from the target source file. If the `.gcov` file contains data for multiple files (unlikely but possible with `gcov -b`), filter to the relevant file.

### 3. Validation Suite

#### Cross-Language CC Corpus

Two new fixture files in `validate/cc_corpus/fixtures/`:

- `equivalence.c` — 12 functions in C syntax (11 for C, since `try_catch` is not available)
- `equivalence.cpp` — 12 functions in C++ syntax (all 12)

The existing `corpus_test.go` already iterates over all drivers; the C and C++ drivers will be picked up automatically once registered. No changes to the test logic needed.

The `expected.json` file gets a `skip_c` array: `["try_catch"]` (C has no try/catch).

The existing `skip_go` entry for `try_catch` is already present.

**C function notes:**
- No `bool` type — use `int` with `1`/`0` for boolean functions
- `for_loop` uses `for (int i = 0; i < n; i++)`
- `switch_case` uses `switch` with `break` in each case
- `boolean_ops` uses `&&` and `||`
- `ternary` uses `a ? b : c`

**C++ function notes:**
- Same as C functions but with `bool` type and modern C++ syntax
- `try_catch` uses C++ `try`/`catch`

#### C++-Specific CC References

File: `validate/cc_ref/fixtures/ref_cpp.cpp`

Constructs without Python equivalents:
- Range-for: `for (auto& x : container)` — should be CC=2 (base + for)
- Multi-catch: `try { } catch (int) { } catch (double) { }` — should be CC=3 (base + 2 catches)
- Lambda as default argument: `void f(std::function<void()> cb = []{})` — lambda doesn't add CC to parent

Reference values generated once via `lizard` and committed as `validate/cc_ref/expected_cpp.json`.

#### Synthetic gcov Coverage

File: `validate/coverage/fixtures/add.c`

Two single-line functions for clean coverage ratios:
```c
int add(int a, int b) { return a + b; }
int sub(int a, int b) { return a - b; }
```

Three gcov fixtures:

| Fixture | Lines covered | Expected % |
|---------|--------------|------------|
| `full.gcov` | 2/2 | 100% |
| `half.gcov` | 1/2 | 50% |
| `none.gcov` | 0/2 | 0% |

The existing `coverage_test.go` needs one new test function: `TestGcovCoverage` that iterates over these three variants, sets `cfg.Coverage.C` to the gcov fixture path, runs `engine.Analyze`, and asserts coverage percentages match `expected.json` within 0.5 tolerance.

### 4. Test Integration

All new tests run under `go test ./...` with zero external dependencies:

```
go test ./internal/driver/c/        # C driver tests
go test ./internal/driver/cpp/      # C++ driver tests
go test ./internal/coverage/        # gcov parser tests
go test ./validate/...              # corpus/ref/coverage (auto-discovers C/C++ drivers)
```

## Future Considerations

- **Header-only coverage:** `.h`/`.hpp` files are currently excluded from language detection. If a project has significant logic in headers (e.g., template-heavy C++), header detection could be added later.
- **Multiple `.gcov` files:** The current design parses one `.gcov` file per run. For projects with multiple `.gcov` outputs, the user can point nocrap at a directory and each file will be loaded by the engine's file-walking logic (same as how `coverage.json` works today).
- **C-only CC refs:** Not included — C has no CC-impacting constructs that C++ doesn't also have. The cross-language corpus and C++ ref tests provide sufficient coverage.
