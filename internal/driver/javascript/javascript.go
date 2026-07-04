package javascript

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"nocrap/internal/driver"
)

type JavaScriptDriver struct{}

func New() *JavaScriptDriver {
	return &JavaScriptDriver{}
}

func (d *JavaScriptDriver) Name() string         { return "javascript" }
func (d *JavaScriptDriver) Extensions() []string { return []string{".js", ".jsx", ".mjs", ".cjs"} }

func (d *JavaScriptDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	return FindFunctionsWithLanguage(source, filePath, javascript.GetLanguage())
}

func (d *JavaScriptDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	return CalcComplexityWithLanguage(source, fn, javascript.GetLanguage())
}

// FindFunctionsWithLanguage is a shared helper used by both JavaScript and TypeScript drivers.
func FindFunctionsWithLanguage(source []byte, filePath string, lang *sitter.Language) ([]driver.Function, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()
	return walkForFunctions(root, source, filePath), nil
}

// CalcComplexityWithLanguage is a shared helper used by both JavaScript and TypeScript drivers.
func CalcComplexityWithLanguage(source []byte, fn driver.Function, lang *sitter.Language) (int, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return 0, fmt.Errorf("parsing for CC: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()
	funcNode := findJSFunctionNode(root, source, fn)
	if funcNode == nil {
		return 1, nil
	}

	cc := 1
	countJSCC(funcNode, &cc)
	return cc, nil
}

func handleVarDeclarator(node *sitter.Node, source []byte, filePath, className string) *driver.Function {
	value := node.ChildByFieldName("value")
	if value == nil {
		return nil
	}
	if value.Type() != "function_expression" && value.Type() != "arrow_function" {
		return nil
	}
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nameNode.Content(source)
	if className != "" {
		name = className + "." + name
	}
	return &driver.Function{
		Name:              name,
		File:              filePath,
		StartLine:         int(value.StartPoint().Row) + 1,
		EndLine:           int(value.EndPoint().Row) + 1,
		CoverageStartLine: int(value.StartPoint().Row) + 1,
		Package:           className,
	}
}

func walkForFunctions(root *sitter.Node, source []byte, filePath string) []driver.Function {
	w := &jsWalker{source: source, filePath: filePath}
	w.walk(root)
	return w.funcs
}

type jsWalker struct {
	source       []byte
	filePath     string
	funcs        []driver.Function
	currentClass string
}

func (w *jsWalker) walkClass(node *sitter.Node) {
	prevClass := w.currentClass
	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		w.currentClass = nameNode.Content(w.source)
	}
	w.walk(node.ChildByFieldName("body"))
	w.currentClass = prevClass
}

func (w *jsWalker) walk(node *sitter.Node) {
	if node == nil {
		return
	}
	switch node.Type() {
	case "class_declaration":
		w.walkClass(node)
	case "function_declaration":
		w.funcs = append(w.funcs, extractJSFunction(node, w.source, w.filePath, w.currentClass))
		w.walk(node.ChildByFieldName("body"))
	case "method_definition":
		w.funcs = append(w.funcs, extractJSMethod(node, w.source, w.filePath, w.currentClass))
		w.walk(node.ChildByFieldName("body"))
	case "variable_declarator":
		if fn := handleVarDeclarator(node, w.source, w.filePath, w.currentClass); fn != nil {
			w.funcs = append(w.funcs, *fn)
		}
	default:
		w.walkChildren(node)
	}
}

func (w *jsWalker) walkChildren(node *sitter.Node) {
	for i := 0; i < int(node.ChildCount()); i++ {
		w.walk(node.Child(i))
	}
}

func extractJSFunction(node *sitter.Node, source []byte, filePath, className string) driver.Function {
	name := ""
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		name = nameNode.Content(source)
	}
	startLine := int(node.StartPoint().Row) + 1
	endLine := int(node.EndPoint().Row) + 1

	fullName := name
	if className != "" {
		fullName = className + "." + name
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

func extractJSMethod(node *sitter.Node, source []byte, filePath, className string) driver.Function {
	name := ""
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		name = nameNode.Content(source)
	}
	startLine := int(node.StartPoint().Row) + 1
	endLine := int(node.EndPoint().Row) + 1

	fullName := name
	if className != "" {
		fullName = className + "." + name
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

func findJSFunctionNode(root *sitter.Node, source []byte, fn driver.Function) *sitter.Node {
	var found *sitter.Node
	var search func(node *sitter.Node)
	search = func(node *sitter.Node) {
		if found != nil {
			return
		}
		switch node.Type() {
		case "function_declaration", "function_expression", "arrow_function", "method_definition":
			startLine := int(node.StartPoint().Row) + 1
			if startLine == fn.StartLine {
				found = node
				return
			}
		}
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil {
				search(child)
			}
		}
	}
	search(root)
	return found
}

func isJSBranchNode(nodeType string) bool {
	switch nodeType {
	case "if_statement", "while_statement", "for_statement",
		"for_in_statement", "do_statement", "catch_clause",
		"switch_case", "switch_default", "ternary_expression",
		"&&", "||", "??", "optional_chain":
		return true
	}
	return false
}

func isOutermostOptionalChain(node *sitter.Node) bool {
	parent := node.Parent()
	if parent == nil || parent.Type() != "member_expression" {
		return true
	}
	grandparent := parent.Parent()
	if grandparent == nil || grandparent.Type() != "member_expression" {
		return true
	}
	for i := uint32(0); i < grandparent.ChildCount(); i++ {
		if child := grandparent.Child(int(i)); child != nil && child.Type() == "optional_chain" {
			return false
		}
	}
	return true
}

func countJSCC(node *sitter.Node, cc *int) {
	if node.Type() == "optional_chain" {
		if isOutermostOptionalChain(node) {
			*cc++
		}
	} else if isJSBranchNode(node.Type()) {
		*cc++
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		if child := node.Child(i); child != nil {
			countJSCC(child, cc)
		}
	}
}
