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

// hasFunctionPattern is a local wrapper around cdriver's implementation
// that checks for C/C++ function-like patterns in source bytes.
var hasFunctionPattern = cdriver.HasFunctionPattern

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
		// If the file has no function-like patterns, it's likely a
		// data-only header that can't be parsed standalone. Skip silently.
		if !hasFunctionPattern(source) {
			return nil, nil
		}
		return nil, fmt.Errorf("parse error in %s", filePath)
	}

	w := &cppWalker{funcs: nil, currentClass: ""}
	w.walk(root, source, filePath)
	return w.funcs, nil
}

// cppWalker walks the C++ AST to find function definitions.
type cppWalker struct {
	funcs        []driver.Function
	currentClass string
}

func (w *cppWalker) walk(node *sitter.Node, source []byte, filePath string) {
	switch node.Type() {
	case "class_specifier", "struct_specifier":
		w.walkClass(node, source, filePath)
	case "function_definition":
		w.walkFunction(node, source, filePath)
	default:
		for i := uint32(0); i < node.ChildCount(); i++ {
			child := node.Child(int(i))
			if child != nil {
				w.walk(child, source, filePath)
			}
		}
	}
}

func (w *cppWalker) walkClass(node *sitter.Node, source []byte, filePath string) {
	nameNode := node.ChildByFieldName("name")
	prevClass := w.currentClass
	if nameNode != nil {
		if w.currentClass != "" {
			w.currentClass = w.currentClass + "::" + nameNode.Content(source)
		} else {
			w.currentClass = nameNode.Content(source)
		}
	}
	body := node.ChildByFieldName("body")
	if body != nil {
		w.walk(body, source, filePath)
	}
	w.currentClass = prevClass
}

func (w *cppWalker) walkFunction(node *sitter.Node, source []byte, filePath string) {
	fn := extractFunction(node, source, filePath, w.currentClass)
	w.funcs = append(w.funcs, fn)
	body := node.ChildByFieldName("body")
	if body != nil {
		w.walk(body, source, filePath)
	}
}

func extractFunction(node *sitter.Node, source []byte, filePath, className string) driver.Function {
	name := ""
	if decl := node.ChildByFieldName("declarator"); decl != nil {
		name = extractDeclaratorName(decl, source)
	}
	if name == "" {
		name = findFallbackName(node, source)
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

// findFallbackName tries to extract a function name when the declarator field is missing.
// This handles conversion operators and out-of-class qualified names.
func findFallbackName(node *sitter.Node, source []byte) string {
	for i := uint32(0); i < node.ChildCount(); i++ {
		child := node.Child(int(i))
		if child == nil {
			continue
		}
		switch child.Type() {
		case "qualified_identifier", "nested_identifier":
			return extractQualifiedName(child, source)
		case "operator_cast":
			return extractOperatorCast(child, source)
		}
	}
	return ""
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
		case "identifier", "field_identifier", "operator_name", "destructor_name":
			return child.Content(source)
		case "qualified_identifier", "nested_identifier":
			return extractQualifiedName(child, source)
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
		case "namespace_identifier", "identifier", "operator_name", "destructor_name":
			parts = append(parts, gc.Content(source))
		case "operator_cast":
			parts = append(parts, extractOperatorCast(gc, source))
		case "qualified_identifier":
			parts = append(parts, gc.Content(source))
		}
	}
	return strings.Join(parts, "::")
}

// extractOperatorCast extracts the name from an operator_cast node.
func extractOperatorCast(node *sitter.Node, source []byte) string {
	var parts []string
	for k := uint32(0); k < node.ChildCount(); k++ {
		inner := node.Child(int(k))
		if inner != nil && inner.Type() != "abstract_function_declarator" {
			parts = append(parts, inner.Content(source))
		}
	}
	return strings.Join(parts, " ")
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
