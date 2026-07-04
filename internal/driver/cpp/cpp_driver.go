package cpp

import (
	"context"
	"fmt"
	"strings"

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
				if currentClass != "" {
					currentClass = currentClass + "::" + nameNode.Content(source)
				} else {
					currentClass = nameNode.Content(source)
				}
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

// extractDeclaratorName walks into a function_declarator to find the function name.
// Handles named functions, operators, destructors, and conversion operators.
// For out-of-class definitions (e.g. Calculator::add), returns the full qualified name.
func extractDeclaratorName(decl *sitter.Node, source []byte) string {
	for i := uint32(0); i < decl.ChildCount(); i++ {
		child := decl.Child(int(i))
		if child == nil {
			continue
		}
		switch child.Type() {
		case "identifier", "field_identifier":
			return child.Content(source)
		case "qualified_identifier", "nested_identifier":
			return extractQualifiedName(child, source)
		case "operator_name":
			// operator+, operator bool, etc.
			return child.Content(source)
		case "destructor_name":
			// ~ClassName
			return child.Content(source)
		case "function_declarator", "pointer_declarator":
			// Recurse into nested declarators
			if n := extractDeclaratorName(child, source); n != "" {
				return n
			}
		}
	}
	return ""
}

// extractQualifiedName builds the full qualified name from a qualified_identifier
// or nested_identifier node, including namespace prefix and handling conversion operators.
func extractQualifiedName(node *sitter.Node, source []byte) string {
	var parts []string
	for j := uint32(0); j < node.ChildCount(); j++ {
		gc := node.Child(int(j))
		if gc == nil {
			continue
		}
		switch gc.Type() {
		case "namespace_identifier", "identifier":
			parts = append(parts, gc.Content(source))
		case "operator_name":
			parts = append(parts, gc.Content(source))
		case "destructor_name":
			parts = append(parts, gc.Content(source))
		case "operator_cast":
			// Conversion operator: extract "operator" + return type, skip parameter list
			var castParts []string
			for k := uint32(0); k < gc.ChildCount(); k++ {
				inner := gc.Child(int(k))
				if inner != nil && inner.Type() != "abstract_function_declarator" {
					castParts = append(castParts, inner.Content(source))
				}
			}
			parts = append(parts, strings.Join(castParts, " "))
		case "qualified_identifier":
			// Nested qualified identifier (e.g., B::method inside A::B::method)
			parts = append(parts, gc.Content(source))
		}
	}
	return strings.Join(parts, "::")
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
