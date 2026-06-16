# gitutil

The single git-execution package for the monorepo. All git operations shell out to the `git` CLI through `Run`; the go-git library must not be used here (it has repeatedly caused issues: it wipes git-ignored files on checkout, cannot fetch or push private repos without explicit auth, and does not resolve linked-worktree configs).

## Invariants

- All git commands go through `Run`, which forces `LC_ALL=C` (so stderr substring checks are stable), sets `GIT_TERMINAL_PROMPT=0` (no interactive credential prompts), and redacts URL-embedded credentials in error messages.
- **Credentials must never be persisted to `.git/config`.** `SetRemote` and `CloneWithConfig` store clean URLs only; credential-embedded URLs from `Config.FullyQualifiedRemote()` are passed as command-line arguments per invocation.
- Minimum supported git version is 2.11 (required by `status --porcelain=v2`). Do not introduce flags that raise the floor (e.g. `git init -b`, which requires 2.28; use `git symbolic-ref HEAD` instead, see `EnsureInit`).

## API guidance

Common operations get named functions (`Clone`, `Fetch`, `Status`, `Pull`, `Push`, `CommitAll`, `SetRemote`, ...). One-off commands should call `Run` directly at the call site instead of adding single-use helpers.

## File map

- `run.go` — the `Run` exec primitive and credential redaction
- `config.go` — `Config` (remote + ephemeral credentials) and `Signature` (commit authorship)
- `repo.go` — repo detection, init, clone, repo-root/subpath inference
- `commit.go` — branch and identity lookup, checkout, commit, commit-and-push
- `remote.go` — remote listing and management
- `sync.go` — status, fetch, pull, push, merge
- `github.go` — GitHub remote URL parsing and normalization
- `gitignore.go` — `.gitignore` management via a `drivers.RepoStore`
