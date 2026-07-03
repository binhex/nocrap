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
	if got := cfg.CoveragePathForLang("python"); got != ".coverage.json" {
		t.Errorf("python = %q, want .coverage.json", got)
	}
	if got := cfg.CoveragePathForLang("javascript"); got != "coverage/lcov.info" {
		t.Errorf("javascript = %q, want coverage/lcov.info", got)
	}
	if got := cfg.CoveragePathForLang("go"); got != "cover.out" {
		t.Errorf("go = %q, want cover.out", got)
	}
}
