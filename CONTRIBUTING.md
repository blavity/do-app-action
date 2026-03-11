# Contributing

## Development

### Prerequisites

- Go 1.23+
- Docker (for local Dockerfile builds)

### Setup

```bash
git clone https://github.com/blavity/do-app-action
cd do-app-action
go mod download
```

### Running tests

```bash
go test ./...
```

### Building binaries locally

```bash
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
