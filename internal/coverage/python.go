package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type pythonCoverageFile struct {
	Meta  map[string]any                       `json:"meta"`
	Files map[string]pythonCoverageFileDetails `json:"files"`
}

type pythonCoverageFileDetails struct {
	ExecutedLines []int `json:"executed_lines"`
	MissingLines  []int `json:"missing_lines"`
}

// ParseDotCoverage reads a .coverage SQLite database (coverage.py's native format)
// using a Python subprocess and returns a CoverageMap with line-level coverage data.
// This matches what pytest-crap gets from data.lines() and is preferred over
// ParsePythonCoverage (which uses statement-level data from coverage.json).
func ParseDotCoverage(path string) (CoverageMap, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf(".coverage file not found: %s", path)
	}

	cmd := exec.Command("python3", "-c", pythonDotCoverageScript, path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running coverage subprocess: %w", err)
	}

	var response struct {
		Root  string           `json:"root"`
		Files map[string][]int `json:"files"`
		Error string           `json:"error,omitempty"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("parsing coverage subprocess output: %w", err)
	}
	if response.Error != "" {
		return nil, fmt.Errorf("coverage subprocess error: %s", response.Error)
	}

	result := make(CoverageMap, len(response.Files))
	for filePath, lines := range response.Files {
		covered := make(map[int]bool, len(lines))
		for _, ln := range lines {
			covered[ln] = true
		}
		result[filePath] = &CoverageData{
			CoveredLines: covered,
			TotalLines:   0, // unknown: .coverage data has no total-line count (use countExecutableLines)
		}
	}
	return result, nil
}

func ParsePythonCoverage(path string) (CoverageMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading coverage file %s: %w", path, err)
	}
	var report pythonCoverageFile
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing coverage JSON %s: %w", path, err)
	}
	result := make(CoverageMap, len(report.Files))
	for filePath, details := range report.Files {
		covered := make(map[int]bool, len(details.ExecutedLines))
		for _, ln := range details.ExecutedLines {
			covered[ln] = true
		}
		result[filePath] = &CoverageData{
			CoveredLines: covered,
			TotalLines:   len(details.ExecutedLines) + len(details.MissingLines),
		}
	}
	return result, nil
}

var pythonDotCoverageScript = `import json, os, sys
try:
    from coverage import Coverage
except ImportError:
    print(json.dumps({"error": "coverage module not available"}))
    sys.exit(1)

data_file = sys.argv[1]
root = os.path.dirname(os.path.abspath(data_file))

cov = Coverage(data_file=data_file)
cov.load()
data = cov.get_data()

result = {}
for f in data.measured_files():
    if not f.endswith(".py"):
        continue
    lines = sorted(data.lines(f) or [])
    relpath = os.path.relpath(f, root)
    result[relpath] = lines

print(json.dumps({"root": root, "files": result}))
`
