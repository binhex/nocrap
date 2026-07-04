package cpp

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"nocrap/internal/driver"
	cdriver "nocrap/internal/driver/c"
)

type CppDriver struct{}

func New() *CppDriver { return &CppDriver{} }

func (d *CppDriver) Name() string         { return "cpp" }
func (d *CppDriver) Extensions() []string { return []string{".cpp", ".cc", ".cxx", ".hpp", ".hh"} }

func (d *CppDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(cpp.GetLanguage())
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
		case "class_specifier", "struct_specifier":
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

func extractFunction(node *sitter.Node, source []byte, filePath, className string) driver.Function {
	name := ""
	if decl := node.ChildByFieldName("declarator"); decl != nil {
		name = extractDeclaratorName(decl, source)
	}
	startLine := int(node.StartPoint().Row) + 1
	endLine := int(node.EndPoint().Row) + 1

	fullName := name
	if className != "" {
		fullName = className + "::" + name
	}

	return driver.Function{
		Name:              fullName,
		File:              filePath,
		StartLine:         startLine,
		EndLine:           endLine,
		CoverageStartLine: startLine,
		Package:           className,
	}
}

// extractDeclaratorName walks into a function_declarator to find the identifier or field_identifier.
func extractDeclaratorName(decl *sitter.Node, source []byte) string {
	for i := uint32(0); i < decl.ChildCount(); i++ {
		child := decl.Child(int(i))
		if child != nil && (child.Type() == "identifier" || child.Type() == "field_identifier") {
			return child.Content(source)
		}
		// Handle nested declarators
		if child != nil && (child.Type() == "function_declarator" || child.Type() == "pointer_declarator") {
			if n := extractDeclaratorName(child, source); n != "" {
				return n
			}
		}
	}
	return ""
}

func (d *CppDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(cpp.GetLanguage())
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
	cdriver.CountCC(funcNode, &cc)
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
