// Package cc_corpus_test verifies that 12 equivalent functions across Python, JS, TS,
// Go, C, and C++ all produce the same cyclomatic complexity when run through nocrap's
// language drivers. Expected values are radon-verified ground truth stored in expected.json.
package cc_corpus_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nocrap/internal/driver"
	cDriver "nocrap/internal/driver/c"
	cppDriver "nocrap/internal/driver/cpp"
	goDriver "nocrap/internal/driver/go"
	"nocrap/internal/driver/javascript"
	"nocrap/internal/driver/python"
	"nocrap/internal/driver/typescript"
)

// expectedData mirrors validate/cc_corpus/expected.json.
type expectedData struct {
	Functions map[string]int `json:"functions"`
	SkipGo    []string       `json:"skip_go"`
	SkipC     []string       `json:"skip_c"`
	SkipCpp   []string       `json:"skip_cpp"`
}

// langEntry binds a language name to its driver and fixture extension.
type langEntry struct {
	name       string
	driver     driver.Driver
	ext        string
	fixtureDir string
}

// TestCCEquivalenceAcrossLanguages verifies that every function in the
// cross-language corpus has the same cyclomatic complexity in all supported
// languages, using expected.json as ground truth.
func TestCCEquivalenceAcrossLanguages(t *testing.T) {
	raw, err := os.ReadFile("expected.json")
	if err != nil {
		t.Fatalf("reading expected.json: %v", err)
	}
	var exp expectedData
	if err := json.Unmarshal(raw, &exp); err != nil {
		t.Fatalf("parsing expected.json: %v", err)
	}

	// Build skip sets per language.
	skipByLang := make(map[string]map[string]bool)
	addSkips := func(lang string, names []string) {
		for _, name := range names {
			if skipByLang[lang] == nil {
				skipByLang[lang] = make(map[string]bool)
			}
			skipByLang[lang][name] = true
		}
	}
	addSkips("go", exp.SkipGo)
	addSkips("c", exp.SkipC)
	addSkips("cpp", exp.SkipCpp)

	// Map expected.json snake_case keys to the function name each driver
	// returns from FindFunctions.
	nameFor := map[string]func(string) string{
		"python":     func(key string) string { return key },
		"javascript": func(key string) string { return key },
		"typescript": func(key string) string { return key },
		"go":         snakeToGo,
		"c":          func(key string) string { return key },
		"cpp":        func(key string) string { return key },
	}

	languages := []langEntry{
		{name: "python", driver: python.New(), ext: ".py", fixtureDir: "fixtures"},
		{name: "javascript", driver: javascript.New(), ext: ".js", fixtureDir: "fixtures"},
		{name: "typescript", driver: typescript.New(), ext: ".ts", fixtureDir: "fixtures"},
		{name: "go", driver: goDriver.New(), ext: ".go", fixtureDir: "fixtures"},
		{name: "c", driver: cDriver.New(), ext: ".c", fixtureDir: "fixtures_cc"},
		{name: "cpp", driver: cppDriver.New(), ext: ".cpp", fixtureDir: "fixtures_cc"},
	}

	for _, lang := range languages {
		lang := lang // capture
		t.Run(lang.name, func(t *testing.T) {
			fixturePath := filepath.Join(lang.fixtureDir, "equivalence"+lang.ext)
			source, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("reading fixture %s: %v", fixturePath, err)
			}

			funcs, err := lang.driver.FindFunctions(source, fixturePath)
			if err != nil {
				t.Fatalf("FindFunctions(%s): %v", fixturePath, err)
			}

			// Build lookup: driver function name → Function.
			fnByName := make(map[string]driver.Function, len(funcs))
			for _, fn := range funcs {
				fnByName[fn.Name] = fn
			}

			toName := nameFor[lang.name]
			skipSet := skipByLang[lang.name]

			for key, expectedCC := range exp.Functions {
				// Skip entries excluded for this language.
				if skipSet[key] {
					continue
				}

				fnName := toName(key)
				fn, ok := fnByName[fnName]
				if !ok {
					t.Errorf("function %q (expected key=%q) not found; available: %v",
						fnName, key, funcNames(funcs))
					continue
				}

				cc, err := lang.driver.CalcComplexity(source, fn)
				if err != nil {
					t.Fatalf("CalcComplexity(%s): %v", key, err)
				}
				if cc != expectedCC {
					t.Errorf("%s: CalcComplexity(%s) = %d, want %d (radon-verified ground truth)",
						lang.name, key, cc, expectedCC)
				}
			}
		})
	}
}

// snakeToGo converts "no_branches" to "fixtures.NoBranches".
func snakeToGo(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return "fixtures." + strings.Join(parts, "")
}

// funcNames returns a sorted-like slice of function names for diagnostics.
func funcNames(funcs []driver.Function) []string {
	names := make([]string, len(funcs))
	for i, f := range funcs {
		names[i] = f.Name
	}
	return names
}
