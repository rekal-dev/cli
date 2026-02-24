# Homebrew tap: where to get the secrets

The release workflow uses **actions/create-github-app-token** to get a token that can push to **rekal-dev/homebrew-tap**. You need a **GitHub App** and two repository secrets.

---

## 1. Create the tap repo

- Create a **new repository** under the **rekal-dev** org (or your org): **homebrew-tap**.
- It can be empty (no need to clone or add files; GoReleaser will push the cask).

---

## 2. Create a GitHub App

1. Go to **GitHub** → **Settings** (your user or org) → **Developer settings** → **GitHub Apps** → **New GitHub App**.
2. Fill in:
   - **Name:** e.g. `rekal-homebrew-tap`
   - **Homepage URL:** e.g. `https://github.com/rekal-dev/cli`
   - **Webhook:** Uncheck "Active" (not needed for release).
   - **Repository permissions:** **Contents** = Read and write.
   - **Where can this GitHub App be installed?** → **Only on this account** (or your org).
3. Click **Create GitHub App**.

---

## 3. Get App ID and private key

- **App ID:** On the app’s page, under "About", you’ll see **App ID**. Copy it (numeric, e.g. `123456`).
- **Private key:** In the same page, go to **Private keys** → **Generate a private key**. A `.pem` file is downloaded. Open it and copy the **entire** contents (including `-----BEGIN RSA PRIVATE KEY-----` and `-----END RSA PRIVATE KEY-----`).

---

## 4. Install the App on the tap repo

1. In the app’s page, click **Install App**.
2. Choose the account/org (**rekal-dev**).
3. Select **Only select repositories** and pick **homebrew-tap**.
4. Click **Install**.

---

## 5. Add secrets to the CLI repo

Use **repository** secrets, not an Environment.

1. Open **rekal-dev/cli** on GitHub → **Settings** → **Secrets and variables** → **Actions**.
2. Under **Repository secrets** (not "Environments"), click **New repository secret**.
3. Add these two secrets:

| Secret name | Value |
|-------------|--------|
| `HOMEBREW_TAP_APP_ID` | The App ID (e.g. `123456`). |
| `HOMEBREW_TAP_APP_PRIVATE_KEY` | The full contents of the `.pem` file (the whole key, including BEGIN/END lines). |

**If GitHub asks you to "create an environment":** you don’t need one for this. Stay on **Repository secrets** and add the two secrets there. The release workflow does not use an Actions environment.

The release workflow passes these into **actions/create-github-app-token**, which uses `repositories: homebrew-tap` so the generated token can push to **rekal-dev/homebrew-tap**.

---

## Alternative: Personal Access Token (PAT)

If you prefer not to use a GitHub App:

1. Create a **Fine-grained** or **Classic** PAT with **Contents: Read and write** on **rekal-dev/homebrew-tap**.
2. In **rekal-dev/cli**, add a secret **`TAP_GITHUB_TOKEN`** with that token.
3. In **.github/workflows/release.yml**, **remove** the step "Generate Homebrew Tap token" and pass the secret into GoReleaser:

   ```yaml
   env:
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
     TAP_GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}
   ```

The **.goreleaser.yaml** already uses `TAP_GITHUB_TOKEN` for the tap; the workflow only needs to provide it.
