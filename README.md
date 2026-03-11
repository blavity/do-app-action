# do-app-action

GitHub Actions for [DigitalOcean App Platform](https://www.digitalocean.com/products/app-platform/).

Provides four composable operations:

| Action | Description |
|--------|-------------|
| [`deploy`](#deploy) | Deploy an app from a spec or existing app name |
| [`delete`](#delete) | Delete an app by ID, name, or PR preview |
| [`archive`](#archive) | Archive (maintenance mode) an app |
| [`unarchive`](#unarchive) | Restore an archived app and wait until live |

> **Attribution:** `deploy` and `delete` are forked from [digitalocean/app_action](https://github.com/digitalocean/app_action) (MIT). `archive` and `unarchive` are original additions using the same `godo` client.

## Prerequisites

Store your DigitalOcean Personal Access Token as a GitHub Actions secret (e.g. `DO_ACCESS_TOKEN`). See [creating a personal access token](https://docs.digitalocean.com/reference/api/create-personal-access-token/).

---

## deploy

Deploy an app to DigitalOcean App Platform.

```yaml
- uses: blavity/do-app-action/deploy@v1
  with:
    token: ${{ secrets.DO_ACCESS_TOKEN }}
    app_spec_location: .do/app.yaml   # default
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `token` | yes | | DigitalOcean Personal Access Token |
| `app_spec_location` | no | `.do/app.yaml` | Path to app spec file. Mutually exclusive with `app_name`. |
| `project_id` | no | | Project to deploy into. Defaults to the account default project. |
| `app_name` | no | | Existing app name to pull spec from. Mutually exclusive with `app_spec_location`. |
| `print_build_logs` | no | `false` | Stream build logs to the Actions log. |
| `print_deploy_logs` | no | `false` | Stream deploy logs to the Actions log. |
| `deploy_pr_preview` | no | `false` | Deploy as a per-PR preview app (derives name from PR, strips domains/alerts, pins branch). |

### Outputs

| Output | Description |
|--------|-------------|
| `app` | Full JSON representation of the deployed app. |
| `build_logs` | Build log output. |
| `deploy_logs` | Deploy log output. |

---

## delete

Delete an app from DigitalOcean App Platform.

```yaml
- uses: blavity/do-app-action/delete@v1
  with:
    token: ${{ secrets.DO_ACCESS_TOKEN }}
    from_pr_preview: 'true'
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `token` | yes | | DigitalOcean Personal Access Token |
| `app_id` | no | | App ID to delete. |
| `app_name` | no | | App name to delete. |
| `from_pr_preview` | no | `false` | Derive app name from the current PR number. |
| `ignore_not_found` | no | `false` | Exit successfully if the app does not exist. |

---

## archive

Archive (put into maintenance mode) an app on DigitalOcean App Platform. The app stops serving traffic and billing drops to storage-only rates.

```yaml
- uses: blavity/do-app-action/archive@v1
  with:
    token: ${{ secrets.DO_ACCESS_TOKEN }}
    app_name: my-pr-preview-app
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `token` | yes | | DigitalOcean Personal Access Token |
| `app_id` | no | | App ID to archive. |
| `app_name` | no | | App name to archive. |
| `from_pr_preview` | no | `false` | Derive app name from the current PR number. |
| `ignore_not_found` | no | `false` | Exit successfully if the app does not exist. |

### Outputs

| Output | Description |
|--------|-------------|
| `app` | Full JSON representation of the app after archiving. |

---

## unarchive

Restore an archived app on DigitalOcean App Platform and wait until it is live.

```yaml
- uses: blavity/do-app-action/unarchive@v1
  with:
    token: ${{ secrets.DO_ACCESS_TOKEN }}
    app_name: my-pr-preview-app
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `token` | yes | | DigitalOcean Personal Access Token |
| `app_id` | no | | App ID to unarchive. |
| `app_name` | no | | App name to unarchive. |
| `from_pr_preview` | no | `false` | Derive app name from the current PR number. |
| `ignore_not_found` | no | `false` | Exit successfully if the app does not exist. |
| `wait_for_live` | no | `true` | Poll until `live_url` is populated after unarchiving. |
| `wait_timeout` | no | `300` | Seconds to wait for the app to become live (0 = no timeout). |

### Outputs

| Output | Description |
|--------|-------------|
| `app` | Full JSON representation of the app after unarchiving. |
| `live_url` | The app's live URL once it is available. |

---

## Versioning

This action uses [Semantic Versioning](https://semver.org/). Pin to a major tag (e.g. `@v1`) for stability. See [releases](https://github.com/blavity/do-app-action/releases) for the changelog.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Trademarks

DigitalOcean and App Platform are trademarks or registered trademarks of DigitalOcean, LLC. Blavity, Inc. is not affiliated with, endorsed by, or sponsored by DigitalOcean, LLC. All other trademarks are the property of their respective owners.

## License

MIT — see [LICENSE](LICENSE). Portions derived from [digitalocean/app_action](https://github.com/digitalocean/app_action), Copyright (c) 2024–2026 DigitalOcean, LLC.
