package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseLCOV reads an LCOV tracefile and returns a CoverageMap.
// Format: SF:<path>, DA:<line>,<hit>, LH:<lines_hit>, LF:<lines_found>, end_of_record
func ParseLCOV(path string) (CoverageMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening LCOV file %s: %w", path, err)
	}
	defer f.Close()

	result := make(CoverageMap)
	scanner := bufio.NewScanner(f)
	var currentFile string
	var coveredLines map[int]bool
	var totalLines int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "SF:"):
			currentFile = line[3:]
			coveredLines = make(map[int]bool)
			totalLines = 0
		case strings.HasPrefix(line, "DA:"):
			parts := strings.SplitN(line[3:], ",", 2)
			if len(parts) != 2 {
				continue
			}
			lineNo, err1 := strconv.Atoi(parts[0])
			count, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				continue
			}
			totalLines++
			if count > 0 {
				if coveredLines == nil {
					coveredLines = make(map[int]bool)
				}
				coveredLines[lineNo] = true
			}
		case line == "end_of_record":
			if currentFile != "" && coveredLines != nil {
				result[currentFile] = &CoverageData{
					CoveredLines: coveredLines,
					TotalLines:   totalLines,
				}
			}
			currentFile = ""
			coveredLines = nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading LCOV file %s: %w", path, err)
	}
	return result, nil
}
