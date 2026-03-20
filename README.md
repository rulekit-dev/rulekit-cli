# rulekit-cli

`rulekit` is the command-line interface for the [RuleKit](https://github.com/rulekit) ecosystem. It pulls published rule bundles (`.zip`) from a [rulekit-registry](https://github.com/rulekit-dev/rulekit-registry) instance at build or deploy time, extracts them into a local `.rulekit/` directory, and maintains a `rulekit.lock` file that records the exact version and SHA-256 checksum of every pulled ruleset — ensuring reproducible, tamper-evident rule deployments that your application can evaluate with [rulekit-sdk](https://github.com/rulekit-dev/rulekit-sdk).

---

## Installation

**go install**
```sh
go install github.com/rulekit-dev/rulekit-cli@latest
```

**Homebrew**
```sh
brew install rulekit-dev/tap/rulekit
```

**Direct download**

Download the latest release binary from the [releases page](https://github.com/rulekit-dev/rulekit-cli/releases) and place it in your `PATH`.

---

## Quickstart

**1. Point the CLI at your registry**

```sh
export RULEKIT_REGISTRY_URL=https://registry.example.com
export RULEKIT_TOKEN=rk_abc123          # omit if registry has no auth
```

Or pass them inline as flags (the registry URL is saved to `rulekit.lock` after the first run):

```sh
rulekit add payout-routing --registry https://registry.example.com --token rk_abc123
```

**Token formats accepted:** a `rk_` prefixed API token or a JWT — both sent as `Authorization: Bearer <token>`. The token is never written to `rulekit.lock`.

**2. Add a ruleset and pull the latest bundle**

```sh
rulekit add payout-routing
# rulekit: pulling payout-routing@latest…
# rulekit: locked payout-routing v1 · sha256:a3f1c8…
```

**3. On subsequent runs (e.g., in CI), pull the locked versions**

```sh
rulekit pull
# rulekit: pulling payout-routing@1…
# rulekit: locked payout-routing v1 · sha256:a3f1c8…
```

**4. Verify checksums haven't been tampered with**

```sh
rulekit verify
# rulekit: ✓ payout-routing: ok
# rulekit: all checksums verified (1 rulesets)
```

---

## Stack management

### Setup wizard

The first time you run `rulekit up`, a setup wizard guides you through configuring the stack:

```sh
rulekit up
```

To reconfigure an existing setup:

```sh
rulekit up --reconfigure
```

For CI or scripted environments (accept all defaults):

```sh
rulekit up --yes
```

The wizard configures:
- Database (SQLite or Postgres)
- Blob storage (filesystem or S3 / R2 / MinIO)
- Authentication (none, API key, or JWT)
- Ports

Config is saved to `~/.rulekit/compose/.env`. Secrets are stored with `0600` permissions (owner-only).

---

Start the full RuleKit stack (registry + dashboard) with one command:

```sh
rulekit up
```

With Postgres:

```sh
rulekit up --postgres
```

Check health:

```sh
rulekit status
```

Open the visual editor:

```sh
rulekit dashboard
```

Stop:

```sh
rulekit down
```

Upgrade to latest:

```sh
rulekit upgrade
```

| Command | Flags | Description |
|---------|-------|-------------|
| `rulekit up` | `--postgres`, `--port`, `--dashboard-port`, `--registry-image`, `--dashboard-image` | Start registry + dashboard via Docker |
| `rulekit down` | — | Stop all containers |
| `rulekit restart` | — | Stop + start, preserving existing config |
| `rulekit status` | — | Infra health check (registry, dashboard, db) + ruleset update status |
| `rulekit dashboard` | — | Open dashboard in default browser |
| `rulekit logs` | `--service`, `--follow` | Tail container logs |
| `rulekit upgrade` | — | Pull latest images + rolling restart |
| `rulekit uninstall` | — | Stop containers + remove `~/.rulekit/compose/` |

---

## Ruleset management

### `rulekit pull`


Pull rule bundles from the registry.

```
rulekit pull [--key <key>] [--version <n|latest>] [--namespace <ns>]
```

| Flag | Description |
|------|-------------|
| `--key` | Pull a specific ruleset key. If omitted, pulls all rulesets in the lockfile. |
| `--version` | Version to pull (`latest` or an integer). Defaults to the locked version, or `latest` if not yet locked. |
| `--namespace` | Namespace override. |

### `rulekit add`

Add a new ruleset to the lockfile and pull it immediately.

```
rulekit add <key> [--version <n|latest>] [--namespace <ns>]
```

| Flag | Description |
|------|-------------|
| `--version` | Version to pull (default: `latest`). |
| `--namespace` | Namespace override. |

### `rulekit remove`

Remove a ruleset from the lockfile and delete its local files.

```
rulekit remove <key>
```

### `rulekit list`

List all locked rulesets with their version, checksum, and pull timestamp.

```
rulekit list
```

Output:
```
key              version  checksum                   pulled_at
payout-routing   v4       sha256:a3f1c8…             2025-01-01
fraud-scoring    v2       sha256:b7e2d1…             2025-01-01
```

### `rulekit verify`

Recompute SHA-256 checksums of all local `dsl.json` files and compare them against the lockfile. Exits with code `2` if any mismatch is detected.

```
rulekit verify
```

### `rulekit status`

Check each locked ruleset against the registry to see if updates are available.

```
rulekit status
```

Output:
```
rulekit: ✓ payout-routing: up to date (v4)
rulekit: → fraud-scoring: update available (local v2, latest v5)
```

### Global flags

| Flag | Description |
|------|-------------|
| `--registry` | Registry base URL. |
| `--namespace` | Namespace (default: `default`). |
| `--dir` | Local output directory (default: `.rulekit`). |
| `--token` | Bearer token for authenticated registries. |
| `--verbose` | Enable structured `slog` output to stderr. |

---

## rulekit.lock

`rulekit.lock` is a JSON file written to your project root. It records the registry URL, namespace, and for each ruleset: the locked version number, SHA-256 checksum, and pull timestamp.

```json
{
  "registry": "http://localhost:8080",
  "namespace": "default",
  "rulesets": {
    "payout-routing": {
      "version": 4,
      "checksum": "sha256:a3f1c8...",
      "pulled_at": "2025-01-01T00:00:00Z"
    }
  }
}
```

**What to commit:** `rulekit.lock` — commit it to version control so all team members and CI pipelines pull identical rule versions.

**What to gitignore:** `.rulekit/` — the extracted bundle contents are ephemeral build artifacts. Add `.rulekit/` to your `.gitignore`.

---

## Testing Locally

**1. Build the binary**
```sh
make build
# binary written to bin/rulekit
```

**2. Run unit tests**
```sh
go test ./...
```

**3. Install globally**
```sh
make install
```

This runs `go install ./` and places the binary in `$(go env GOPATH)/bin`. If `rulekit` is not found after install, add that directory to your PATH:

```sh
# Add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload
source ~/.zshrc
```

**4. Manual smoke test**

```sh
mkdir /tmp/rulekit-test && cd /tmp/rulekit-test

# Add a ruleset (pulls immediately)
rulekit add payout-routing --registry http://localhost:8080

# List locked rulesets
rulekit list

# Verify checksums
rulekit verify

# Check for updates
rulekit status

# Pull a specific version
rulekit pull --key payout-routing --version 3

# Remove a ruleset
rulekit remove payout-routing
```

**5. Test with verbose logging**
```sh
rulekit pull --verbose
```

---

## CI Usage

```yaml
# .github/workflows/deploy.yml
- name: Pull rule bundles
  env:
    RULEKIT_REGISTRY_URL: ${{ secrets.RULEKIT_REGISTRY_URL }}
    RULEKIT_TOKEN: ${{ secrets.RULEKIT_TOKEN }}
  run: |
    go install github.com/rulekit-dev/rulekit-cli@latest
    rulekit pull
    rulekit verify
```

---

## Configuration Reference

Configuration is resolved in priority order: CLI flags > environment variables > `rulekit.lock` values > defaults.

| Environment Variable | CLI Flag | Default | Description |
|---|---|---|---|
| `RULEKIT_REGISTRY_URL` | `--registry` | `http://localhost:8080` | Registry base URL |
| `RULEKIT_NAMESPACE` | `--namespace` | `default` | Ruleset namespace |
| `RULEKIT_DIR` | `--dir` | `.rulekit` | Local output directory |
| `RULEKIT_TOKEN` | `--token` | _(empty)_ | Bearer token (JWT or `rk_` API key) |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error (registry unreachable, missing lockfile, etc.) |
| `2` | Checksum mismatch — use in CI to detect bundle tampering |

---

## Related Projects

- [rulekit-registry](https://github.com/rulekit-dev/rulekit-registry) — Control plane: stores, versions, and publishes rulesets
- [rulekit-sdk](https://github.com/rulekit-dev/rulekit-sdk) — Evaluates rules locally at runtime
- [rulekit-dashboard](https://github.com/rulekit-dev/rulekit-dashboard) — Visual editor for rules

---

## License

MIT — see [LICENSE](./LICENSE).
