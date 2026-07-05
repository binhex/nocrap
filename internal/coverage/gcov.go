package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseGcov reads a gcov .gcov text file and returns a CoverageMap.
func ParseGcov(path string) (CoverageMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening gcov file %s: %w", path, err)
	}
	defer f.Close()

	var sourcePath string
	covered := make(map[int]bool)
	total := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimLeft(line, " \t")

		// Header line
		if strings.Contains(line, ":Source:") {
			parts := strings.SplitN(line, ":Source:", 2)
			if len(parts) == 2 {
				sourcePath = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Parse execution count and line number
		lineNo, count, ok := parseGcovLine(line)
		if !ok || count == "-" {
			continue
		}
		if lineNo == 0 {
			continue
		}

		total++
		if count != "#####" && count != "0" {
			covered[lineNo] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading gcov file %s: %w", path, err)
	}

	if sourcePath == "" {
		return nil, fmt.Errorf("no Source: header found in %s", path)
	}

	return CoverageMap{
		sourcePath: &CoverageData{
			CoveredLines: covered,
			TotalLines:   total,
		},
	}, nil
}

// parseGcovLine parses a gcov data line and returns the line number and count.
// Returns (0, "", false) if the line cannot be parsed.
func parseGcovLine(line string) (int, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return 0, "", false
	}
	countStr := strings.TrimSpace(line[:idx])
	rest := line[idx+1:]

	idx2 := strings.Index(rest, ":")
	if idx2 < 0 {
		return 0, "", false
	}
	lineNoStr := strings.TrimSpace(rest[:idx2])

	lineNo, err := strconv.Atoi(lineNoStr)
	if err != nil {
		return 0, "", false
	}
	return lineNo, countStr, true
}
