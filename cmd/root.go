// Package cmd contains the CLI commands for nocrap.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nocrap/internal/config"
	"nocrap/internal/engine"
	"nocrap/internal/reporter"
)

var (
	cfgPath   string
	lang      string
	threshold float64
	topN      int
	jsonOut   bool
	excludes  []string

	rootCmd = &cobra.Command{
		Use:   "nocrap [flags] <path...>",
		Short: "Calculate CRAP scores for source code",
		Long: `nocrap calculates Change Risk Anti-Patterns (CRAP) scores for Python,
JavaScript, TypeScript, and Go source code using pre-generated coverage data.`,
		Args: cobra.MinimumNArgs(1),
		RunE: run,
	}
)

func init() {
	rootCmd.Flags().StringVar(&cfgPath, "config", "", "Path to config file (default: .crap.toml)")
	rootCmd.Flags().StringVar(&lang, "lang", "", "Force language (python, javascript, typescript, go)")
	rootCmd.Flags().Float64Var(&threshold, "threshold", -1, "CRAP threshold for highlighting (default: 30; use 0 for none)")
	rootCmd.Flags().IntVar(&topN, "top-n", -1, "Number of items per table. 0 = show all (default: 20)")
	rootCmd.Flags().BoolVar(&jsonOut, "json", false, "Output machine-readable JSON instead of tables")
	rootCmd.Flags().StringArrayVar(&excludes, "exclude", nil, "Glob patterns to exclude (repeatable)")
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg = config.MergeFlags(cfg, threshold, topN, lang, excludes)

	scores, err := engine.Analyze(args, cfg)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if len(scores) == 0 {
		fmt.Fprintln(os.Stderr, "No functions found in the specified paths.")
		return nil
	}

	if jsonOut {
		return reporter.WriteJSON(scores, os.Stdout)
	}

	wd, _ := os.Getwd()
	r := reporter.New(wd)
	r.RenderFunctionTable(scores, cfg.TopN)
	r.RenderFileSummary(scores, cfg.TopN, cfg.Threshold)
	r.RenderFolderSummary(scores, cfg.TopN, cfg.Threshold)

	return nil
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
