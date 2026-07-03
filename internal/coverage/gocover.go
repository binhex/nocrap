package coverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		rangeParts := strings.SplitN(parts[0], ":", 2)
		if len(rangeParts) != 2 {
			continue
		}
		fileKey := rangeParts[0]

		rangeDetails := strings.SplitN(rangeParts[1], ",", 2)
		if len(rangeDetails) != 2 {
			continue
		}
		startParts := strings.SplitN(rangeDetails[0], ".", 2)
		endParts := strings.SplitN(rangeDetails[1], ".", 2)
		if len(startParts) != 2 || len(endParts) != 2 {
			continue
		}
		startLine, err1 := strconv.Atoi(startParts[0])
		endLine, err2 := strconv.Atoi(endParts[0])
		if err1 != nil || err2 != nil {
			continue
		}

		count, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		bareFile := filepath.Base(fileKey)

		if _, exists := result[bareFile]; !exists {
			result[bareFile] = &CoverageData{
				CoveredLines: make(map[int]bool),
				TotalLines:   0,
			}
		}
		if _, exists := result[fileKey]; !exists {
			result[fileKey] = &CoverageData{
				CoveredLines: make(map[int]bool),
				TotalLines:   0,
			}
		}

		for ln := startLine; ln <= endLine; ln++ {
			if count > 0 {
				result[bareFile].CoveredLines[ln] = true
				result[fileKey].CoveredLines[ln] = true
			}
			result[bareFile].TotalLines++
			result[fileKey].TotalLines++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading cover file %s: %w", path, err)
	}
	return result, nil
}
