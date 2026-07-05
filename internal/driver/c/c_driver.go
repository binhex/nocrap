package c

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"nocrap/internal/driver"
)

type CDriver struct{}

func New() *CDriver { return &CDriver{} }

func (d *CDriver) Name() string         { return "c" }
func (d *CDriver) Extensions() []string { return []string{".c", ".h"} }

func (d *CDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(c.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	if root.HasError() {
		return nil, fmt.Errorf("parse error in %s", filePath)
	}

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
	if root.HasError() {
		return 0, fmt.Errorf("parse error computing CC for %s in %s", fn.Name, fn.File)
	}

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
