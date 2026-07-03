package engine

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"nocrap/internal/calculator"
	"nocrap/internal/config"
	"nocrap/internal/coverage"
	"nocrap/internal/driver"
	goDriver "nocrap/internal/driver/go"
	jsDriver "nocrap/internal/driver/javascript"
	pyDriver "nocrap/internal/driver/python"
	tsDriver "nocrap/internal/driver/typescript"
)

type FunctionScore struct {
	Name            string
	File            string
	StartLine       int
	EndLine         int
	CC              int
	CoveragePercent float64
	CRAP            float64
}

var drivers = []driver.Driver{
	pyDriver.New(),
	jsDriver.New(),
	tsDriver.New(),
	goDriver.New(),
}

func Analyze(paths []string, cfg *config.Config) ([]FunctionScore, error) {
	files, err := collectFiles(paths, cfg.Exclude)
	if err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	byLang := groupByLanguage(files)

	// If a specific language is forced, only keep that language's files
	if cfg.Lang != "" {
		filtered := make(map[string][]string)
		for lang, langFiles := range byLang {
			if strings.EqualFold(lang, cfg.Lang) {
				filtered[lang] = langFiles
			}
		}
		byLang = filtered
	}

	var allScores []FunctionScore

	for lang, langFiles := range byLang {
		drv := findDriver(lang)
		if drv == nil {
			continue
		}

		covPath := cfg.CoveragePathForLang(lang)
		covMap, err := loadCoverage(covPath, lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load coverage for %s: %v\n", lang, err)
		}

		for _, filePath := range langFiles {
			scores, err := analyzeFile(drv, filePath, covMap)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: %v\n", err)
				continue
			}
			allScores = append(allScores, scores...)
		}
	}

	return allScores, nil
}

func collectFiles(paths []string, excludes []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}
		if info.IsDir() {
			err := filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: %v\n", err)
					return nil
				}
				if d.IsDir() {
					return nil
				}
				for _, pattern := range excludes {
					if matchesExclude(pattern, path) {
						return nil
					}
				}
				files = append(files, path)
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking %s: %w", p, err)
			}
		} else {
			files = append(files, p)
		}
	}
	return files, nil
}

func matchesExclude(pattern, path string) bool {
	matched, _ := filepath.Match(pattern, filepath.Base(path))
	if matched {
		return true
	}
	matched, _ = filepath.Match(pattern, path)
	if matched {
		return true
	}
	if strings.Contains(pattern, "**") {
		return matchGlobstar(pattern, path)
	}
	return false
}

func matchGlobstar(pattern, path string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) == 1 {
		return false
	}
	// Check prefix (before first **)
	if parts[0] != "" {
		prefix := strings.TrimSuffix(parts[0], "/")
		if !strings.HasPrefix(path, prefix) {
			return false
		}
		path = path[len(prefix):]
		path = strings.TrimPrefix(path, "/")
	}
	// Check suffix (after last **) with glob support
	last := parts[len(parts)-1]
	if last != "" {
		suffix := strings.TrimPrefix(last, "/")
		if strings.ContainsAny(suffix, "*?[") {
			for _, comp := range strings.Split(path, "/") {
				if ok, _ := filepath.Match(suffix, comp); ok {
					return true
				}
			}
			return false
		}
		if strings.Contains(path, suffix) {
			return true
		}
		return false
	}
	// ** at end (suffix empty) — check middle parts exist in path
	// e.g., "**/vendor/**" → parts=["", "/vendor/", ""]
	// Each middle segment must appear as a path component, in order.
	if len(parts) > 2 {
		remaining := path
		segs := parts[1 : len(parts)-1]
		for i, seg := range segs {
			seg = strings.Trim(seg, "/")
			if seg == "" {
				continue
			}
			last := i == len(segs)-1
			// Check as path component with / on both sides
			idx := strings.Index(remaining, "/"+seg+"/")
			// Check as final path component (e.g., path ending with /vendor)
			if idx < 0 && last {
				if strings.HasSuffix(remaining, "/"+seg) {
					return true
				}
				return false
			}
			if idx < 0 {
				return false
			}
			remaining = remaining[idx+len(seg)+1:]
		}
		return true
	}
	// Pattern like "src/**" with only prefix and trailing ** — prefix already checked
	return true
}

func groupByLanguage(files []string) map[string][]string {
	byLang := make(map[string][]string)
	for _, f := range files {
		lang := detectLanguage(f)
		if lang != "" {
			byLang[lang] = append(byLang[lang], f)
		}
	}
	return byLang
}

func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".py":
		return "python"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".go":
		return "go"
	default:
		return ""
	}
}

func findDriver(lang string) driver.Driver {
	for _, d := range drivers {
		if strings.EqualFold(d.Name(), lang) {
			return d
		}
	}
	return nil
}

func loadCoverage(path, lang string) (coverage.CoverageMap, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	switch lang {
	case "python":
		return coverage.ParsePythonCoverage(path)
	case "javascript", "typescript":
		return coverage.ParseLCOV(path)
	case "go":
		return coverage.ParseGoCover(path)
	default:
		return nil, fmt.Errorf("unknown coverage format for language %s", lang)
	}
}

func analyzeFile(drv driver.Driver, filePath string, covMap coverage.CoverageMap) ([]FunctionScore, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	funcs, err := drv.FindFunctions(source, filePath)
	if err != nil {
		return nil, fmt.Errorf("finding functions in %s: %w", filePath, err)
	}

	var scores []FunctionScore
	for _, fn := range funcs {
		cc, err := drv.CalcComplexity(source, fn)
		if err != nil {
			return nil, fmt.Errorf("calculating CC for %s in %s: %w", fn.Name, filePath, err)
		}

		coveragePct := 0.0
		if covMap != nil {
			covStart := fn.CoverageStartLine
			if covStart == 0 {
				covStart = fn.StartLine
			}
			coveragePct = computeCoverage(covMap, filePath, covStart, fn.EndLine, source)
		}

		crap := calculator.CRAP(cc, coveragePct)

		scores = append(scores, FunctionScore{
			Name:            fn.Name,
			File:            filePath,
			StartLine:       fn.StartLine,
			EndLine:         fn.EndLine,
			CC:              cc,
			CoveragePercent: coveragePct,
			CRAP:            crap,
		})
	}

	return scores, nil
}

func countExecutableLines(source []byte, startLine, endLine int) int {
	lines := strings.Split(string(source), "\n")
	count := 0
	for ln := startLine; ln <= endLine; ln++ {
		if ln < 1 || ln > len(lines) {
			continue
		}
		stripped := strings.TrimSpace(lines[ln-1])
		if stripped == "" {
			continue
		}
		if strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, "//") {
			continue
		}
		if strings.HasPrefix(stripped, "/*") || strings.HasPrefix(stripped, "*/") {
			continue
		}
		count++
	}
	return count
}

func computeCoverage(covMap coverage.CoverageMap, filePath string, startLine, endLine int, source []byte) float64 {
	data, ok := covMap[filePath]
	if !ok {
		base := filepath.Base(filePath)
		data, ok = covMap[base]
	}
	if !ok || data == nil {
		// Try suffix matching: coverage keys are often relative paths ("src/pkg/file.py")
		// but WalkDir produces absolute paths ("/project/src/pkg/file.py").
		for covKey, covData := range covMap {
			// Only match if the coverage key is a path-suffix at a path boundary.
			// e.g., "/project/src/pkg/file.py" ends with "/src/pkg/file.py" ✓
			// but "/project/old_cli.py" must NOT match covKey "cli.py" (handled by basename above).
			if strings.HasSuffix(filePath, "/"+covKey) {
				data = covData
				ok = true
				break
			}
		}
	}
	if !ok || data == nil {
		return 0.0
	}

	totalLines := countExecutableLines(source, startLine, endLine)
	if totalLines <= 0 {
		totalLines = 1
	}

	covered := 0
	for ln := startLine; ln <= endLine; ln++ {
		if data.CoveredLines[ln] {
			covered++
		}
	}

	return (math.Min(float64(covered)/float64(totalLines), 1.0)) * 100.0
}
