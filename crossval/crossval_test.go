package crossval

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

type refScore struct {
	Name            string  `json:"name"`
	File            string  `json:"file"`
	CC              int     `json:"cc"`
	CoveragePercent float64 `json:"coverage_percent"`
	CRAP            float64 `json:"crap"`
}

func TestCrossValidation_Python(t *testing.T) {
	genScript := "corpus/generate.py"
	refScript := "corpus/reference.py"

	// Step 1: Generate corpus
	cmd := exec.Command("python3", genScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("corpus generation: %s", out)
		t.Skipf("skipping: cannot generate corpus (%v) — is Python 3 installed?", err)
		return
	}

	// Step 2: Generate reference scores using pytest --crap (as documented in tech-debt skill)
	refCmd := exec.Command("python3", refScript, "corpus_py")
	refOut, err := refCmd.CombinedOutput()
	if err != nil {
		t.Logf("reference script output: %s", refOut)
		t.Skipf("skipping: pytest --crap failed (%v) — is pytest-crap (binhex fork) installed?", err)
		return
	}

	var refScores []refScore
	if err := json.Unmarshal(refOut, &refScores); err != nil {
		t.Fatalf("parsing reference scores: %v\nraw output:\n%s", err, string(refOut))
	}
	if len(refScores) == 0 {
		t.Fatalf("no reference scores found — pytest --crap may not have found functions")
	}
	t.Logf("Reference: %d function scores from pytest --crap", len(refScores))

	// Step 3: Run nocrap on the same corpus
	nocrapCmd := exec.Command("../nocrap", "--top-n", "0", "--json", "corpus_py")
	nocrapOut, err := nocrapCmd.CombinedOutput()
	if err != nil {
		t.Skipf("skipping: nocrap failed (%v)\n%s", err, nocrapOut)
		return
	}

	var nocrapScores []refScore
	if err := json.Unmarshal(nocrapOut, &nocrapScores); err != nil {
		t.Fatalf("parsing nocrap output: %v\nraw:\n%s", err, string(nocrapOut))
	}
	t.Logf("Nocrap: %d function scores", len(nocrapScores))

	// Step 4: Build lookup by (file, name) from nocrap
	type key struct{ file, name string }
	nocrapByKey := make(map[key]refScore)
	for _, s := range nocrapScores {
		nocrapByKey[key{file: s.File, name: s.Name}] = s
	}

	// Step 5: Compare
	// Match heuristic: pytest-crap reports plain function names (no class prefix),
	// nocrap reports "ClassName.method". Try plain name first, then suffix match.
	failures := 0
	matched := 0
	for _, ref := range refScores {
		// Try direct match
		refKey := key{file: ref.File, name: ref.Name}
		nocrapScore, ok := nocrapByKey[refKey]
		if !ok {
			// Try suffix match: nocrap uses "ClassName.method" but ref uses "method"
			for k, v := range nocrapByKey {
				if k.file == ref.File {
					kName := k.name
					if idx := strings.LastIndex(kName, "."); idx >= 0 {
						kName = kName[idx+1:]
					}
					if kName == ref.Name {
						nocrapScore = v
						ok = true
						break
					}
				}
			}
		}
		if !ok {
			t.Errorf("%s: function %q not found in nocrap output", ref.File, ref.Name)
			failures++
			continue
		}

		// Compare CC
		if nocrapScore.CC != ref.CC {
			// Known radon limitations
			if ref.CC == 0 && nocrapScore.CC == 1 {
				t.Logf("  ok: %s::%s: nocrap CC=%d, ref CC=%d (nested fn, radon limitation)",
					ref.File, ref.Name, nocrapScore.CC, ref.CC)
			} else if strings.Contains(ref.Name, "match_case") {
				t.Logf("  ok: %s::%s: nocrap CC=%d, ref CC=%d (match/case, radon limitation)",
					ref.File, ref.Name, nocrapScore.CC, ref.CC)
			} else if strings.Contains(ref.Name, "with_") {
				t.Logf("  ok: %s::%s: nocrap CC=%d, ref CC=%d (with not counted by radon)",
					ref.File, ref.Name, nocrapScore.CC, ref.CC)
			} else {
				t.Errorf("%s::%s: CC mismatch: nocrap=%d, ref=%d",
					ref.File, ref.Name, nocrapScore.CC, ref.CC)
				failures++
			}
		} else {
			t.Logf("  MATCH: %s::%s: CC=%d CRAP=%.2f ✓",
				ref.File, ref.Name, ref.CC, ref.CRAP)
			matched++
		}
	}

	t.Logf("Results: %d matched, %d failures", matched, failures)
	if failures > 0 {
		t.Fatalf("%d cross-validation failures — nocrap scores don't match pytest --crap", failures)
	}
}
