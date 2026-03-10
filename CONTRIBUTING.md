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

1. Merge to `main`.
2. Create a GitHub release with a semver tag (`v1.2.3`).
3. Move the floating major tag (`v1`) to the new commit:
   ```bash
   git tag -fa v1 -m "Update v1 to v1.2.3"
   git push origin v1 --force
   ```

## Attribution

The `deploy` and `delete` operations are forked from [digitalocean/app_action](https://github.com/digitalocean/app_action) (MIT, Copyright 2024 DigitalOcean, LLC). The `archive` and `unarchive` operations are original additions.
