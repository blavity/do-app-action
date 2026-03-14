# Agent Guide — do-app-action

## Constitution check is mandatory

**Before making any change**, read the full constitution:

> `.specify/memory/constitution.md`

Every principle applies to every task. There are no exceptions scoped to
"small" or "quick" changes. If you have not read the constitution in this
session, read it now before proceeding.

---

## Repo at a glance

Four independent Go binaries, each in its own package, each with its own
`action.yml`:

```
deploy/      archive/      delete/      unarchive/
  main.go      main.go       main.go      main.go
  inputs.go    inputs.go     inputs.go    inputs.go
  action.yml   action.yml    action.yml   action.yml

utils/       — shared helpers (inputs, env, apps, preview)
```

Rules:
- Packages MUST NOT import each other. Shared code goes in `utils/`.
- `action.yml` inputs/outputs are a public API — removals and renames are
  BREAKING CHANGES (Principle II).
- Actions are deployed as Docker images built from the root `Dockerfile`.

## Toolchain

Managed by `mise`. Run `mise install` once after cloning.

| Command | What it does |
|---|---|
| `task check` | fmt + lint + test — run before every commit |
| `task fmt` | golangci-lint fmt |
| `task lint` | golangci-lint run |
| `task test` | go test -race -coverprofile=coverage.out ./... |
| `task build` | build all four binaries |

**Always run `task check` before proposing a PR.** Never open a PR that
fails CI.

## Commit format

```
type(scope): description
```

Valid scopes: `deploy`, `delete`, `archive`, `unarchive`, `utils`, `ci`,
`docs`. Commits without a valid scope break release-please automation.

## Before opening a PR

1. `task check` passes cleanly.
2. Every changed principle is addressed in the PR description (Compliance
   Statement — required, not advisory).
3. If you changed `action.yml` inputs or outputs, state the Principle II
   impact and semver bump type.
4. If you added a Go dependency, include a Principle VII justification.
5. The PR description states that changes are agent-generated.
6. You have pushed the branch and confirmed `git status` is clean.

A human maintainer must approve before merge. You may not self-approve or
dismiss reviews.

## Hard stops — ask before proceeding

Stop and ask the user if you encounter any of the following:

- A task that would change `action.yml` inputs or outputs
- A task that would add a new direct Go dependency
- Ambiguous requirements where multiple interpretations are plausible
- A pre-existing bug or CI failure that is non-trivial to fix
- Anything that would expand scope beyond what was explicitly requested

## What to do with pre-existing issues

- **Trivial** (one-line lint fix, stale comment): fix in the current PR.
- **Non-trivial**: open a GitHub issue with context and root cause, then
  continue. Do not silently work around or ignore.

## What this repo does not have

No MkDocs site, no integration test suite, no observability pipeline, no
org-level secrets. If a task assumes otherwise, stop and verify before
implementing.
