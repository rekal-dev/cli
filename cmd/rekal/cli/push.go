package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push Rekal data to the remote branch",
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

			return runPush(cmd, gitRoot, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force push (overwrite remote with local data)")
	return cmd
}

func runPush(cmd *cobra.Command, gitRoot string, force bool) error {
	branch := rekalBranchName()

	// Check if local branch exists — if not, nothing to push.
	if err := exec.Command("git", "-C", gitRoot, "rev-parse", "--verify", branch).Run(); err != nil {
		return nil
	}

	// Compare local SHA vs remote tracking SHA — skip if identical.
	localSHA, err := exec.Command("git", "-C", gitRoot, "rev-parse", branch).Output()
	if err != nil {
		return nil
	}
	remoteSHA, err := exec.Command("git", "-C", gitRoot, "rev-parse", "origin/"+branch).Output()
	if err == nil && strings.TrimSpace(string(localSHA)) == strings.TrimSpace(string(remoteSHA)) {
		return nil // already up to date
	}

	if force {
		forceCmd := exec.Command("git", "-C", gitRoot, "push", "--no-verify", "--force", "origin", branch)
		forceCmd.Stdin = nil
		if err := forceCmd.Run(); err != nil {
			return nil // network or other error — fail silently
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "rekal: force pushed to origin/%s\n", branch)
		return nil
	}

	// Push with --no-verify to prevent recursive pre-push hook.
	pushCmd := exec.Command("git", "-C", gitRoot, "push", "--no-verify", "origin", branch)
	pushCmd.Stdin = nil // disconnect stdin so git doesn't hang in hook context
	output, err := pushCmd.CombinedOutput()
	if err != nil {
		if isNonFastForward(string(output)) {
			fmt.Fprintf(cmd.ErrOrStderr(), "rekal: push rejected (non-fast-forward) for origin/%s\n", branch)
			fmt.Fprintln(cmd.ErrOrStderr(), "rekal: your remote branch has diverged from local — review and run 'rekal push --force' to overwrite remote with local data")
			return nil
		}
		// Other errors (network, no remote) — return silently.
		return nil
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "rekal: pushed to origin/%s\n", branch)
	return nil
}

// isNonFastForward checks if git push output indicates a non-fast-forward rejection.
func isNonFastForward(output string) bool {
	return strings.Contains(output, "non-fast-forward") ||
		strings.Contains(output, "[rejected]") ||
		strings.Contains(output, "fetch first")
}
