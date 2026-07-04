package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"nocrap/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Threshold != 30 {
		t.Errorf("default threshold = %f, want 30", cfg.Threshold)
	}
	if cfg.TopN != 20 {
		t.Errorf("default top_n = %d, want 20", cfg.TopN)
	}
}

func TestLoadConfig_File(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".crap.toml")
	content := `threshold = 9
top_n = 10
exclude = ["**/test_*"]

[coverage]
python = "custom_coverage.json"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Threshold != 9 {
		t.Errorf("threshold = %f, want 9", cfg.Threshold)
	}
	if cfg.Coverage.Python != "custom_coverage.json" {
		t.Errorf("coverage.python = %q, want custom_coverage.json", cfg.Coverage.Python)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "**/test_*" {
		t.Errorf("exclude = %v", cfg.Exclude)
	}
}

func TestCoveragePathForLang(t *testing.T) {
	cfg := config.DefaultConfig()
	if got := cfg.CoveragePathForLang("python"); got != "coverage.json" {
		t.Errorf("python = %q, want coverage.json", got)
	}
	if got := cfg.CoveragePathForLang("javascript"); got != "coverage/lcov.info" {
		t.Errorf("javascript = %q, want coverage/lcov.info", got)
	}
	if got := cfg.CoveragePathForLang("go"); got != "cover.out" {
		t.Errorf("go = %q, want cover.out", got)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	cfg, err := config.LoadConfig("/nonexistent/path/.crap.toml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Threshold != 30 {
		t.Errorf("expected default threshold 30, got %f", cfg.Threshold)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	t.Setenv("CRAP_COVERAGE_PYTHON", "custom.json")
	cfg, err := config.LoadConfig("/nonexistent/.crap.toml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Coverage.Python != "custom.json" {
		t.Errorf("expected custom.json from env, got %q", cfg.Coverage.Python)
	}
}

func TestMergeFlags_Threshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg = config.MergeFlags(cfg, 15.0, -1, "", nil)
	if cfg.Threshold != 15.0 {
		t.Errorf("threshold = %f, want 15", cfg.Threshold)
	}
}

func TestMergeFlags_TopN(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg = config.MergeFlags(cfg, -1.0, 5, "", nil)
	if cfg.TopN != 5 {
		t.Errorf("topN = %d, want 5", cfg.TopN)
	}
}

func TestMergeFlags_Lang(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg = config.MergeFlags(cfg, -1.0, -1, "python", nil)
	if cfg.Lang != "python" {
		t.Errorf("lang = %q, want python", cfg.Lang)
	}
}

func TestMergeFlags_Excludes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Exclude = []string{"**/test_*"}
	cfg = config.MergeFlags(cfg, -1.0, -1, "", []string{"**/vendor/**"})
	if len(cfg.Exclude) != 2 {
		t.Fatalf("expected 2 excludes, got %d", len(cfg.Exclude))
	}
	if cfg.Exclude[1] != "**/vendor/**" {
		t.Errorf("exclude[1] = %q", cfg.Exclude[1])
	}
}

func TestMergeFlags_DefaultSentinel(t *testing.T) {
	// Negative sentinel (-1) should not override defaults
	cfg := config.DefaultConfig()
	cfg = config.MergeFlags(cfg, -1.0, -1, "", nil)
	if cfg.Threshold != 30 {
		t.Errorf("threshold changed to %f, want 30", cfg.Threshold)
	}
	if cfg.TopN != 20 {
		t.Errorf("topN changed to %d, want 20", cfg.TopN)
	}
}
