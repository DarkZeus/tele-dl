package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Config holds all application configuration
type Config struct {
	URL                string
	Workers            int
	Timeout            time.Duration
	TelegraphAPIBase   string
	TelegraphFileBase  string
	OutputDir          string
	Progress           bool
	Quiet              bool
	Retries            int
	JSONOutput         bool
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Workers:           50,
		Timeout:           30 * time.Second,
		TelegraphAPIBase:  "https://api.telegra.ph/getPage/",
		TelegraphFileBase: "https://telegra.ph/file/",
		OutputDir:         ".",
		Progress:          true,
		Quiet:             false,
		Retries:           3,
		JSONOutput:        false,
	}
}

// FromCobraCommand creates a config from Cobra command flags
func FromCobraCommand(cmd *cobra.Command) (*Config, error) {
	cfg := DefaultConfig()
	
	// Get flag values
	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return nil, err
	}
	if link == "" {
		return nil, fmt.Errorf("telegraph URL is required")
	}
	
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		return nil, err
	}
	
	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		return nil, err
	}
	if timeout == 0 {
		timeout = cfg.Timeout // Use default
	}
	
	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}
	
	progress, err := cmd.Flags().GetBool("progress")
	if err != nil {
		return nil, err
	}
	
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return nil, err
	}
	
	retries, err := cmd.Flags().GetInt("retries")
	if err != nil {
		return nil, err
	}
	
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return nil, err
	}
	
	// Override quiet mode if progress is disabled
	if quiet {
		progress = false
	}
	
	// Populate config
	cfg.URL = link
	cfg.Workers = workers
	cfg.Timeout = timeout
	cfg.OutputDir = outputDir
	cfg.Progress = progress
	cfg.Quiet = quiet
	cfg.Retries = retries
	cfg.JSONOutput = jsonOutput
	
	return cfg, nil
}

 