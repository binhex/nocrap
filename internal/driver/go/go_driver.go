package go_driver

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"nocrap/internal/driver"
)

type GoDriver struct{}

func New() *GoDriver {
	return &GoDriver{}
}

func (d *GoDriver) Name() string         { return "go" }
func (d *GoDriver) Extensions() []string { return []string{".go"} }

func (d *GoDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	var funcs []driver.Function

	packageName := ""
	for i := uint32(0); i < root.ChildCount(); i++ {
		child := root.Child(int(i))
		if child != nil && child.Type() == "package_clause" {
			for j := uint32(0); j < child.ChildCount(); j++ {
				gc := child.Child(int(j))
				if gc != nil && gc.Type() == "package_identifier" {
					packageName = gc.Content(source)
				}
			}
		}
	}

	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		switch node.Type() {
		case "function_declaration":
			fn := extractGoFunc(node, source, filePath, packageName)
			funcs = append(funcs, fn)
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}

		case "method_declaration":
			fn := extractGoMethod(node, source, filePath, packageName)
			funcs = append(funcs, fn)
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}
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

// extractTypeName extracts the type name from a tree-sitter node that may be
// a type_identifier, pointer_type, or a parameter_declaration containing the type.
func extractTypeName(node *sitter.Node, source []byte) string {
	switch node.Type() {
	case "type_identifier":
		return node.Content(source)
	case "pointer_type":
		// pointer_type has child * then type_identifier
		for i := uint32(0); i < node.ChildCount(); i++ {
			c := node.Child(int(i))
			if c != nil && c.Type() == "type_identifier" {
				return c.Content(source)
			}
		}
		return ""
	case "parameter_declaration":
		// Look for the type child (either type_identifier or pointer_type)
		for i := uint32(0); i < node.ChildCount(); i++ {
			c := node.Child(int(i))
			if c != nil {
				if t := extractTypeName(c, source); t != "" {
					return t
				}
			}
		}
	}
	return ""
}

func extractGoFunc(node *sitter.Node, source []byte, filePath, pkg string) driver.Function {
	name := ""
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		name = nameNode.Content(source)
	}
	fullName := pkg + "." + name
	startLine := int(node.StartPoint().Row) + 1
	return driver.Function{
		Name:              fullName,
		File:              filePath,
		StartLine:         startLine,
		EndLine:           int(node.EndPoint().Row) + 1,
		CoverageStartLine: startLine,
		Package:           pkg,
	}
}

func extractGoMethod(node *sitter.Node, source []byte, filePath, pkg string) driver.Function {
	name := ""
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		name = nameNode.Content(source)
	}

	receiver := node.ChildByFieldName("receiver")
	recvName := ""
	if receiver != nil {
		for i := uint32(0); i < receiver.ChildCount(); i++ {
			rc := receiver.Child(int(i))
			if rc != nil {
				recvName = extractTypeName(rc, source)
				if recvName != "" {
					break
				}
			}
		}
	}

	fullName := name
	if recvName != "" {
		fullName = recvName + "." + name
	}
	startLine := int(node.StartPoint().Row) + 1
	return driver.Function{
		Name:              fullName,
		File:              filePath,
		StartLine:         startLine,
		EndLine:           int(node.EndPoint().Row) + 1,
		CoverageStartLine: startLine,
		Package:           pkg,
	}
}

func (d *GoDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return 0, fmt.Errorf("parsing for CC: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()
	funcNode := findGoFuncNode(root, source, fn)
	if funcNode == nil {
		return 1, nil
	}

	cc := 1
	countGoCC(funcNode, &cc)
	return cc, nil
}

func findGoFuncNode(root *sitter.Node, source []byte, fn driver.Function) *sitter.Node {
	var found *sitter.Node
	var search func(node *sitter.Node)
	search = func(node *sitter.Node) {
		if found != nil {
			return
		}
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			startLine := int(node.StartPoint().Row) + 1
			if startLine == fn.StartLine {
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

func countGoCC(node *sitter.Node, cc *int) {
	switch node.Type() {
	case "if_statement":
		*cc++
	case "for_statement":
		*cc++
	case "expression_switch_statement":
		*cc++
	case "type_switch_statement":
		*cc++
	case "expression_case":
		*cc++
	case "type_case":
		*cc++
	case "default_case":
		*cc++
	case "select_statement":
		*cc++
	case "communication_case":
		*cc++
	case "&&", "||":
		*cc++
	}
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child != nil {
			countGoCC(child, cc)
		}
	}
}
