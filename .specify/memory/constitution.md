<!--
SYNC IMPACT REPORT
==================
Version change: 1.3.0 → 1.3.1
Modified principles: none
Added sections:
  - Responsible Agentic Use & Pull Request Policy (new standalone section,
    covering scope of autonomy, authorship transparency, PR review policy,
    reversibility preference, and escalation over assumption)
Removed sections: none
Wording changes (1.3.0 → 1.3.1):
  - Principle VII: replaced "Renovate manages automated dependency updates"
    with "Dependabot manages automated dependency updates"
  - Security & Supply Chain: replaced "Dependabot/Renovate MUST remain enabled"
    with "Dependabot MUST remain enabled"
    Rationale: repo migrated from Renovate to Dependabot; Renovate references
    were also a Principle IX violation (org-specific shared preset exposed).
Templates requiring updates:
  - .specify/templates/plan-template.md  ✅ no structural changes required
  - .specify/templates/spec-template.md  ✅ no structural changes required
  - .specify/templates/tasks-template.md ✅ no structural changes required
Follow-up TODOs:
  - TODO(COVERAGE_CI_GATE): wire `go tool cover -func coverage.out` threshold
    check into ci.yml. The 80% target is now constitutionally mandated but not
    yet enforced by CI automation. Current baseline: 18.5% total (main.go
    business logic is 0% — refactoring needed before gate can be activated).
    Tracked as a prerequisite before the CI gate is enabled.
-->

# do-app-action Constitution

## Core Principles

### I. Composable, Single-Responsibility Actions

Each action (`deploy`, `delete`, `archive`, `unarchive`) MUST do exactly one
thing and do it well. Cross-action orchestration is the caller's responsibility
via GitHub Actions workflow composition. No action binary MAY import logic from
another action's package; shared code MUST live in `utils/`.

**Rationale**: Composability is the primary user value. Tight coupling between
actions increases blast radius for bugs and makes independent versioning and
testing impossible.

### II. Public API Stability

Inputs and outputs declared in each `action.yml` are a public contract.
Removing or renaming an existing input or output is a BREAKING CHANGE and MUST
trigger a semver MAJOR bump. Adding optional inputs with defaults is
backward-compatible (MINOR). Callers pin to `@v1`; breaking that tag silently
is unacceptable.

**Rationale**: This action is consumed by external CI pipelines. Unexpected
breakage in a dependency they don't control erodes trust and causes production
incidents.

### III. Idempotency & Safe Defaults

Every action MUST be safe to re-run on the same target without corrupting state.
Flags like `ignore_not_found` MUST default to the least-destructive option
(`false`). Operations that wait for remote state (e.g., `unarchive`) MUST
surface timeout controls with sane defaults rather than blocking indefinitely.

**Rationale**: GitHub Actions retries and human re-runs are common. An action
that fails differently on second invocation is a reliability bug.

### IV. Explicit Credential Handling — No Ambient Auth

The DigitalOcean token MUST be passed explicitly via the `token` input on every
action invocation. No action MAY read credentials from environment variables,
files, or any source other than the declared `token` input. Tokens MUST NOT be
logged, echoed, or included in outputs.

**Rationale**: Explicit credential flow makes secret exposure auditable and
prevents accidental token leakage in logs — critical for a public, open-source
action running in third-party CI environments.

### V. Test Every Behavior Unit

All non-trivial input parsing, validation, and business logic MUST have unit
tests using `github.com/stretchr/testify`. Tests MUST be run with the race
detector (`go test -race`). Integration tests against the live DO API are out
of scope for CI but MAY be run manually.

**Coverage threshold**: The minimum acceptable statement coverage for the
repository is **80%**, measured by `go test -race -coverprofile=coverage.out
./...` followed by `go tool cover -func coverage.out`. CI MUST fail if the
total statement coverage falls below this threshold once the gate is activated.

**Exclusions**: Only bare `main()` entry-point functions (those whose sole
purpose is wiring `os.Exit` and delegating to a `run()` or equivalent) MAY be
excluded from coverage counting. Business logic that currently lives in
`main.go` files (`createSpec`, `deploy`, `waitForDeployment`, `printLogs`,
`waitForLiveURL`, etc.) is NOT excluded — it MUST be refactored into testable
packages and covered before the CI gate is enabled.

**Ratchet rule**: Statement coverage MUST NOT decrease between merges to
`main`. PRs that reduce coverage MUST include a justification in the PR
description; a net reduction without justification is a CI failure once the
gate is active.

**New code**: Code added in a PR that has no test coverage MUST be justified in
the PR description. "I'll add tests later" is not a justification.

**Rationale**: The action binaries are distributed as Docker images with no
runtime introspection. Unit tests are the primary quality gate. The race
detector is mandatory because actions may be called concurrently in matrix
workflows. 80% is the widely recognized industry floor for non-trivial Go
libraries; it balances meaningful coverage signal against diminishing returns
on entry-point and integration wiring code.

### VI. Conventional Commits Drive Automated Releases

All commits to `main` MUST follow the Conventional Commits specification with
scopes matching the changed component (`deploy`, `delete`, `archive`,
`unarchive`, `utils`, `ci`, `docs`). `feat:` bumps MINOR, `fix:` bumps PATCH,
`feat!:` / `BREAKING CHANGE:` bumps MAJOR. Release-please automation MUST NOT
be bypassed. Manual tagging or force-pushing version tags is prohibited.

**Rationale**: Automated changelogs and semver accuracy depend entirely on
commit discipline. This is a public repo — consumers read the changelog.

### VII. Minimal, Pinned Dependencies

Direct dependencies MUST be limited to what is strictly necessary:
`godo` (DO API client), `go-githubactions` (Actions SDK), `testify` (testing),
`sigs.k8s.io/yaml` (spec parsing). Adding a new direct dependency requires
explicit justification in the PR. All GitHub Actions workflow steps MUST be
pinned to a full commit SHA alongside the version tag comment (e.g.,
`uses: actions/checkout@<sha> # v4`). Dependabot manages automated dependency
updates.

**Rationale**: A smaller dependency surface reduces supply-chain attack area,
speeds Docker builds, and keeps the binary portable. SHA pinning in workflows
prevents tag-hijacking attacks in a public repo.

### VIII. GitHub Token Sufficiency & Secrets Independence

All workflows and scripts in this repository MUST operate using only the
built-in `GITHUB_TOKEN` (granted automatically by GitHub Actions) and the
caller-supplied DigitalOcean `token` input governed by Principle IV. No
workflow, script, or action MAY require an organization-level or repository
secret beyond those two. The `GITHUB_TOKEN` MUST be scoped to the minimum
permissions required for the job (e.g., `contents: read`, `packages: write`)
and those scopes MUST be declared explicitly in the workflow `permissions` block.

Any caller of this action — including forks and external organizations — MUST
be able to run all workflows successfully with only a DigitalOcean API token
and the default `GITHUB_TOKEN`. No reliance on secrets that only the
originating organization possesses is permitted.

**Rationale**: This is a public, forkable action. Requiring org-specific secrets
creates an invisible dependency that silently breaks the action for any adopter
outside the originating organization. Least-privilege scoping minimizes the
blast radius of a compromised token.

### IX. Organizational Information Confidentiality

No artifact committed to this repository — including but not limited to source
code, documentation, scripts, agent instructions, specs, plans, and this
constitution — MAY reference, describe, infer, or otherwise expose:

- Internal processes, workflows, or tooling of any host organization.
- Organizational policies, standards, or internal URLs that are not publicly
  documented.
- Trade secrets, proprietary methods, or competitive information belonging to
  any organization using this action.
- Personally identifiable information of employees, contractors, or teams beyond
  what is already public (e.g., a public GitHub username on a commit).

Guidance, examples, and documentation MUST be written generically so that any
organization adopting this action can use it without needing context about the
originating organization.

**Rationale**: This repository is public and forkable. Documentation or tooling
that embeds org-specific context leaks internal information and reduces the
action's utility to the broader community. Generic, portable artifacts protect
both the originating organization and future adopters.

### X. Own the Codebase

Pre-existing issues encountered during work MUST be surfaced and tracked, never
silently ignored or dismissed as "not my change."

- When a CI run, workflow output, or code review reveals a broken state or
  pre-existing bug, contributors and agents MUST clearly distinguish what is
  caused by the current change versus what is pre-existing.
- Pre-existing issues MUST NOT be silently worked around, dismissed as
  out-of-scope, or annotated with "nothing to do with our changes" without a
  tracking artifact.
- Trivial fixes (missing `strings.TrimSpace` calls, stale comments, golangci-lint
  findings that can be corrected in one line) MUST be fixed in the current PR when
  the fix is low-risk and does not expand the PR's scope unreasonably.
- Non-trivial pre-existing issues MUST have a GitHub issue created with context,
  root cause, and suggested remediation. Deferred without an issue is the same
  as forgotten.
- CI findings from linting, `go vet`, or supply-chain scans that do not block
  the current merge MUST still be tracked via a GitHub issue. Silently passing
  CI by ignoring pre-existing findings is prohibited.

**Rationale**: A public action that accumulates silent technical debt becomes
unreliable for all consumers. Every contributor — human or agent — is a
co-owner responsible for the overall health of the codebase, not just their
diff.

### XI. No Stranded Work

Completed work MUST reach its destination. Code that is committed locally but
never pushed, pushed but never PR'd, or PR'd but never merged is stranded work
and MUST be treated as a defect.

- Every commit that lands on a local `main` (or default branch) MUST be pushed
  to the remote within the same work session. If pushing is blocked, a branch
  and PR MUST be created instead.
- After every merge, rebase, or cherry-pick, `git status` MUST be run to confirm
  the working tree is clean and no commits are unpushed.
- Release-please version bump PRs MUST be merged promptly after their triggering
  feature lands. A release commit that sits unmerged defeats automated versioning;
  downstream consumers cannot pin to a version that does not exist on the remote.
- CI/CD pipelines MUST NOT be considered complete until the deployed artifact
  (GitHub Release + floating major tag) reflects the intended changes. Merging a
  PR is not shipping; verifying the release tag was created and the floating `v1`
  tag was updated is.
- Agents MUST run `git status` after every commit and resolve any unpushed state
  before ending the session.

**Rationale**: In an automated release pipeline, stranded commits delay every
consumer pinned to the floating major tag. The cost of a missed push is
invisible until a downstream pipeline silently runs old code.

### XII. Documentation as Artifact

User-facing and operational changes MUST include corresponding documentation
updates committed in the same PR.

- New action inputs or outputs MUST be documented in the relevant `action.yml`
  description fields and in the `README.md` input/output tables.
- Breaking changes MUST include a migration note in the PR description and, where
  appropriate, in `README.md`. The automated changelog (release-please) is
  supplementary, not a substitute for in-repo guidance.
- Changes to developer workflow (new `task` targets, changed prerequisites,
  updated toolchain versions) MUST be reflected in `CONTRIBUTING.md`.
- Documentation additions MUST be accurate and self-contained; they MUST NOT
  reference internal systems, internal runbooks, or org-specific tooling
  (see Principle IX).

**Rationale**: `README.md` and `CONTRIBUTING.md` are the sole documentation
surface for this repository. When they drift from the implementation, external
adopters and contributors operate on false information — a silent reliability
failure with no error message.

## Security & Supply Chain

Actions in this repository run with access to caller secrets (DO tokens). The
following controls are NON-NEGOTIABLE:

- All GitHub Actions steps MUST use SHA-pinned action references.
- CI workflow permissions MUST follow least-privilege (`contents: read` unless
  a broader scope is explicitly justified in a comment).
- No secrets or token values MAY appear in Go `log`, `fmt.Print*`, or Actions
  `core.Info`/`core.Debug` output.
- Dependabot MUST remain enabled for Go modules and GitHub Actions.
- The Docker base image MUST be SHA-pinned (e.g.,
  `golang:1.26-alpine@sha256:...`). Floating tags are prohibited in
  `Dockerfile`.
- PRs from external contributors MUST pass CI before any maintainer reviews
  secrets-adjacent code paths.
- Action binaries MUST surface failure reasons through GitHub Actions error
  annotations (`core.SetFailed`) or structured stderr output. Silent failures
  that produce a zero exit code are prohibited.

## Agentic Development Standards

This repository is designed for heavy agentic (AI agent) use. The following
rules govern agent-driven contributions:

- **Scope discipline**: Agents MUST confine changes to the action(s) named in
  the task. Do not modify sibling actions unless explicitly instructed.
- **No speculative dependencies**: Agents MUST NOT add Go module dependencies
  without explicit user approval. Run `go mod tidy` after any dependency
  changes; commit `go.sum` atomically with `go.mod`.
- **Conventional commits are mandatory**: Every agent commit MUST use a scoped
  conventional commit message. Unscoped commits break release-please automation.
- **Always run `task check`** before proposing a PR: `task fmt && task lint &&
  task test`. Never propose a PR that fails CI.
- **No token logging**: Agents MUST audit any code they write for accidental
  credential exposure before committing.
- **Lockfiles**: `go.sum` is the lockfile equivalent. It MUST be regenerated
  with `go mod tidy` after any `go.mod` changes and committed in the same PR.
- **Clean working tree rule**: After every commit, verify `git status` is clean.
  Orphaned files are never acceptable.
- **Public repo awareness**: This repo is public. Agents MUST NOT commit
  `.env` files, token values, internal hostnames, or any non-public information.
- **No org-specific context**: Agents MUST NOT embed references to the host
  organization's internal tooling, processes, team names, or policies in any
  committed artifact. All documentation and guidance MUST remain generic and
  portable (Principle IX).
- **Surface and track pre-existing issues**: When agents encounter CI failures,
  lint warnings, or bugs unrelated to the current task, they MUST create a
  GitHub issue before proceeding (Principle X). Silent workarounds are
  prohibited.
- **No stranded commits**: Agents MUST push all commits and verify the remote is
  up to date before ending a session (Principle XI).
- **Agent directories are gitignored**: `.claude/`, `.codex/`, `.opencode/`, and
  `.specify/**` (except `memory/constitution.md`) MUST remain in `.gitignore`.
  Agents MUST NOT commit their own working directories.

## Responsible Agentic Use & Pull Request Policy

AI agents are permitted contributors but operate under stricter constraints
than human contributors. The combination of speed, confidence, and context
gaps makes agentic work a higher-risk surface for this public, versioned repo.

**Scope of autonomy**:

- Agents MUST operate only within the scope explicitly defined by the current
  task. Refactoring, restructuring, or adding features beyond the stated task
  is prohibited without explicit user instruction.
- Agents MUST NOT open speculative PRs — PRs for work the user did not
  explicitly request in the current session.
- When requirements are ambiguous, agents MUST ask before implementing.
  Confident execution of an incorrect interpretation is a worse outcome than
  a clarifying question.

**Authorship transparency**:

- PRs authored by agents MUST include a statement in the PR description
  identifying the changes as agent-generated and summarizing the session scope
  (e.g., "Agent-generated: implements X as requested."). The statement MUST be
  human-readable and placed where a reviewer will see it before approving.
- Agent commits MUST NOT be attributed to a human identity. Co-authorship
  trailers (e.g., `Co-authored-by:`) are encouraged when the agent worked
  from a human-provided design or specification.

**Pull request policy**:

- Every agent-opened PR MUST receive at least one human maintainer approval
  before merge, without exception. Agents MUST NOT self-approve, use the
  GitHub API to dismiss reviews, or otherwise circumvent the review
  requirement.
- Agent PRs MUST include a Compliance Statement in the PR description listing
  each principle materially affected by the change and confirming compliance
  or documenting a justified exception. This is a merge requirement, not
  advisory.
- PRs that add, remove, or rename `action.yml` inputs or outputs MUST
  explicitly state the Principle II (Public API Stability) impact and the
  semver bump type in the PR description.
- PRs that introduce a new direct Go dependency MUST include a Principle VII
  justification. Dependency additions are never implicitly approved.
- Agent PRs MUST NOT be merged while CI is failing. A passing CI run is a
  necessary but not sufficient condition for merge.

**Reversibility preference**:

- Agents MUST prefer small, targeted changes over large-surface rewrites.
  When a task can be accomplished with a narrower diff, the narrower approach
  MUST be taken.
- Agents MUST NOT force-push to any branch. History rewriting by agents is
  prohibited.

**Rationale**: This is a public action pinned by external pipelines via a
floating major tag. An agent-generated change that merges without human review
can silently break every caller before anyone notices. The policies above keep
a human in the loop on every change that reaches `main` while still enabling
productive agentic development for the bulk of implementation work.

## Governance

This constitution supersedes all other development guidance. When a rule in
`CONTRIBUTING.md`, a PR comment, or an agent instruction conflicts with this
document, this document wins. To resolve genuine conflicts, amend this
constitution first.

**Amendment procedure**:
1. Open a PR with the proposed change to this file.
2. State the version bump type (MAJOR/MINOR/PATCH) and rationale in the PR
   description.
3. At least one maintainer MUST approve before merge.
4. `LAST_AMENDED_DATE` and `CONSTITUTION_VERSION` MUST be updated in the
   same commit as the content change.

**Versioning policy**: Semantic versioning as defined in Principle VI applies.
Principle removals or redefinitions = MAJOR. New principles or sections = MINOR.
Clarifications, wording, typos = PATCH.

**Compliance review**: Every PR description MUST include a brief statement
confirming compliance with affected principles or explicitly noting a justified
exception. Agents MUST include this in auto-generated PR bodies.

**Version**: 1.3.1 | **Ratified**: 2026-03-10 | **Last Amended**: 2026-03-13
