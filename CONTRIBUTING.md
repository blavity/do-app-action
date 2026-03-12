# Contributing

## Development

### Prerequisites

- [mise](https://mise.jdx.dev/) — manages Go, golangci-lint, and task versions

### Setup

```bash
git clone https://github.com/blavity/do-app-action
cd do-app-action
mise install        # installs Go, golangci-lint, task (pinned in mise.toml)
go mod download
```

Optional: install pre-commit hooks to catch issues before pushing:

```bash
pip install pre-commit
pre-commit install
```

### Common tasks

| Command      | What it does                           |
| ------------ | -------------------------------------- |
| `task check` | fmt + lint + test — run before pushing |
| `task fmt`   | Format code (golangci-lint fmt)        |
| `task lint`  | Run golangci-lint                      |
| `task test`  | Run tests with race detector           |
| `task build` | Build all action binaries              |

### Building binaries locally

```bash
task build
# or directly:
go build -o /tmp/deploy ./deploy
go build -o /tmp/delete ./delete
go build -o /tmp/archive ./archive
go build -o /tmp/unarchive ./unarchive
```

### Building the Docker image

```bash
docker build -t do-app-action:local .
```

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): description`.

Scopes: `deploy`, `delete`, `archive`, `unarchive`, `utils`, `ci`, `docs`.

## Releasing

Releases are fully automated via [release-please](https://github.com/googleapis/release-please):

1. Merge commits to `main` using [Conventional Commits](https://www.conventionalcommits.org/) (e.g. `fix(delete): ...`, `feat(deploy): ...`).
2. release-please will open or update a release PR aggregating the changes.
3. Merge the release PR — release-please creates the semver tag (`v1.x.x`) and GitHub Release automatically.
4. The `release.yml` workflow fires and moves the floating `v1` tag to the new release commit.

No manual tagging or `git push --force` is required.

## Attribution

The `deploy` and `delete` operations are forked from [digitalocean/app_action](https://github.com/digitalocean/app_action) (MIT, Copyright 2024 DigitalOcean, LLC). The `archive` and `unarchive` operations are original additions.
