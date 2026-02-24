# Rekal CLI

Rekal gives your agent precise memory — the exact context it needs for what it's working on.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/rekal-dev/cli/main/scripts/install.sh | bash
```

Or with a specific version:

```bash
REKAL_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/rekal-dev/cli/main/scripts/install.sh | bash
```

Install location: `~/.local/bin` (override with `REKAL_INSTALL_DIR`).

## Usage

```bash
rekal version
```

*(M0: only version is implemented. More commands coming in later milestones.)*

## Development

Uses [mise](https://mise.jdx.dev/) for tools and tasks (same pattern as [Entire](https://github.com/entireio/cli)).

```bash
mise install          # install Go, golangci-lint
mise run fmt          # format code
mise run lint         # run linters
mise run test         # run tests
mise run test:ci      # run tests with -race
mise run build        # build rekal binary
```

CI runs `mise run test:ci` on push/PR; lint and license-check run as separate workflows.

## Release (GoReleaser + Homebrew)

On tag push `v*`, the release workflow:

1. Runs GoReleaser: builds binaries, creates GitHub Release, uploads artifacts and checksums.
2. Updates the Homebrew tap: pushes the cask to **rekal-dev/homebrew-tap** (so `brew install rekal-dev/tap/rekal` works).

**Setup:** Create the repo **rekal-dev/homebrew-tap** and either:

- **Option A (GitHub App):** See [docs/HOMEBREW_TAP_SETUP.md](docs/HOMEBREW_TAP_SETUP.md) for where to get `HOMEBREW_TAP_APP_ID` and `HOMEBREW_TAP_APP_PRIVATE_KEY`.
- **Option B (PAT):** Create a Personal Access Token with `contents: write` on `homebrew-tap`, set secret `TAP_GITHUB_TOKEN`, and in the release workflow use that instead of the app-token step.

Until the tap exists, you can remove or comment out the "Generate Homebrew Tap token" step and the `homebrew_casks` block in `.goreleaser.yaml` to still get GitHub Releases and the curl-install script.

## License

Apache-2.0 — see [LICENSE](LICENSE).
