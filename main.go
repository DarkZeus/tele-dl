package main

import (
	"fmt"
	"os"

	"tele-dl/internal/app"

	"github.com/spf13/cobra"
)

var (
	version = "2.0.0"
	commit  = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "tele-dl",
		Short:   "Download media from Telegraph pages",
		Long: `tele-dl is a high-performance tool for downloading images and videos from Telegraph (telegra.ph) pages.

It supports concurrent downloads, progress tracking, and handles both Telegraph-hosted 
and external media with automatic retry logic.`,
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
		RunE:    app.RunDownload,
	}

	// Add flags
	flags := rootCmd.Flags()
	flags.StringP("link", "l", "", "Telegraph page URL (required)")
	flags.StringP("output", "o", ".", "Output directory")
	flags.IntP("workers", "w", 50, "Number of concurrent downloads")
	flags.DurationP("timeout", "t", 0, "HTTP request timeout (0 = 30s default)")
	flags.BoolP("progress", "p", true, "Show progress bar")
	flags.BoolP("quiet", "q", false, "Quiet mode (no progress, minimal output)")
	flags.Int("retries", 3, "Number of retry attempts for failed downloads")
	flags.Bool("json", false, "Output results in JSON format")

	// Mark required flags
	rootCmd.MarkFlagRequired("link")

	// Add completion command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(tele-dl completion bash)

Zsh:
  $ tele-dl completion zsh > "${fpath[1]}/_tele-dl"

Fish:
  $ tele-dl completion fish | source

PowerShell:
  PS> tele-dl completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
} 