package c

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"nocrap/internal/driver"
)

type CDriver struct{}

func New() *CDriver { return &CDriver{} }

func (d *CDriver) Name() string         { return "c" }
func (d *CDriver) Extensions() []string { return []string{".c", ".h"} }

// HasFunctionPattern does a quick scan of source bytes to check if
// the file contains any C/C++ function-like patterns (i.e., `)`
// followed by `{`). Data-only headers (like expat's asciitab.h) that
// are meant to be #included inside array initializers have no such
// patterns and should be skipped silently.
func HasFunctionPattern(source []byte) bool {
	s := string(source)
	return strings.Contains(s, "){") ||
		strings.Contains(s, ") {") ||
		strings.Contains(s, ")\n{") ||
		strings.Contains(s, ")\t{") ||
		strings.Contains(s, ")\r\n{") ||
		strings.Contains(s, ")\n\t{") ||
		strings.Contains(s, ")\n    {")
}

func (d *CDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()

	// Always walk the tree to find function definitions, even if there are
	// parse errors. Tree-sitter produces a partial AST with valid subtrees
	// for the parseable portions of the file.
	var funcs []driver.Function
	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if node.Type() == "function_definition" {
			fn := extractFunction(node, source, filePath)
			funcs = append(funcs, fn)
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			if child != nil {
				walk(child)
			}
		}
	}
	walk(root)

	// After the walk, check whether we actually found any functions.
	// If the tree had errors and no functions were found in the partial AST,
	// decide whether to skip silently (data-only header) or report the error.
	if len(funcs) == 0 && root.HasError() {
		if !HasFunctionPattern(source) {
			return nil, nil
		}
		return nil, fmt.Errorf("parse error in %s", filePath)
	}

	return funcs, nil
}

func extractFunction(node *sitter.Node, source []byte, filePath string) driver.Function {
	name := ""
	if decl := node.ChildByFieldName("declarator"); decl != nil {
		name = extractDeclaratorName(decl, source)
	}
	startLine := int(node.StartPoint().Row) + 1
	endLine := int(node.EndPoint().Row) + 1
	return driver.Function{
		Name:              name,
		File:              filePath,
		StartLine:         startLine,
		EndLine:           endLine,
		CoverageStartLine: startLine,
	}
}

// extractDeclaratorName walks into a function_declarator to find the identifier.
func extractDeclaratorName(decl *sitter.Node, source []byte) string {
	for i := uint32(0); i < decl.ChildCount(); i++ {
		child := decl.Child(int(i))
		if child != nil && child.Type() == "identifier" {
			return child.Content(source)
		}
		// Handle nested declarators (e.g. function pointers)
		if child != nil && (child.Type() == "function_declarator" || child.Type() == "pointer_declarator") {
			if n := extractDeclaratorName(child, source); n != "" {
				return n
			}
		}
	}
	return ""
}

func (d *CDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return 0, fmt.Errorf("parsing for CC: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()
	// Attempt to find the function node even if the tree has errors.
	// Tree-sitter produces a partial AST with valid subtrees, so
	// function_definition nodes may still be found and analyzed.
	funcNode := findFunctionNode(root, source, fn)
	if funcNode == nil {
		return 1, nil
	}

	cc := 1
	CountCC(funcNode, &cc)
	return cc, nil
}

func findFunctionNode(root *sitter.Node, source []byte, fn driver.Function) *sitter.Node {
	var found *sitter.Node
	var search func(node *sitter.Node)
	search = func(node *sitter.Node) {
		if found != nil {
			return
		}
		if node.Type() == "function_definition" {
			nodeStart := int(node.StartPoint().Row) + 1
			if nodeStart == fn.StartLine {
				found = node
				return
			}
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			if child != nil {
				search(child)
			}
		}
	}
	search(root)
	return found
}

// CountCC counts cyclomatic complexity decision points in a C/C++ function node.
// Exported so the C++ driver can reuse it.
func CountCC(node *sitter.Node, cc *int) {
	countCCDecision(node.Type(), cc)
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child != nil {
			CountCC(child, cc)
		}
	}
}

// countCCDecision increments CC if the node type is a McCabe decision point.
func countCCDecision(nodeType string, cc *int) {
	switch nodeType {
	case "if_statement", "for_statement", "while_statement", "do_statement",
		"case_statement", "catch_clause", "conditional_expression":
		*cc++
	case "&&", "||":
		*cc++
	}
}
