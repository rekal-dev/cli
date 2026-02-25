package cli

import (
	"fmt"
	"os"

	"github.com/rekal-dev/cli/cmd/rekal/cli/versioncheck"
	"github.com/spf13/cobra"
)

// NewRootCmd returns the root command for the rekal CLI.
func NewRootCmd() *cobra.Command {
	var (
		fileFilter       string
		commitFilter     string
		checkpointFilter string
		authorFilter     string
		actorFilter      string
		limitFlag        int
	)

	cmd := &cobra.Command{
		Use:           "rekal [filters...] [query]",
		Short:         "Rekal — gives your agent precise memory",
		Long:          "Rekal gives your agent precise memory — the exact context it needs for what it's working on.",
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			versioncheck.CheckAndNotify(cmd.OutOrStdout(), Version)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no args and no filters, show help.
			if len(args) == 0 && fileFilter == "" && commitFilter == "" &&
				checkpointFilter == "" && authorFilter == "" && actorFilter == "" {
				return cmd.Help()
			}

			// Recall: preconditions required.
			cmd.SilenceUsage = true

			gitRoot, err := EnsureGitRoot()
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return NewSilentError(err)
			}
			if err := EnsureInitDone(gitRoot); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
				return NewSilentError(err)
			}

			// Suppress unused warnings — these will be used when recall is implemented.
			_ = fileFilter
			_ = commitFilter
			_ = checkpointFilter
			_ = authorFilter
			_ = actorFilter
			_ = limitFlag

			fmt.Fprintln(cmd.ErrOrStderr(), "rekal recall: not yet implemented")
			return nil
		},
	}

	// Recall filter flags on root command.
	cmd.Flags().StringVar(&fileFilter, "file", "", "Filter by file path (regex)")
	cmd.Flags().StringVar(&commitFilter, "commit", "", "Filter by git commit SHA")
	cmd.Flags().StringVar(&checkpointFilter, "checkpoint", "", "Query as of checkpoint ref")
	cmd.Flags().StringVar(&authorFilter, "author", "", "Filter by author email")
	cmd.Flags().StringVar(&actorFilter, "actor", "", "Filter by actor type (human|agent)")
	cmd.Flags().IntVarP(&limitFlag, "limit", "n", 0, "Max results (0 = no limit)")

	cmd.SetVersionTemplate("rekal {{.Version}}\n")
	cmd.Version = Version

	// Register all subcommands.
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCleanCmd())
	cmd.AddCommand(newCheckpointCmd())
	cmd.AddCommand(newPushCmd())
	cmd.AddCommand(newIndexCmd())
	cmd.AddCommand(newLogCmd())
	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newSyncCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "rekal", Version)
			return nil
		},
	}
}

// Run executes the root command and exits with the appropriate code.
func Run() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		if !IsSilentError(err) {
			fmt.Fprintln(rootCmd.ErrOrStderr(), err)
		}
		os.Exit(1)
	}
}
