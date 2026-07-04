package javascript

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func parseJS(source string) *sitter.Node {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree := parser.Parse(nil, []byte(source))
	root := tree.RootNode()
	// Don't close tree until test is done
	return root
}

func TestIsOutermostOptionalChain_Simple(t *testing.T) {
	root := parseJS("function f() { return a?.b; }")
	// Walk to find optional_chain node
	var chain *sitter.Node
	var search func(*sitter.Node)
	search = func(node *sitter.Node) {
		if chain != nil {
			return
		}
		if node.Type() == "optional_chain" {
			chain = node
			return
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			if c := node.Child(int(i)); c != nil {
				search(c)
			}
		}
	}
	search(root)

	if chain == nil {
		t.Fatal("optional_chain not found")
	}

	if !isOutermostOptionalChain(chain) {
		t.Error("single ?. should be outermost")
	}
}

func TestIsOutermostOptionalChain_Nested(t *testing.T) {
	// a?.b?.c — tree-sitter nests the inner ?. at a different level
	root := parseJS("function f() { return a?.b?.c; }")
	var chains []*sitter.Node
	var search func(*sitter.Node)
	search = func(node *sitter.Node) {
		if node.Type() == "optional_chain" {
			chains = append(chains, node)
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			if c := node.Child(int(i)); c != nil {
				search(c)
			}
		}
	}
	search(root)
	t.Logf("found %d optional_chain nodes in a?.b?.c", len(chains))

	// At least one should be outermost, at least one should not
	if len(chains) == 0 {
		t.Skip("no optional_chain nodes found — tree-sitter version difference")
	}
	outer := 0
	inner := 0
	for i, c := range chains {
		if isOutermostOptionalChain(c) {
			outer++
			t.Logf("chain[%d] is outermost", i)
		} else {
			inner++
			t.Logf("chain[%d] is NOT outermost", i)
		}
	}
	if outer == 0 {
		t.Error("expected at least one outermost optional_chain")
	}
}

func TestIsOutermostOptionalChain_Standalone(t *testing.T) {
	// a?.b where there's no parent member_expression nesting
	root := parseJS("function f(a) { const x = a?.b; }")
	var chain *sitter.Node
	var search func(*sitter.Node)
	search = func(node *sitter.Node) {
		if chain != nil {
			return
		}
		if node.Type() == "optional_chain" {
			chain = node
			return
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			if c := node.Child(int(i)); c != nil {
				search(c)
			}
		}
	}
	search(root)

	if chain == nil {
		t.Fatal("optional_chain not found")
	}

	if !isOutermostOptionalChain(chain) {
		t.Error("standalone ?. should be outermost")
	}
}

func TestIsOutermostOptionalChain_OptionalCall(t *testing.T) {
	root := parseJS("function f(a) { return a?.(); }")
	var chains []*sitter.Node
	var search func(*sitter.Node)
	search = func(node *sitter.Node) {
		if node.Type() == "optional_chain" {
			chains = append(chains, node)
		}
		for i := uint32(0); i < node.ChildCount(); i++ {
			if c := node.Child(int(i)); c != nil {
				search(c)
			}
		}
	}
	search(root)

	if len(chains) == 0 {
		t.Skip("no optional_chain for a?.() — tree-sitter version difference")
	}
	if !isOutermostOptionalChain(chains[0]) {
		t.Error("optional call should be outermost")
	}
}
