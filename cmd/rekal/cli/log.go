package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show recent checkpoints",
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			_ = limit // will be used when implemented
			fmt.Fprintln(cmd.ErrOrStderr(), "rekal log: not yet implemented")
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Max entries to show")
	return cmd
}
