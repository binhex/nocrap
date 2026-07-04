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
	cDriver "nocrap/internal/driver/c"
	cppDriver "nocrap/internal/driver/cpp"
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
	cDriver.New(),
	cppDriver.New(),
}

func processLang(lang string, langFiles []string, drv driver.Driver, covMap coverage.CoverageMap) []FunctionScore {
	var scores []FunctionScore
	for _, filePath := range langFiles {
		fileScores, err := analyzeFile(drv, filePath, covMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			continue
		}
		scores = append(scores, fileScores...)
	}
	return scores
}

func filterByLang(byLang map[string][]string, targetLang string) map[string][]string {
	if targetLang == "" {
		return byLang
	}
	filtered := make(map[string][]string)
	for lang, langFiles := range byLang {
		if strings.EqualFold(lang, targetLang) {
			filtered[lang] = langFiles
		}
	}
	return filtered
}

func Analyze(paths []string, cfg *config.Config) ([]FunctionScore, error) {
	files, err := collectFiles(paths, cfg.Exclude)
	if err != nil {
		return nil, fmt.Errorf("collecting files: %w", err)
	}

	byLang := filterByLang(groupByLanguage(files), cfg.Lang)
	sourceDirs := computeSourceDirs(paths)
	var allScores []FunctionScore

	for lang, langFiles := range byLang {
		drv := findDriver(lang)
		if drv == nil {
			continue
		}
		covPath := cfg.CoveragePathForLang(lang)
		covMap, err := loadCoverage(covPath, lang, sourceDirs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load coverage for %s: %v\n", lang, err)
		}
		allScores = append(allScores, processLang(lang, langFiles, drv, covMap)...)
	}

	return allScores, nil
}

func walkDirForFiles(path string, excludes []string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		for _, pattern := range excludes {
			if matchesExclude(pattern, p) {
				return nil
			}
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", path, err)
	}
	return files, nil
}

func collectFiles(paths []string, excludes []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}
		if info.IsDir() {
			dirFiles, err := walkDirForFiles(p, excludes)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
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

func matchGlobstarSuffix(path, suffix string) bool {
	if strings.ContainsAny(suffix, "*?[") {
		for _, comp := range strings.Split(path, "/") {
			if ok, _ := filepath.Match(suffix, comp); ok {
				return true
			}
		}
		return false
	}
	return strings.Contains(path, suffix)
}

func matchGlobstarMiddle(path string, segs []string) bool {
	remaining := path
	for i, seg := range segs {
		seg = strings.Trim(seg, "/")
		if seg == "" {
			continue
		}
		last := i == len(segs)-1
		if idx := strings.Index(remaining, "/"+seg+"/"); idx >= 0 {
			remaining = remaining[idx+len(seg)+1:]
			continue
		}
		if last && strings.HasSuffix(remaining, "/"+seg) {
			return true
		}
		return false
	}
	return true
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
	// Check suffix (after last **)
	last := parts[len(parts)-1]
	if last != "" {
		return matchGlobstarSuffix(path, strings.TrimPrefix(last, "/"))
	}
	// ** at end — check middle parts exist in path, in order
	if len(parts) > 2 {
		return matchGlobstarMiddle(path, parts[1:len(parts)-1])
	}
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
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx":
		return "cpp"
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

// computeSourceDirs extracts directory paths from the user-provided source paths.
// These are used to search for coverage files when the coverage path is relative.
func computeSourceDirs(paths []string) []string {
	dirs := make([]string, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirs = append(dirs, p)
		} else {
			dirs = append(dirs, filepath.Dir(p))
		}
	}
	return dirs
}

func loadCoverage(path, lang string, sourceDirs []string) (coverage.CoverageMap, error) {
	if path == "" {
		return nil, nil
	}

	// Try the given path directly (CWD-relative, matches current behavior)
	if _, err := os.Stat(path); err == nil {
		return parseCoverageByLang(path, lang)
	}

	// If the path is relative and not found, search each source directory
	// and its parents for the coverage file. This handles the common case
	// where the user runs nocrap from outside the project directory:
	//   $ ./nocrap ../boozarr/src   # coverage is at ../boozarr/coverage.json
	if !filepath.IsAbs(path) {
		for _, dir := range sourceDirs {
			// Walk up the directory tree from the source dir looking for the coverage file.
			// Stop before we reach the root (where filepath.Dir(dir) == dir).
			for cur := filepath.Clean(dir); ; cur = filepath.Dir(cur) {
				candidate := filepath.Join(cur, path)
				if _, err := os.Stat(candidate); err == nil {
					return parseCoverageByLang(candidate, lang)
				}
				// Stop if we can't go up any further (reached root or filesystem boundary)
				parent := filepath.Dir(cur)
				if parent == cur {
					break
				}
			}
		}
	}

	return nil, nil
}

func parseCoverageByLang(path, lang string) (coverage.CoverageMap, error) {
	switch lang {
	case "python":
		// Prefer .coverage SQLite database (line-level coverage via data.lines()).
		// This matches what pytest-crap uses internally and gives correct coverage
		// percentages for multi-line statements. Fall back to coverage.json if
		// .coverage is not available or parsing fails.
		dotCovPath := filepath.Join(filepath.Dir(path), ".coverage")
		if _, err := os.Stat(dotCovPath); err == nil {
			if covMap, err := coverage.ParseDotCoverage(dotCovPath); err == nil {
				return covMap, nil
			} else {
				fmt.Fprintf(os.Stderr, "warning: .coverage parse failed (%s), falling back to coverage.json: %v\n", dotCovPath, err)
			}
		}
		return coverage.ParsePythonCoverage(path)
	case "javascript", "typescript":
		return coverage.ParseLCOV(path)
	case "go":
		return coverage.ParseGoCover(path)
	case "c", "cpp":
		return coverage.ParseGcov(path)
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

func isCodeLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	// Braces, semicolons, and other single-character structural tokens are not executable.
	if line == "{" || line == "}" {
		return false
	}
	if line[0] == '#' {
		return false
	}
	if len(line) >= 2 {
		switch line[:2] {
		case "//", "/*", "*/":
			return false
		}
	}
	return true
}

func countExecutableLines(source []byte, startLine, endLine int) int {
	lines := strings.Split(string(source), "\n")
	count := 0
	for ln := startLine; ln <= endLine; ln++ {
		if ln < 1 || ln > len(lines) {
			continue
		}
		if isCodeLine(lines[ln-1]) {
			count++
		}
	}
	return count
}

func findCoverageData(covMap coverage.CoverageMap, filePath string) *coverage.CoverageData {
	if data, ok := covMap[filePath]; ok && data != nil {
		return data
	}
	if data, ok := covMap[filepath.Base(filePath)]; ok && data != nil {
		return data
	}
	for covKey, covData := range covMap {
		if strings.HasSuffix(filePath, "/"+covKey) && covData != nil {
			return covData
		}
	}
	return nil
}

func computeCoverage(covMap coverage.CoverageMap, filePath string, startLine, endLine int, source []byte) float64 {
	data := findCoverageData(covMap, filePath)
	if data == nil {
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
