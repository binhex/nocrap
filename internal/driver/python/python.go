package python

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"nocrap/internal/driver"
)

type PythonDriver struct{}

func New() *PythonDriver {
	return &PythonDriver{}
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

	var funcs []driver.Function
	var currentClass string

	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		switch node.Type() {
		case "class_definition":
			nameNode := node.ChildByFieldName("name")
			prevClass := currentClass
			if nameNode != nil {
				currentClass = nameNode.Content(source)
			}
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}
			currentClass = prevClass

		case "function_definition":
			fn := extractFunction(node, source, filePath, currentClass)
			funcs = append(funcs, fn)
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}

		default:
			for i := uint32(0); i < node.ChildCount(); i++ {
				child := node.Child(int(i))
				if child != nil {
					walk(child)
				}
			}
		}
	}
	walk(root)

	return funcs, nil
}

// extractFunction builds a Function from a tree-sitter function_definition node.
// Excludes decorator lines from StartLine and skips docstrings for CoverageStartLine.
func extractFunction(node *sitter.Node, source []byte, filePath string, className string) driver.Function {
	name := ""
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		name = nameNode.Content(source)
	}

	startLine := int(node.StartPoint().Row) + 1
	endLine := int(node.EndPoint().Row) + 1

	// Skip decorators: find first non-decorator child's start line
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child != nil && child.Type() != "decorator" && child.Type() != "comment" {
			startLine = int(child.StartPoint().Row) + 1
			break
		}
	}

	// Compute CoverageStartLine: skip docstring if first body statement is a string expression
	coverageStartLine := startLine
	body := node.ChildByFieldName("body")
	if body != nil && body.ChildCount() > 0 {
		firstStmt := body.Child(0)
		if firstStmt != nil && firstStmt.Type() == "expression_statement" {
			for i := uint32(0); i < firstStmt.ChildCount(); i++ {
				child := firstStmt.Child(int(i))
				if child != nil && child.Type() == "string" {
					docEndLine := int(child.EndPoint().Row) + 1
					coverageStartLine = docEndLine + 1
					break
				}
			}
		}
	}
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

func (d *PythonDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
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
	countCC(funcNode, &cc)
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
			nameNode := node.ChildByFieldName("name")
			if nameNode != nil {
				nodeName := nameNode.Content(source)
				nodeStart := int(node.StartPoint().Row) + 1
				// Strip class prefix: fn.Name is "Calculator.__init__" but nodeName is "__init__"
				fnName := fn.Name
				if idx := strings.LastIndex(fnName, "."); idx >= 0 {
					fnName = fnName[idx+1:]
				}
				if nodeStart == fn.StartLine && nodeName == fnName {
					found = node
					return
				}
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

func countCC(node *sitter.Node, cc *int) {
	switch node.Type() {
	case "if_statement":
		*cc++
	case "elif_clause":
		*cc++
	case "while_statement":
		*cc++
	case "for_statement":
		*cc++
	case "except_clause":
		*cc++
	// Note: `with` is NOT counted for CC to match radon's behavior
	// (pytest-crap via radon treats `with` as CC=1, not +1).
	case "match_statement":
		*cc++
	case "case_clause":
		*cc++
	case "and":
		*cc++
	case "or":
		*cc++
	}
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child != nil {
			countCC(child, cc)
		}
	}
}
