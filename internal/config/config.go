// Package config handles .crap.toml parsing, environment variables, and CLI flag merging.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config holds all configuration for a nocrap run.
type Config struct {
	Threshold float64        `toml:"threshold"`
	TopN      int            `toml:"top_n"`
	Exclude   []string       `toml:"exclude"`
	Lang      string         `toml:"lang"`
	Coverage  CoverageConfig `toml:"coverage"`
}

// CoverageConfig holds coverage file paths per language.
type CoverageConfig struct {
	Python     string `toml:"python"`
	JavaScript string `toml:"javascript"`
	Go         string `toml:"go"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Threshold: 30,
		TopN:      20,
		Exclude:   []string{},
		Coverage: CoverageConfig{
			Python:     ".coverage.json",
			JavaScript: "coverage/lcov.info",
			Go:         "cover.out",
		},
	}
}

// LoadConfig loads configuration from .crap.toml if it exists, then applies
// environment variable overrides.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = ".crap.toml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnv(cfg) // env vars apply even without config file
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	applyEnv(cfg)
	return cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("CRAP_COVERAGE_PYTHON"); v != "" {
		cfg.Coverage.Python = v
	}
	if v := os.Getenv("CRAP_COVERAGE_JAVASCRIPT"); v != "" {
		cfg.Coverage.JavaScript = v
	}
	if v := os.Getenv("CRAP_COVERAGE_GO"); v != "" {
		cfg.Coverage.Go = v
	}
}

// MergeFlags applies CLI flag overrides to the config. Zero values for
// threshold and topN are treated as "not set" to allow explicit zero.
// Use negative sentinel (-1) internally to suppress a flag.
func MergeFlags(cfg *Config, threshold float64, topN int, lang string, excludes []string) *Config {
	// threshold uses negative sentinel to allow 0
	if threshold >= 0 {
		cfg.Threshold = threshold
	}
	// topN uses negative sentinel to allow 0
	if topN >= 0 {
		cfg.TopN = topN
	}
	if lang != "" {
		cfg.Lang = lang
	}
	if len(excludes) > 0 {
		cfg.Exclude = append(cfg.Exclude, excludes...)
	}
	return cfg
}

// CoveragePathForLang returns the coverage file path for a given language.
func (c *Config) CoveragePathForLang(lang string) string {
	switch strings.ToLower(lang) {
	case "python":
		return c.Coverage.Python
	case "javascript", "typescript":
		return c.Coverage.JavaScript
	case "go":
		return c.Coverage.Go
	default:
		return ""
	}
}
