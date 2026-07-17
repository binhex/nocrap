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

func TestCalcComplexity_RadonUnavailable(t *testing.T) {
	// Verify that when radon is unavailable, CalcComplexity returns
	// CC=1 without error (graceful degradation).
	source := []byte("def add(a, b):\n    return a + b\n")
	d := New()
	funcs, err := d.FindFunctions(source, "test.py")
	if err != nil {
		t.Fatalf("FindFunctions: %v", err)
	}
	if len(funcs) < 1 {
		t.Fatal("expected at least 1 function")
	}

	for _, fn := range funcs {
		cc, err := d.CalcComplexity(source, fn)
		if err != nil {
			t.Errorf("CalcComplexity(%q): unexpected error: %v", fn.Name, err)
		}
		if cc < 1 {
			t.Errorf("CalcComplexity(%q) = %d, want >= 1", fn.Name, cc)
		}
	}
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
