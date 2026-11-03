# Verified: cli.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **14** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| cli-02 | MEDIUM | P1 | PKCE state map has concurrent read/write race | `cli/pkg/local/app.go:59`; writes/reads/deletes remain unlocked at `cli/pkg/local/server.go:825,845,852` | Protect store and load-delete with a mutex or `sync.Map`. |
| cli-03 | MEDIUM | P1 | Shared Helper clients are mutated concurrently | `cli/pkg/cmdutil/helper.go:170-277`; admin/activity caches and config fields are mutated without a common lock | Lock cache/config state or initialize immutable clients before serving concurrent RPCs. |
| cli-04 | MEDIUM | P2 | Local auth/logout accepts open redirects | `cli/pkg/local/server.go:815-820,879,897-901`; caller-provided redirect is persisted/forwarded without same-origin validation | Accept only local relative paths or validate against the known local origin. |
| cli-05 | MEDIUM | P1 | Credentials are written world-readable | `cli/pkg/dotrill/dotrill.go:113,362`; files use `0644`, directory uses `os.ModePerm` | Use `0600` for credentials and `0700` for the config directory; repair existing modes on write. |
| cli-06 | MEDIUM | P2 | Project listing swallows non-inference errors | `cli/pkg/local/server.go:727-745`; non-`ErrInferProjectFailed` errors fall through to an empty successful response | Return the original error from the error branch. |
| cli-08 | LOW | P3 | PKCE callback validates `code` twice | `cli/pkg/local/server.go:835-845`; second guard still tests `code` instead of `state` | Test `state == ""` and return the intended missing-state error. |
| cli-09 | LOW | P3 | Device-auth interface is dead and mismatched | `cli/pkg/deviceauth/authenticator.go:29-32,80+`; interface signatures do not match the concrete implementation and have no consumers | Remove it or align signatures and add a compile-time assertion. |
| cli-11 | LOW | P3 | PKCE verifier generation has modulo bias | `cli/pkg/pkce/authenticator.go:150-159`; bytes are reduced modulo a 66-character alphabet | Generate RFC 7636 verifier bytes with raw URL-safe base64 or rejection sampling. |
| cli-14 | LOW | P2 | Connector error silently drops app-start telemetry | `cli/pkg/local/app.go:536-556`; connector error branch returns the earlier nil `err` | Return `c.Err`, or explicitly document and instrument an intentional skip. |
| cli-12 | LOW | P2 | `.env` upload is invisible to users | `cli/cmd/project/deploy.go:309,437-451`; default is enabled and variable names are not shown; local UI flow has no opt-out | Show names and request confirmation in interactive flows; expose/advertise an opt-out everywhere. |
| cli-13 | LOW | P2 | Server polling can accept the wrong process | `cli/pkg/local/app.go:480-529`; a 250 ms sleep races cancellation against another server's healthy port | Surface the listener bind result directly and fail deterministically on an occupied port. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| cli-01 | FIXED — remove | `cli/pkg/local/server.go:454-457` now sends `.env` variables to `projResp.Project.OrgName` / `.Name`, the actually created suffixed project | PR #9699, commit `89761750be0efe09c98e7fdbb0c5b802fdc6aaea` |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| cli-07 | LOW | P3 | VALID BUT CURRENTLY UNREACHABLE | `cli/pkg/cmdutil/helper.go:515-516` compares the `org` argument but constructs with `h.Org`; current callers first synchronize `h.Org`, so add a unit test with differing values before changing behavior. |
| cli-10 | LOW | P3 | VALID BUT CURRENTLY UNREACHABLE | `cli/pkg/deviceauth/authenticator.go:84` still manually `QueryEscape`s a value that `url.Values.Encode` escapes again, but all current callers pass an empty redirect. Remove the manual escape and add a non-empty redirect test. |

Counts: **11 pending, 1 fixed/remove, 2 partial/needs-test, 0 invalid = 14 total**.

## Machine-readable ledger

```text
FIXED|cli.md|cli-01|HIGH|-|-|DeployProject pushed env to wrong project|cli/pkg/local/server.go:454-457; PR #9699 89761750be0efe09c98e7fdbb0c5b802fdc6aaea
PENDING|cli.md|cli-02|MEDIUM|P1|-|PKCE authenticator map race|cli/pkg/local/app.go:59; cli/pkg/local/server.go:825,845,852
PENDING|cli.md|cli-03|MEDIUM|P1|-|Helper client cache races|cli/pkg/cmdutil/helper.go:170-277
PENDING|cli.md|cli-04|MEDIUM|P2|-|Local PKCE/logout open redirect|cli/pkg/local/server.go:815-820,879,897-901
PENDING|cli.md|cli-05|MEDIUM|P1|-|World-readable CLI credentials|cli/pkg/dotrill/dotrill.go:113,362
PENDING|cli.md|cli-06|MEDIUM|P2|-|ListMatchingProjects swallows errors|cli/pkg/local/server.go:727-745
NEEDS_TEST|cli.md|cli-07|LOW|P3|-|GitHelper uses struct org instead of argument|cli/pkg/cmdutil/helper.go:515-516
PENDING|cli.md|cli-08|LOW|P3|-|PKCE callback does not validate state|cli/pkg/local/server.go:835-845
PENDING|cli.md|cli-09|LOW|P3|-|Dead mismatched device-auth interface|cli/pkg/deviceauth/authenticator.go:29-32,80+
NEEDS_TEST|cli.md|cli-10|LOW|P3|-|Device redirect double escaping|cli/pkg/deviceauth/authenticator.go:84
PENDING|cli.md|cli-11|LOW|P3|-|PKCE verifier modulo bias|cli/pkg/pkce/authenticator.go:150-159
PENDING|cli.md|cli-14|LOW|P2|-|App-start telemetry silently dropped|cli/pkg/local/app.go:536-556
PENDING|cli.md|cli-12|LOW|P2|-|Invisible env upload|cli/cmd/project/deploy.go:309,437-451
PENDING|cli.md|cli-13|LOW|P2|-|PollServer health-check race|cli/pkg/local/app.go:480-529
```
