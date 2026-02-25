package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rekal-dev/cli/cmd/rekal/cli/db"
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

			return runQuery(cmd, gitRoot, args[0], useIndex)
		},
	}

	cmd.Flags().BoolVar(&useIndex, "index", false, "Run SQL against the index DB instead of the data DB")
	return cmd
}

func runQuery(cmd *cobra.Command, gitRoot, query string, useIndex bool) error {
	// Read-only: only allow SELECT statements.
	normalized := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(normalized, "SELECT") {
		return fmt.Errorf("only SELECT statements are allowed")
	}

	var d *sql.DB
	var err error
	if useIndex {
		d, err = db.OpenIndex(gitRoot)
	} else {
		d, err = db.OpenData(gitRoot)
	}
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer d.Close()

	rows, err := d.Query(query)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("columns: %w", err)
	}

	out := cmd.OutOrStdout()
	first := true

	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			v := values[i]
			// Convert []byte to string for JSON output.
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			row[col] = v
		}

		data, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		if !first {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, string(data))
		first = false
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}

	// Trailing newline if we printed anything.
	if !first {
		fmt.Fprintln(out)
	}

	return nil
}
