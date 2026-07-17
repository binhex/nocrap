package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"nocrap/internal/driver"
)

type PythonDriver struct {
	mu             sync.Mutex
	cache          map[string]map[int]int // filePath -> startLine -> CC
	radonAvailable bool                   // cached result of radon availability check
	radonChecked   bool                   // true once radon has been checked
	radonWarned    bool                   // true once the missing-radon warning has been printed
}

func New() *PythonDriver {
	return &PythonDriver{}
}

// checkRadonAvailable checks whether the radon Python module is available.
func (d *PythonDriver) checkRadonAvailable() bool {
	cmd := exec.Command("python3", "-c", "from radon.complexity import cc_visit; print('ok')")
	return cmd.Run() == nil
}

func (d *PythonDriver) Name() string         { return "python" }
func (d *PythonDriver) Extensions() []string { return []string{".py"} }

func (d *PythonDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	if root.HasError() {
		return nil, fmt.Errorf("parse error in %s", filePath)
	}

	w := &pyWalker{source: source, filePath: filePath}
	w.walk(root)
	return w.funcs, nil
}

type pyWalker struct {
	source       []byte
	filePath     string
	funcs        []driver.Function
	currentClass string
}

func (w *pyWalker) walk(node *sitter.Node) {
	if node == nil {
		return
	}
	switch node.Type() {
	case "class_definition":
		w.walkClass(node)
		return
	case "function_definition":
		w.walkFunction(node)
		return
	}
	for i := uint32(0); i < node.ChildCount(); i++ {
		w.walk(node.Child(int(i)))
	}
}

func (w *pyWalker) walkClass(node *sitter.Node) {
	prevClass := w.currentClass
	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		w.currentClass = nameNode.Content(w.source)
	}
	w.walk(node.ChildByFieldName("body"))
	w.currentClass = prevClass
}

func (w *pyWalker) walkFunction(node *sitter.Node) {
	w.funcs = append(w.funcs, extractFunction(node, w.source, w.filePath, w.currentClass))
	w.walk(node.ChildByFieldName("body"))
}

func skipDecoratorLines(node *sitter.Node) int {
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child != nil && child.Type() != "decorator" && child.Type() != "comment" {
			return int(child.StartPoint().Row) + 1
		}
	}
	return int(node.StartPoint().Row) + 1
}

func skipDocstring(body *sitter.Node, defaultLine int) int {
	if body == nil || body.ChildCount() == 0 {
		return defaultLine
	}
	firstStmt := body.Child(0)
	if firstStmt == nil || firstStmt.Type() != "expression_statement" {
		return defaultLine
	}
	for i := uint32(0); i < firstStmt.ChildCount(); i++ {
		if child := firstStmt.Child(int(i)); child != nil && child.Type() == "string" {
			return int(child.EndPoint().Row) + 2
		}
	}
	return defaultLine
}

// extractFunction builds a Function from a tree-sitter function_definition node.
// Excludes decorator lines from StartLine and skips docstrings for CoverageStartLine.
func extractFunction(node *sitter.Node, source []byte, filePath string, className string) driver.Function {
	name := ""
	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		name = nameNode.Content(source)
	}

	startLine := skipDecoratorLines(node)
	endLine := int(node.EndPoint().Row) + 1
	coverageStartLine := skipDocstring(node.ChildByFieldName("body"), startLine)
	if coverageStartLine > endLine {
		coverageStartLine = endLine + 1
	}

	fullName := name
	if className != "" {
		fullName = className + "." + name
	}

	return driver.Function{
		Name:              fullName,
		File:              filePath,
		StartLine:         startLine,
		EndLine:           endLine,
		CoverageStartLine: coverageStartLine,
		Package:           className,
	}
}

func resolveFilePath(source []byte, filePath string) (string, func(), error) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return filePath, func() {}, nil
	}
	tmpFile, err := os.CreateTemp("", "nocrap-*.py")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp file: %w", err)
	}
	if _, err := tmpFile.Write(source); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }, nil
}

func runRadonCC(filePath string) (map[int]int, error) {
	cmd := exec.Command("python3", "-c", radonScript, filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("radon CC: %w", err)
	}
	var blocks []struct {
		Name       string `json:"name"`
		StartLine  int    `json:"start_line"`
		Complexity int    `json:"complexity"`
	}
	if err := json.Unmarshal(output, &blocks); err != nil {
		return nil, fmt.Errorf("parsing radon output: %w", err)
	}
	ccMap := make(map[int]int, len(blocks))
	for _, b := range blocks {
		ccMap[b.StartLine] = b.Complexity
	}
	return ccMap, nil
}

const radonScript = `import json, sys
from radon.complexity import cc_visit
file_path = sys.argv[1]
with open(file_path) as f:
    src = f.read()
blocks = cc_visit(src)
results = []
for b in blocks:
    results.append({"name": b.name, "start_line": b.lineno, "complexity": b.complexity})
print(json.dumps(results))
`

func (d *PythonDriver) getCached(file string, startLine int) (int, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cache == nil {
		return 0, false
	}
	ccMap, ok := d.cache[file]
	if !ok {
		return 0, false
	}
	cc, ok := ccMap[startLine]
	return cc, ok
}

func (d *PythonDriver) mergeCache(file string, newMap map[int]int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if existing, ok := d.cache[file]; ok {
		for k, v := range existing {
			if _, has := newMap[k]; !has {
				newMap[k] = v
			}
		}
	}
	if d.cache == nil {
		d.cache = make(map[string]map[int]int)
	}
	d.cache[file] = newMap
}

func (d *PythonDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	// Check radon availability once and cache the result.
	if !d.radonChecked {
		d.mu.Lock()
		if !d.radonChecked {
			d.radonChecked = true
			d.radonAvailable = d.wrappedRadonCheck()
			if !d.radonAvailable && !d.radonWarned {
				d.radonWarned = true
				fmt.Fprintf(os.Stderr, "warning: radon not available, install with: pip install radon\n")
			}
		}
		d.mu.Unlock()
	}

	if d.radonAvailable {
		if cc, ok := d.getCached(fn.File, fn.StartLine); ok {
			return cc, nil
		}

		filePath, cleanup, err := resolveFilePath(source, fn.File)
		if err != nil {
			return 0, err
		}
		defer cleanup()

		ccMap, err := runRadonCC(filePath)
		if err != nil {
			return 0, fmt.Errorf("radon CC for %s:%d: %w", fn.File, fn.StartLine, err)
		}

		d.mergeCache(fn.File, ccMap)
		if cc, ok := ccMap[fn.StartLine]; ok {
			return cc, nil
		}
	}

	return 1, nil
}

// wrappedRadonCheck delegates to checkRadonAvailable.
// Wrapped as a method so the call site in CalcComplexity reads clearly.
func (d *PythonDriver) wrappedRadonCheck() bool {
	return d.checkRadonAvailable()
}
