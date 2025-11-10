package cmd

import (
	"errors"
	"fmt"
	"github.com/LukaSikic/subzy/runner"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
)

var opts = runner.Config{}

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run subzy",
	Aliases: []string{"r"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Input validation
		if opts.Target == "" && opts.Targets == "" {
			return fmt.Errorf("either --target or --targets must be specified")
		}

		if opts.Target != "" && opts.Targets != "" {
			return fmt.Errorf("cannot specify both --target and --targets")
		}

		if opts.Targets != "" {
			if _, err := os.Stat(opts.Targets); err != nil {
				return fmt.Errorf("targets file does not exist: %v", err)
			}
		}

		if opts.Concurrency <= 0 {
			return fmt.Errorf("concurrency must be greater than 0")
		}

		if opts.Timeout <= 0 {
			return fmt.Errorf("timeout must be greater than 0")
		}

		fingerprintsPath, err := runner.GetFingerprintPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(fingerprintsPath); errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("[ * ] Fingerprints not found; saving them to %q\n",
				fingerprintsPath)
			if err := runner.CheckFingerprints(); err != nil {
				return err
			}
		}

		if err := runner.Process(&opts); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	// Target configuration
	runCmd.Flags().StringVar(&opts.Target, "target", "", "Comma separated list of domains")
	runCmd.Flags().StringVar(&opts.Targets, "targets", "", "File containing the list of subdomains")
	runCmd.Flags().StringVar(&opts.Output, "output", "", "JSON output filename")

	// HTTP configuration
	runCmd.Flags().StringVar(&opts.UserAgent, "user-agent", "", "Custom User-Agent string (default: Subzy/1.1.0)")
	runCmd.Flags().BoolVar(&opts.HTTPS, "https", false, "Force https protocol if not no protocol defined for target (default false)")
	runCmd.Flags().BoolVar(&opts.VerifySSL, "verify_ssl", false, "If set to true it won't check sites with insecure SSL and return HTTP Error")
	runCmd.Flags().IntVar(&opts.Timeout, "timeout", 10, "Request timeout in seconds")

	// Output configuration
	runCmd.Flags().BoolVar(&opts.HideFails, "hide_fails", false, "Don't display failed results")
	runCmd.Flags().BoolVar(&opts.OnlyVuln, "vuln", false, "Save only vulnerable subdomains")

	// Performance configuration
	runCmd.Flags().IntVar(&opts.Concurrency, "concurrency", 10, "Number of concurrent checks")

	// Logging configuration
	runCmd.Flags().StringVar(&opts.LogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	runCmd.Flags().StringVar(&opts.LogFormat, "log-format", "console", "Log format: console, json")
	runCmd.Flags().StringVar(&opts.GraylogHost, "graylog-host", "", "Graylog server (e.g., graylog.example.com:12201)")
	runCmd.Flags().StringVar(&opts.GraylogApp, "graylog-app", "subzy", "Application name for Graylog")
	runCmd.Flags().BoolVar(&opts.LogToFile, "log-file", false, "Enable logging to file")
	runCmd.Flags().StringVar(&opts.LogFilePath, "log-file-path", "subzy.log", "Log file path")

	rootCmd.AddCommand(runCmd)
}
