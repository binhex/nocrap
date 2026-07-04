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

		// Header line: "-:    0:Source:/path/to/file.c"
		if strings.Contains(line, ":Source:") {
			parts := strings.SplitN(line, ":Source:", 2)
			if len(parts) == 2 {
				sourcePath = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Split on first colon to get count: "#####:    3:    return -1;"
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		countStr := strings.TrimSpace(line[:idx])
		rest := line[idx+1:]

		// Extract line number: "    3:    return -1;"
		idx2 := strings.Index(rest, ":")
		if idx2 < 0 {
			continue
		}
		lineNoStr := strings.TrimSpace(rest[:idx2])

		lineNo, err := strconv.Atoi(lineNoStr)
		if err != nil {
			continue
		}

		if countStr == "-" {
			// Non-executable line — skip
			continue
		}

		total++
		if countStr != "#####" {
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
