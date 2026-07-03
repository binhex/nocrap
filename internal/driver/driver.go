// Package driver defines the language-agnostic interface that every language
// driver must implement, plus the shared Function data type.
package driver

// Function represents a function or method discovered in source code.
type Function struct {
	Name              string // function name (dot-joined: "MyClass.method")
	File              string // path to the source file (as passed to the driver)
	StartLine         int    // 1-based line where the function definition starts (decorators excluded)
	EndLine           int    // 1-based line where the function ends
	CoverageStartLine int    // 1-based line where executable code starts (skips docstring if present)
	Package           string // class name, module name, namespace, or "" for top-level
}

// Driver is the interface that every language driver must implement.
type Driver interface {
	// Name returns the language name (e.g. "python", "javascript").
	Name() string

	// Extensions returns the file extensions this driver handles (e.g. [".py"]).
	Extensions() []string

	// FindFunctions parses source with tree-sitter and returns all
	// function/method definitions found in the source.
	FindFunctions(source []byte, filePath string) ([]Function, error)

	// CalcComplexity walks the CST subtree rooted at the given function and
	// returns its cyclomatic complexity (always >= 1).
	CalcComplexity(source []byte, fn Function) (int, error)
}
