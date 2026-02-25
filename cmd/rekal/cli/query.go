package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	var useIndex bool

	cmd := &cobra.Command{
		Use:   "query <sql>",
		Short: "Run raw SQL against the Rekal data model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			_ = useIndex // will be used when implemented
			_ = args[0]  // the SQL statement
			fmt.Fprintln(cmd.ErrOrStderr(), "rekal query: not yet implemented")
			return nil
		},
	}

	cmd.Flags().BoolVar(&useIndex, "index", false, "Run SQL against the index DB instead of the data DB")
	return cmd
}
