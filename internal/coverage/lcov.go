package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseLCovDA(line string) (lineNo, count int, ok bool) {
	parts := strings.SplitN(line[3:], ",", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	ln, err1 := strconv.Atoi(parts[0])
	cnt, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return ln, cnt, true
}

func processLCovLine(line string, state *lcovState) {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, "SF:"):
		state.currentFile = line[3:]
		state.coveredLines = make(map[int]bool)
		state.totalLines = 0
	case strings.HasPrefix(line, "DA:"):
		if lineNo, count, ok := parseLCovDA(line); ok {
			state.totalLines++
			if count > 0 {
				state.coveredLines[lineNo] = true
			}
		}
	case line == "end_of_record":
		if state.currentFile != "" {
			state.result[state.currentFile] = &CoverageData{
				CoveredLines: state.coveredLines,
				TotalLines:   state.totalLines,
			}
		}
		state.currentFile = ""
		state.coveredLines = nil
	}
}

type lcovState struct {
	result       CoverageMap
	currentFile  string
	coveredLines map[int]bool
	totalLines   int
}

// ParseLCOV reads an LCOV tracefile and returns a CoverageMap.
// Format: SF:<path>, DA:<line>,<hit>, LH:<lines_hit>, LF:<lines_found>, end_of_record
func ParseLCOV(path string) (CoverageMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening LCOV file %s: %w", path, err)
	}
	defer f.Close()

	state := &lcovState{result: make(CoverageMap)}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		processLCovLine(scanner.Text(), state)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading LCOV file %s: %w", path, err)
	}
	return state.result, nil
}
