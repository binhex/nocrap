package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/term"

	"nocrap/internal/engine"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

type Reporter struct {
	rootDir string
	width   int
}

func New(rootDir string) *Reporter {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	if width > 200 {
		width = 200
	}
	return &Reporter{rootDir: rootDir, width: width}
}

func (r *Reporter) RenderFunctionTable(scores []engine.FunctionScore, topN int) {
	r.renderFunctionTable(os.Stdout, scores, topN)
}

func (r *Reporter) renderFunctionTable(w io.Writer, scores []engine.FunctionScore, topN int) {
	sorted := make([]engine.FunctionScore, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CRAP > sorted[j].CRAP })

	if topN > 0 && topN < len(sorted) {
		sorted = sorted[:topN]
	}

	fmt.Fprintln(w, "\n── CRAP by Function ──")
	fmt.Fprintf(w, "%-10s %-5s %-9s %-30s %s\n", "CRAP", "CC", "Coverage", "Function", "File")
	fmt.Fprintln(w, strings.Repeat("─", r.width))

	for _, s := range sorted {
		cc := colorize(s.CRAP)
		coverageStr := fmt.Sprintf("%.1f%%", s.CoveragePercent)
		funcName := truncateRight(s.Name, 30)
		relPath := r.relativePath(s.File)
		fileDisplay := truncateMiddle(relPath, r.width-60)

		fmt.Fprintf(w, "%s%-8.2f%s  %-5d %-9s %-30s %s\n",
			cc, s.CRAP, colorReset, s.CC, coverageStr, funcName, fileDisplay)
	}
}

type fileSummary struct {
	file       string
	maxCRAP    float64
	countAbove int
}

func groupScores(scores []engine.FunctionScore, threshold float64, keyFn func(engine.FunctionScore) string) []*fileSummary {
	byGroup := make(map[string]*fileSummary)
	for _, s := range scores {
		key := keyFn(s)
		fs := byGroup[key]
		if fs == nil {
			fs = &fileSummary{file: key, maxCRAP: s.CRAP}
			byGroup[key] = fs
		}
		if s.CRAP > fs.maxCRAP {
			fs.maxCRAP = s.CRAP
		}
		if s.CRAP >= threshold {
			fs.countAbove++
		}
	}
	summaries := make([]*fileSummary, 0, len(byGroup))
	for _, fs := range byGroup {
		summaries = append(summaries, fs)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].maxCRAP > summaries[j].maxCRAP })
	return summaries
}

func renderGrouped(w io.Writer, label string, summaries []*fileSummary, topN int, r *Reporter) {
	if topN > 0 && topN < len(summaries) {
		summaries = summaries[:topN]
	}

	fmt.Fprintln(w, "\n── CRAP by "+label+" ──")
	fmt.Fprintf(w, "%-10s %-10s %s\n", "CRAP (max)", "#>=thr", label)
	fmt.Fprintln(w, strings.Repeat("─", r.width))

	for _, fs := range summaries {
		cc := colorize(fs.maxCRAP)
		fmt.Fprintf(w, "%s%-8.2f%s  %-10d %s\n",
			cc, fs.maxCRAP, colorReset, fs.countAbove, truncateMiddle(fs.file, r.width-25))
	}
}

func (r *Reporter) RenderFileSummary(scores []engine.FunctionScore, topN int, threshold float64) {
	summaries := groupScores(scores, threshold, func(s engine.FunctionScore) string { return s.File })
	renderGrouped(os.Stdout, "File", summaries, topN, r)
}

func (r *Reporter) RenderFolderSummary(scores []engine.FunctionScore, topN int, threshold float64) {
	summaries := groupScores(scores, threshold, func(s engine.FunctionScore) string { return filepath.Dir(s.File) })
	renderGrouped(os.Stdout, "Folder", summaries, topN, r)
}

func WriteJSON(scores []engine.FunctionScore, w io.Writer) error {
	type jsonScore struct {
		Name            string  `json:"name"`
		File            string  `json:"file"`
		StartLine       int     `json:"start_line"`
		EndLine         int     `json:"end_line"`
		CC              int     `json:"cc"`
		CoveragePercent float64 `json:"coverage_percent"`
		CRAP            float64 `json:"crap"`
	}
	output := make([]jsonScore, len(scores))
	for i, s := range scores {
		output[i] = jsonScore{
			Name:            s.Name,
			File:            s.File,
			StartLine:       s.StartLine,
			EndLine:         s.EndLine,
			CC:              s.CC,
			CoveragePercent: s.CoveragePercent,
			CRAP:            s.CRAP,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func colorize(crap float64) string {
	switch {
	case crap > 30:
		return colorRed
	case crap > 15:
		return colorYellow
	default:
		return colorGreen
	}
}

func (r *Reporter) relativePath(path string) string {
	if r.rootDir == "" {
		return path
	}
	absPath, _ := filepath.Abs(path)
	absRoot, _ := filepath.Abs(r.rootDir)
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return path
	}
	return rel
}

func truncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	available := maxLen - 3
	left := (available + 1) / 2
	right := available - left
	return s[:left] + "..." + s[len(s)-right:]
}

func truncateRight(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
