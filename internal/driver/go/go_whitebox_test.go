package go_driver

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func TestExtractPackageName(t *testing.T) {
	source := []byte("package main\n\nfunc main() {\n}\n")
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree := parser.Parse(nil, source)
	defer tree.Close()

	name := extractPackageName(tree.RootNode(), source)
	if name != "main" {
		t.Errorf("expected 'main', got %q", name)
	}
}

func TestExtractPackageName_NoPackage(t *testing.T) {
	source := []byte("\n\nfunc main() {\n}\n")
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree := parser.Parse(nil, source)
	defer tree.Close()

	name := extractPackageName(tree.RootNode(), source)
	if name != "" {
		t.Errorf("expected empty, got %q", name)
	}
}
