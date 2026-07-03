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

func walkForFunctions(root *sitter.Node, source []byte, filePath string) []driver.Function {
	var funcs []driver.Function
	var currentClass string

	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		switch node.Type() {
		case "class_declaration":
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
			return // do NOT recurse via default — body explicitly walked above

		case "function_declaration":
			fn := extractJSFunction(node, source, filePath, currentClass)
			funcs = append(funcs, fn)
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}
			return

		case "method_definition":
			fn := extractJSMethod(node, source, filePath, currentClass)
			funcs = append(funcs, fn)
			body := node.ChildByFieldName("body")
			if body != nil {
				walk(body)
			}
			return

		case "variable_declarator":
			value := node.ChildByFieldName("value")
			if value != nil && (value.Type() == "function_expression" || value.Type() == "arrow_function") {
				nameNode := node.ChildByFieldName("name")
				if nameNode != nil {
					name := nameNode.Content(source)
					startLine := int(value.StartPoint().Row) + 1
					endLine := int(value.EndPoint().Row) + 1
					nameToUse := name
					if currentClass != "" {
						nameToUse = currentClass + "." + name
					}
					funcs = append(funcs, driver.Function{
						Name:              nameToUse,
						File:              filePath,
						StartLine:         startLine,
						EndLine:           endLine,
						CoverageStartLine: startLine,
						Package:           currentClass,
					})
				}
			}
			return

		default:
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					walk(child)
				}
			}
		}
	}
	walk(root)
	return funcs
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

func countJSCC(node *sitter.Node, cc *int) {
	switch node.Type() {
	case "if_statement":
		*cc++
	case "while_statement":
		*cc++
	case "for_statement":
		*cc++
	case "for_in_statement":
		*cc++
	case "do_statement":
		*cc++
	case "catch_clause":
		*cc++
	case "switch_case", "switch_default":
		*cc++
	case "ternary_expression":
		*cc++
	case "optional_chain_expression":
		*cc++
	case "&&", "||", "??":
		*cc++
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			countJSCC(child, cc)
		}
	}
}
