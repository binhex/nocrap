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

	sourcePath, covered, total, err := processGcovLines(bufio.NewScanner(f))
	if err != nil {
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

// processGcovLines scans all gcov lines and returns source path, covered lines, total lines.
func processGcovLines(scanner *bufio.Scanner) (string, map[int]bool, int, error) {
	var sourcePath string
	covered := make(map[int]bool)
	total := 0

	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " \t")

		if strings.Contains(line, ":Source:") {
			parts := strings.SplitN(line, ":Source:", 2)
			if len(parts) == 2 {
				sourcePath = strings.TrimSpace(parts[1])
			}
			continue
		}

		lineNo, count, ok := parseGcovLine(line)
		if !ok {
			continue
		}
		if lineNo == 0 {
			continue
		}

		total++
		if isCovered(count) {
			covered[lineNo] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, 0, err
	}
	return sourcePath, covered, total, nil
}

// parseGcovLine parses a gcov data line and returns the line number and count.
// Returns (0, "", false) if the line cannot be parsed or is non-executable (-).
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

	if countStr == "-" {
		return 0, "", false
	}
	return lineNo, countStr, true
}

func isCovered(count string) bool {
	return count != "#####" && count != "0"
}
