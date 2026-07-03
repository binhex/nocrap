package typescript

import (
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"nocrap/internal/driver"
	"nocrap/internal/driver/javascript"
)

type TypeScriptDriver struct{}

func New() *TypeScriptDriver {
	return &TypeScriptDriver{}
}

func (d *TypeScriptDriver) Name() string         { return "typescript" }
func (d *TypeScriptDriver) Extensions() []string { return []string{".ts", ".tsx"} }

func (d *TypeScriptDriver) FindFunctions(source []byte, filePath string) ([]driver.Function, error) {
	return javascript.FindFunctionsWithLanguage(source, filePath, typescript.GetLanguage())
}

func (d *TypeScriptDriver) CalcComplexity(source []byte, fn driver.Function) (int, error) {
	return javascript.CalcComplexityWithLanguage(source, fn, typescript.GetLanguage())
}
