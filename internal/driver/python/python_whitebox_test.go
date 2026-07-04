package python

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

func findFuncNode(root *sitter.Node) *sitter.Node {
	var result *sitter.Node
	var search func(*sitter.Node)
	search = func(node *sitter.Node) {
		if result != nil {
			return
		}
		if node.Type() == "function_definition" {
			result = node
			return
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			if child := node.Child(int(i)); child != nil {
				search(child)
			}
		}
	}
	search(root)
	return result
}

func TestSkipDocstring_WithDocstring(t *testing.T) {
	source := []byte("def foo():\n    \"\"\"doc\"\"\"\n    return 1\n")
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree := parser.Parse(nil, source)
	defer tree.Close()

	fnNode := findFuncNode(tree.RootNode())
	if fnNode == nil {
		t.Fatal("function_definition not found")
	}

	body := fnNode.ChildByFieldName("body")
	if body == nil {
		t.Fatal("body not found")
	}

	result := skipDocstring(body, 1)
	if result != 3 {
		t.Errorf("skipDocstring = %d, want 3", result)
	}
}

func TestSkipDocstring_NoDocstring(t *testing.T) {
	source := []byte("def foo():\n    return 1\n")
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree := parser.Parse(nil, source)
	defer tree.Close()

	fnNode := findFuncNode(tree.RootNode())
	if fnNode == nil {
		t.Fatal("function_definition not found")
	}

	body := fnNode.ChildByFieldName("body")
	if body == nil {
		t.Fatal("body not found")
	}

	result := skipDocstring(body, 5)
	if result != 5 {
		t.Errorf("skipDocstring = %d, want 5", result)
	}
}

func TestSkipDocstring_NilBody(t *testing.T) {
	result := skipDocstring(nil, 10)
	if result != 10 {
		t.Errorf("skipDocstring(nil) = %d, want 10", result)
	}
}

func TestSkipDocstring_EmptyBody(t *testing.T) {
	source := []byte("def foo(): pass\n")
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree := parser.Parse(nil, source)
	defer tree.Close()

	fnNode := findFuncNode(tree.RootNode())
	if fnNode == nil {
		t.Fatal("function_definition not found")
	}

	body := fnNode.ChildByFieldName("body")
	if body == nil {
		t.Fatal("body not found")
	}

	result := skipDocstring(body, 5)
	// pass statement is not a string expression, should return default
	if result != 5 {
		t.Errorf("skipDocstring(pass) = %d, want 5", result)
	}
}
