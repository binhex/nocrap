package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseGoRange(rangeStr string) (fileKey string, startLine, endLine int, ok bool) {
	parts := strings.SplitN(rangeStr, ":", 2)
	if len(parts) != 2 {
		return "", 0, 0, false
	}
	fileKey = parts[0]
	rangeDetails := strings.SplitN(parts[1], ",", 2)
	if len(rangeDetails) != 2 {
		return "", 0, 0, false
	}
	startParts := strings.SplitN(rangeDetails[0], ".", 2)
	endParts := strings.SplitN(rangeDetails[1], ".", 2)
	if len(startParts) != 2 || len(endParts) != 2 {
		return "", 0, 0, false
	}
	var err1, err2 error
	startLine, err1 = strconv.Atoi(startParts[0])
	endLine, err2 = strconv.Atoi(endParts[0])
	if err1 != nil || err2 != nil {
		return "", 0, 0, false
	}
	return fileKey, startLine, endLine, true
}

func parseGoCoverLine(line string) (bareFile, fileKey string, startLine, endLine, count int, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "mode:") {
		return "", "", 0, 0, 0, false
	}
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return "", "", 0, 0, 0, false
	}
	fileKey, startLine, endLine, ok = parseGoRange(parts[0])
	if !ok {
		return "", "", 0, 0, 0, false
	}
	var err3 error
	count, err3 = strconv.Atoi(parts[2])
	if err3 != nil {
		return "", "", 0, 0, 0, false
	}
	bareFile = filepath.Base(fileKey)
	return bareFile, fileKey, startLine, endLine, count, true
}

func ensureCoverageData(result CoverageMap, keys []string) {
	for _, key := range keys {
		if _, exists := result[key]; !exists {
			result[key] = &CoverageData{CoveredLines: make(map[int]bool)}
		}
	}
}

func applyCoverageRange(result CoverageMap, keys []string, startLine, endLine, count int) {
	for ln := startLine; ln <= endLine; ln++ {
		for _, key := range keys {
			if count > 0 {
				result[key].CoveredLines[ln] = true
			}
			result[key].TotalLines++
		}
	}
}

// ParseGoCover reads a Go cover profile and returns a CoverageMap.
// Format: <module>/<package>/<file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
func ParseGoCover(path string) (CoverageMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening cover file %s: %w", path, err)
	}
	defer f.Close()

	result := make(CoverageMap)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		bareFile, fileKey, startLine, endLine, count, ok := parseGoCoverLine(scanner.Text())
		if !ok {
			continue
		}
		keys := []string{bareFile, fileKey}
		ensureCoverageData(result, keys)
		applyCoverageRange(result, keys, startLine, endLine, count)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading cover file %s: %w", path, err)
	}
	return result, nil
}
