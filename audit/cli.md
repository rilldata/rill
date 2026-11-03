# CLI & Local App — Code Review

**Scope:** `cli/cmd` (root, auth, deploy, project, start, sudo/user) and `cli/pkg` (local: app/server/git/slots/project/admin, deviceauth, pkce, dotrill, cmdutil helper/githelper, devtool, browser, web) | **Health:** fair

## Overview

This area implements Rill Developer's local application bootstrap, the device-auth and PKCE login flows, credential storage in `~/.rill`, and the deploy/redeploy flows to Rill Cloud (managed Git, GitHub, and archive). The code is generally readable and the deploy flows carefully guard against pushing to the wrong repo/subpath. The biggest risks are: (1) a deploy path that pushes local `.env` secrets to the *wrong* project when a name collision triggers the retry/suffix logic; (2) shared mutable state (`Helper`'s cached admin client and the local server's PKCE authenticator map) accessed from concurrent local-server request handlers without synchronization; and (3) credential files written world-readable. Most auth logic is a faithful port of PlanetScale's CLI, but the local PKCE HTTP handlers add a localhost open-redirect and a copy-paste validation bug. None of these block normal single-user usage, but several are real correctness/security defects that will surface under concurrency, multi-user machines, or name collisions.

## Findings

### cli-01 [HIGH] `DeployProject` pushes `.env` variables to the wrong project after a name collision

**Status (2026-07-25):** ✅ Fixed and merged in [#9699](https://github.com/rilldata/rill/pull/9699).

- **Category:** correctness (data-loss)  **Confidence:** high  **Location:** `cli/pkg/local/server.go:456`
- **Problem:** `DeployProject` creates the project inside a retrier that appends a numeric suffix on `"name already exists"` errors, so the created project may be `foo-1` (available as `projResp.Project.Name`). But the subsequent `UpdateProjectVariables` call passes `Project: r.Msg.ProjectName` (the original, unsuffixed `foo`) instead of `projResp.Project.Name`. The sibling `RedeployProject` gets this right at line 576 (`Project: projResp.Project.Name`), confirming the intent.
- **Failure scenario:** A user deploys a local folder named `foo` from the UI while a project `foo` already exists in the org (e.g. a re-clone, or a concurrent deploy). The retrier creates `foo-1`, then the local `.env` (which typically contains connector secrets/credentials) is pushed via `UpdateProjectVariables` to the pre-existing `foo` — overwriting that project's variables — while the newly created `foo-1` silently receives none. Local secrets land in a project the user did not intend, and the intended project is misconfigured.
- **Recommendation:** Use `projResp.Project.Name` (and `projResp.Project.OrgName`) for the `UpdateProjectVariables` call, matching `RedeployProject`.

**Verification:** CONFIRMED — `cli/pkg/local/server.go:434-443` retries `CreateProject` with a `-N` suffix on name conflicts, but the `UpdateProjectVariables` call at `server.go:454-457` passes `Org: r.Msg.Org, Project: r.Msg.ProjectName` (the original unsuffixed name) instead of `projResp.Project.Name`. The sibling `RedeployProject` correctly uses `projResp.Project.OrgName`/`projResp.Project.Name` at `server.go:574-576`.

**Manual Reproduction Steps (for QA):**
1. Log in to Rill Cloud: `rill login`. In your org, ensure a project named `foo` already exists (e.g. deploy any project as `foo` first, or create one in the cloud UI).
2. On your machine, create a local project in a folder named `foo` (`mkdir foo && cd foo && rill init` or use an existing project folder renamed to `foo`) and add a `.env` file with a recognizable variable, e.g. `MY_TEST_SECRET=hello123`.
3. Run `rill start` in that folder, open the browser UI at http://localhost:9009, and click Deploy (the UI-driven deploy, not `rill deploy` from the terminal).
4. Observe the deploy creates a project named `foo-1` (because `foo` was taken).
5. Check variables on both projects: `rill env show --project foo` and `rill env show --project foo-1`.
6. Bug confirmed if `MY_TEST_SECRET` appears on the pre-existing `foo` (which you did not deploy to) and is missing on the newly created `foo-1`.

### cli-02 [MEDIUM] Unsynchronized `pkceAuthenticators` map races between `/auth` and `/auth/callback`
- **Category:** concurrency  **Confidence:** high  **Location:** `cli/pkg/local/app.go:59`, `cli/pkg/local/server.go:825,845,852`
- **Problem:** `App.pkceAuthenticators` is a plain `map[string]*pkce.Authenticator`. `authHandler` writes to it (`s.app.pkceAuthenticators[state] = authenticator`), while `authCallbackHandler` reads it and then `delete`s the entry. These are ordinary `http.HandlerFunc`s registered on the runtime's mux; Connect/HTTP handlers each run in their own goroutine with no serialization.
- **Failure scenario:** Two overlapping requests (e.g. two browser tabs each hitting `/auth`, or an `/auth` overlapping a stale `/auth/callback`) perform a concurrent map write / read-write. Go's runtime detects concurrent map access and throws a fatal, unrecoverable `concurrent map writes` panic, taking down the entire local server (and the embedded runtime) rather than returning an error.
- **Recommendation:** Guard the map with a `sync.Mutex` (or use `sync.Map`), taking the lock around the store in `authHandler` and the load+delete in `authCallbackHandler`.

**Verification:** CONFIRMED — `pkceAuthenticators` is declared as a plain `map[string]*pkce.Authenticator` at `cli/pkg/local/app.go:59` (no accompanying mutex in the `App` struct). `authHandler` writes at `cli/pkg/local/server.go:825`, and `authCallbackHandler` reads at `server.go:845` and deletes at `server.go:852`. Both are plain `http.HandlerFunc`s served concurrently by `net/http`, so overlapping requests cause an unsynchronized map read/write, which Go's runtime turns into a fatal panic.

**Manual Reproduction Steps (for QA):**
1. Run `rill start` in any project folder.
2. In a terminal, fire many concurrent requests at the auth endpoint: `for i in $(seq 1 50); do curl -s -o /dev/null "http://localhost:9009/auth" & done; wait` and re-run this a few times.
3. Bug confirmed if the `rill start` process crashes with a `fatal error: concurrent map writes` message in the terminal. (The race is timing-dependent; it may take several attempts. Engineers can confirm deterministically by building with `-race` and repeating step 2.)

### cli-03 [MEDIUM] `Helper.Client()` / `Helper.Telemetry()` mutate shared state from concurrent local-server RPCs
- **Category:** concurrency  **Confidence:** high  **Location:** `cli/pkg/cmdutil/helper.go:170-205` (`Client`), `helper.go:211-277` (`Telemetry`), `helper.go:136-154` (`ReloadAdminConfig`)
- **Problem:** `Client()` reads and writes `h.adminClient`/`h.adminClientHash`, and `Telemetry()` writes `h.activityClient`/`h.activityClientHash`, with no lock (only `gitHelperMu` exists, and it only protects `gitHelper`). `ReloadAdminConfig()` also writes `h.AdminTokenDefault`/`h.AdminURLDefault`. The single `Helper` is shared by every local-server handler, and the web UI fires several RPCs concurrently on load (`GetMetadata`, `GetCurrentUser`, `GitStatus`, `ListOrganizationsAndBillingMetadata`, …), each calling `ch.Client()`.
- **Failure scenario:** On first page load, multiple goroutines observe `h.adminClient == nil` and race to construct and assign it (leaking/duplicating clients), while others read `h.AdminTokenDefault` mid-write in `ReloadAdminConfig`. This is a data race (fires under `-race`) and can intermittently close a client another goroutine is using, or read a torn token/URL.
- **Recommendation:** Protect the admin-client and activity-client caches (and `ReloadAdminConfig`'s writes to the token/URL fields) with a mutex, or build the client once at startup and hand out an immutable reference.

**Verification:** CONFIRMED — `Client()` at `cli/pkg/cmdutil/helper.go:170-205` reads/writes `h.adminClient`/`h.adminClientHash` and calls `ReloadAdminConfig()` (`helper.go:136-154`, which writes `h.AdminURLDefault`/`h.AdminTokenDefault`) with no lock; `Telemetry()` at `helper.go:211-277` writes `h.activityClient`/`h.activityClientHash` unlocked. The only mutex on `Helper` is `gitHelperMu`, used solely for `gitHelper` (`helper.go:510-519`). Local-server handlers (`GetCurrentUser`, `ListProjectsForOrg`, `DeployProject`, etc. in `cli/pkg/local/server.go`) all call `s.app.ch.Client()` concurrently.

**Manual Reproduction Steps (for QA):**
1. Run `rill start` in a project folder while logged in (`rill login` first).
2. Fire concurrent requests that all construct/use the admin client: `for i in $(seq 1 30); do curl -s -o /dev/null http://localhost:9009/local/v1/get-current-user -X POST -H 'Content-Type: application/json' -d '{}' & done; wait` (or simply open the local UI in several tabs simultaneously right after startup).
3. A crash or intermittent "client closed" errors would confirm it, but the race is usually silent in normal builds. Not reliably user-reproducible — confirm in code: `cli/pkg/cmdutil/helper.go:181-204` (unlocked lazy init of `adminClient`), or run a `-race` build of `rill start` and repeat step 2; the race detector reports writes to `h.adminClient`/`h.AdminTokenDefault` from multiple goroutines.

### cli-04 [MEDIUM] Open redirect in the local PKCE auth/logout handlers
- **Category:** security  **Confidence:** high  **Location:** `cli/pkg/local/server.go:815-827,879` (auth), `server.go:897-901` (logout)
- **Problem:** `authHandler` reads `origin := r.URL.Query().Get("redirect")` and stores it verbatim; `authCallbackHandler` then does `http.Redirect(w, r, authenticator.OriginURL, http.StatusFound)` with no validation that the target is a same-origin/relative path. `logoutHandler` similarly forwards an attacker-controlled `redirect` param to the cloud logout URL.
- **Failure scenario:** An attacker page navigates the victim's browser to `http://localhost:9009/auth?redirect=https://evil.example.com`. The local server mints a valid `state`, the user completes the (real) Rill OAuth login, their token is written to `~/.rill`, and the callback redirects the browser to `https://evil.example.com` — a convincing phishing hand-off immediately after a legitimate login. Scope is limited to a running local instance, but the redirect target is fully attacker-controlled.
- **Recommendation:** Reject `redirect` values that are not relative paths (e.g. require a leading `/` and no scheme/host, or validate the host against the known local origin) before storing/using them in both handlers.

**Verification:** CONFIRMED — `authHandler` stores the unvalidated `redirect` query param as the PKCE authenticator's origin at `cli/pkg/local/server.go:815-820` (passed to `pkce.NewAuthenticator`, stored as `OriginURL` at `cli/pkg/pkce/authenticator.go:46`), and `authCallbackHandler` redirects to it verbatim at `server.go:879`. `logoutHandler` forwards the `redirect` param into the cloud logout URL at `server.go:897-901`. No same-origin or relative-path validation exists anywhere in this path.

**Manual Reproduction Steps (for QA):**
1. Run `rill start` in any project folder (server on http://localhost:9009).
2. In a browser, visit `http://localhost:9009/auth?redirect=https://example.com`.
3. Complete the Rill login in the page that opens.
4. Bug confirmed if, after login succeeds, the browser lands on `https://example.com` (an arbitrary external site) instead of the local Rill UI.
5. For the logout variant: visit `http://localhost:9009/logout?redirect=https://example.com` and observe the redirect chain passes the external URL to the cloud logout endpoint (check the `redirect` param in the address bar / network tab).

### cli-05 [MEDIUM] Credentials and config in `~/.rill` are written world-readable
- **Category:** security  **Confidence:** high  **Location:** `cli/pkg/dotrill/dotrill.go:113` (`0o644`), `dotrill.go:362` (`os.ModePerm`)
- **Problem:** Every `DotRill.Set` writes with `0o644`, and the `.rill` directory is created with `os.ModePerm` (`0777`, subject to umask). `credentials.yaml` holds the long-lived access token, the assume-user `backup_token`, and per-env tokens (`tokens.*`). On a shared/multi-user host these are readable by any local user.
- **Failure scenario:** On a multi-user Linux box or a shared build agent, another local account reads `~/.rill/credentials.yaml` and exfiltrates the user's Rill access token, gaining their cloud permissions until the token is revoked.
- **Recommendation:** Write `credentials.yaml` with `0o600` and create `~/.rill` with `0o700` (matching the convention used by `aws`, `gcloud`, `kubectl`). At minimum special-case the credentials file.

**Verification:** CONFIRMED — `DotRill.Set` writes every config file (including `credentials.yaml`, see `CredentialsFilename` at `cli/pkg/dotrill/dotrill.go:19` and `SetAccessToken` at `dotrill.go:153`) with `os.WriteFile(filename, data, 0o644)` at `dotrill.go:113`, and `ResolveFilename` creates `~/.rill` with `os.MkdirAll(dotrill, os.ModePerm)` (0777 before umask) at `dotrill.go:362`.

**Manual Reproduction Steps (for QA):**
1. Remove or back up your existing `~/.rill` directory: `mv ~/.rill ~/.rill.bak`.
2. Run `rill login` and complete the login.
3. Run `ls -la ~/.rill ~/.rill/credentials.yaml`.
4. Bug confirmed if `credentials.yaml` shows `-rw-r--r--` (world-readable) and the `.rill` directory shows `drwxr-xr-x` (world-traversable) — i.e. any other local user account can read your access token.
5. Restore your backup afterwards: `rm -rf ~/.rill && mv ~/.rill.bak ~/.rill`.

### cli-06 [MEDIUM] `ListMatchingProjects` swallows non-inference errors and returns an empty list
- **Category:** correctness  **Confidence:** high  **Location:** `cli/pkg/local/server.go:733-745`
- **Problem:** After `projects, err := s.app.ch.InferProjects(...)`, the handler only handles `err` when it `Is(ErrInferProjectFailed)` (returning an empty list). For any other error the `if err != nil` block falls through without returning — the function then returns `Projects: projects` (nil) with a `nil` error.
- **Failure scenario:** A transient admin API error (network blip, expired token, 500) during `InferProjects` is silently converted into "no matching projects." The UI shows the user that their local folder matches no cloud project and offers to create a new one, instead of surfacing the real error — potentially leading to a duplicate deploy.
- **Recommendation:** Add `return nil, err` at the end of the `if err != nil` block (mirroring the pattern in the other handlers such as `GitStatus`).

**Verification:** CONFIRMED — `ListMatchingProjects` at `cli/pkg/local/server.go:733-740`: the `if err != nil` block only returns for `errors.Is(err, cmdutil.ErrInferProjectFailed)`; any other error falls through to `return connect.NewResponse(... Projects: projects ...)` at `server.go:743-745` with `projects == nil` and a nil error. Contrast `GetCurrentProject` at `server.go:652-659`, which ends the error block with `return nil, err`.

**Manual Reproduction Steps (for QA):**
1. Log in (`rill login`) and run `rill start` in a project folder that was previously deployed to Rill Cloud.
2. Break connectivity to the admin API mid-session (e.g. disconnect from the network, or point `/etc/hosts` at an unreachable IP for the admin host) so admin API calls fail.
3. In the local UI (http://localhost:9009), open the Deploy flow, or call the endpoint directly: `curl -s http://localhost:9009/local/v1/list-matching-projects -X POST -H 'Content-Type: application/json' -d '{}'`.
4. Bug confirmed if the response is a success (`{"projects":[]}` / empty list) with no error, and the UI behaves as if no matching cloud project exists (offering to create a new one) instead of showing a connectivity error.

### cli-07 [LOW] `Helper.GitHelper` constructs the helper with `h.Org` instead of the `org` argument
- **Category:** correctness  **Confidence:** high  **Location:** `cli/pkg/cmdutil/helper.go:510-519`
- **Problem:** The cache-validity check compares the stored helper's org against the `org` parameter (`h.gitHelper.org != org`), but construction uses the struct field: `newGitHelper(h, h.Org, project, localPath)`. The `org` parameter is effectively ignored. Today every caller passes `ch.Org`, so `org == h.Org` and the bug is latent, but the signature invites misuse.
- **Failure scenario:** A future caller passes an org different from the helper's current default (e.g. deploying to org B while `ch.Org` is A). Git credentials would be fetched for `h.Org` (A) rather than the requested org (B), and the cache check would additionally never hit, silently reconstructing on every call.
- **Recommendation:** Construct with the `org` parameter: `newGitHelper(h, org, project, localPath)`; or drop the parameter and document that it always uses `h.Org`.

**Verification:** PLAUSIBLE — the mismatch exists exactly as described: the cache check at `cli/pkg/cmdutil/helper.go:515` compares `h.gitHelper.org != org` (the parameter), but construction at `helper.go:516` passes `h.Org` (the struct field). It is latent: current callers (e.g. `cli/pkg/local/server.go:520,530` after `SetOrg` at `server.go:491-496`) always have `org == h.Org`, so behavior is currently correct.

**Manual Reproduction Steps (for QA):**
Not user-reproducible today — no code path passes an `org` different from `h.Org`. Confirm in code: `cli/pkg/cmdutil/helper.go:515-516` — the comparison uses the `org` parameter while `newGitHelper` is constructed with `h.Org`.

### cli-08 [LOW] `authCallbackHandler` checks `code == ""` twice instead of validating `state`
- **Category:** correctness  **Confidence:** high  **Location:** `cli/pkg/local/server.go:839-843`
- **Problem:** After reading `state := r.URL.Query().Get("state")`, the guard is `if code == "" { http.Error(w, "missing state", ...) }` — a copy-paste of the preceding `code` check. The intended `state == ""` check never runs. An empty `state` currently falls through to the map lookup (which fails with "invalid state"), so behavior is merely a misleading error path today, but the guard is dead.
- **Failure scenario:** A callback with a present `code` but empty `state` skips the "missing state" 400 and instead reports "invalid state"; more importantly the intended validation is silently disabled, so any future logic relying on it would be wrong.
- **Recommendation:** Change the condition to `if state == ""`.

**Verification:** CONFIRMED — `cli/pkg/local/server.go:839-843`: after `state := r.URL.Query().Get("state")`, the guard is `if code == ""` (a duplicate of the check at `server.go:835`), so the "missing state" 400 is dead code and an empty `state` falls through to the map lookup at `server.go:845`, which returns "invalid state".

**Manual Reproduction Steps (for QA):**
1. Run `rill start` in any project folder.
2. Call the callback with a code but no state: `curl -s "http://localhost:9009/auth/callback?code=abc"`.
3. Bug confirmed if the response says `invalid state` (HTTP 400) instead of the intended `missing state` — proving the empty-state guard never fires.

### cli-09 [LOW] `deviceauth.Authenticator` interface doesn't match its implementation (dead interface)
- **Category:** readability (dead-code)  **Confidence:** high  **Location:** `cli/pkg/deviceauth/authenticator.go:28-32`
- **Problem:** The declared interface has `VerifyDevice(ctx) (*DeviceVerification, error)` and `GetAccessTokenForDevice(ctx, v DeviceVerification) (string, error)`, but `*DeviceAuthenticator` implements `VerifyDevice(ctx, redirectURL string)` and `GetAccessTokenForDevice(ctx, v *DeviceVerification) (*oauth.TokenResponse, error)`. The concrete type does not satisfy the interface, and nothing references the interface. It is stale, misleading dead code.
- **Failure scenario:** A reader assumes `DeviceAuthenticator` conforms to `Authenticator` and tries to program against the interface, only to find it uncompilable; the mismatch also hides that the real signatures diverged.
- **Recommendation:** Delete the unused interface or update it to match the implementation (and add a `var _ Authenticator = (*DeviceAuthenticator)(nil)` assertion if it is meant to be honored).

**Verification:** CONFIRMED — the interface at `cli/pkg/deviceauth/authenticator.go:29-32` declares `VerifyDevice(ctx context.Context) (*DeviceVerification, error)` and `GetAccessTokenForDevice(ctx, v DeviceVerification) (string, error)`, while the concrete type implements `VerifyDevice(ctx, redirectURL string)` (`authenticator.go:80`) and `GetAccessTokenForDevice(ctx, v *DeviceVerification) (*oauth.TokenResponse, error)`. A repo-wide grep found no references to `deviceauth.Authenticator` outside its declaration.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — no runtime behavior. Confirm in code: `cli/pkg/deviceauth/authenticator.go:29-32` (interface) vs `authenticator.go:80` (implementation signature); grep the repo for `deviceauth.Authenticator` to verify nothing uses the interface.

### cli-10 [LOW] Device-auth `redirect` form value is double-escaped
- **Category:** correctness  **Confidence:** high  **Location:** `cli/pkg/deviceauth/authenticator.go:84`
- **Problem:** `VerifyDevice` sets `"redirect": []string{url.QueryEscape(redirectURL)}` inside a `url.Values`, but `url.Values.Encode()` (called in `newFormRequest`) percent-encodes each value again — so the server receives a double-encoded redirect. The double-escape is certain; it is latent (severity low) only because every current caller passes `redirectURL == ""` (`Login`/`LoginWithTelemetry` at `cli/cmd/auth/login.go:34`, `cli/cmd/deploy/deploy.go:24`).
- **Failure scenario:** If/when a non-empty redirect is threaded through login (the parameter already exists on `Login`/`LoginWithTelemetry`), the server would receive e.g. `https%3A%2F%2F...` and either reject it or redirect to a malformed URL.
- **Recommendation:** Drop the manual `url.QueryEscape`; let `url.Values.Encode()` handle escaping.

**Verification:** PLAUSIBLE — the double-escape is present: `cli/pkg/deviceauth/authenticator.go:84` sets `"redirect": []string{url.QueryEscape(redirectURL)}` inside a `url.Values` that `newFormRequest` encodes again via `Encode()`. It is latent because every current caller passes an empty `redirectURL` (`cli/cmd/auth/login.go` and `cli/cmd/deploy/deploy.go` pass `""`), so no double-encoded value is ever sent today.

**Manual Reproduction Steps (for QA):**
Not user-reproducible today — the `redirect` value is always empty from all CLI entry points. Confirm in code: `cli/pkg/deviceauth/authenticator.go:81-85` (manual `url.QueryEscape` inside `url.Values`), plus `newFormRequest` in the same file calling `.Encode()`; or add a temporary non-empty redirect and observe the server receives `https%3A%2F%2F...`.

### cli-11 [LOW] PKCE code-verifier generation has modulo bias
- **Category:** security  **Confidence:** high  **Location:** `cli/pkg/pkce/authenticator.go:157`
- **Problem:** `b[i] = charset[temp[i]%byte(len(charset))]` maps a uniform byte (0–255) onto a 66-character set; since 256 is not a multiple of 66, low-index characters are slightly over-represented. Not a practical break given the 43–128 char length, but it is a textbook non-uniform sampling in security-sensitive code.
- **Failure scenario:** Reduced per-character entropy; negligible in isolation but the kind of subtle weakness worth removing from an OAuth secret generator.
- **Recommendation:** Use rejection sampling, or `crypto/rand` with `base64.RawURLEncoding` over random bytes (the RFC 7636 recommended construction).

**Verification:** CONFIRMED — `cli/pkg/pkce/authenticator.go:157` does `b[i] = charset[temp[i]%byte(len(charset))]` over the 66-character charset defined at `authenticator.go:20`; 256 mod 66 = 58, so characters at indices 0-57 are sampled 4/256 of the time and indices 58-65 only 3/256 — a textbook modulo bias. Practical impact is negligible at 43-128 characters.

**Manual Reproduction Steps (for QA):**
Not user-reproducible — the bias is statistical and invisible in normal use. Confirm in code: `cli/pkg/pkce/authenticator.go:155-158` (modulo mapping of a uniform byte onto a 66-char set); an engineer can demonstrate it by generating ~1M verifiers and observing characters `-._~` and high letters underrepresented by ~25 percent relative frequency.

### cli-14 [LOW] `emitStartEvent` returns a nil error instead of the connector error, silently dropping the app-start telemetry event
- **Category:** correctness  **Confidence:** high  **Location:** `cli/pkg/local/app.go:544-547`
- **Problem:** In the loop `for _, c := range connectors { if c.Err != nil { return err } }`, the returned `err` is the value from the preceding `parser.Parse` call, which is `nil` at that point — not `c.Err`. When any connector fails analysis, the function returns early with `nil`, skipping `RecordBehavioralLegacy(BehavioralEventAppStart)`. The caller (`NewApp`, `app.go:356-359`) only logs when the returned error is non-nil, so nothing is logged either: the intended connector error is swallowed and the start event is dropped.
- **Failure scenario:** A project whose connector cannot be analyzed (e.g. a misconfigured or unconfigured connector) starts normally, but the app-start telemetry event is silently dropped and the underlying error is discarded. Because the loop returns on the first failing connector, connector-name reporting is skipped for the whole project.
- **Recommendation:** Return `c.Err` to propagate the error (letting the caller log it), or `return nil` with a comment if skipping the event for projects with connector errors is intentional.

**Verification:** CONFIRMED — `cli/pkg/local/app.go:544-549`: the loop `for _, c := range connectors { if c.Err != nil { return err } }` returns `err`, which at that point holds the nil result of `parser.Parse` at `app.go:539-542`. So a failing connector aborts `emitStartEvent` with a nil error, skipping the `RecordBehavioralLegacy(BehavioralEventAppStart)` call at `app.go:556`; the caller at `app.go:356-359` only logs on non-nil error, so nothing surfaces.

**Manual Reproduction Steps (for QA):**
Not directly user-observable (the only effect is a silently missing telemetry event and a swallowed error). Confirm in code: `cli/pkg/local/app.go:545-548` — `return err` inside the `c.Err != nil` branch returns the nil `err` from `parser.Parse`, not `c.Err`. An engineer can confirm at runtime by starting a project with a misconfigured connector (e.g. a source referencing an undefined connector) under a debugger/log line and observing `emitStartEvent` returns nil early and no `app_start` behavioral event is recorded.

## Product Improvements

### cli-12 [LOW] `.env` push offers no visibility into which variables/secrets are uploaded
- **Category:** product-improvement  **Confidence:** medium  **Location:** `cli/cmd/project/deploy.go:437-451`, `cli/cmd/project/connect_github.go:193-212`, `cli/pkg/local/server.go:448-462`
- **Problem:** All deploy flows silently parse the local `.env` and push every variable to Rill Cloud (`push-env` defaults to `true`). The user gets no preview of which keys — often including secrets — are being transmitted, and no opt-out per-variable.
- **Recommendation:** In interactive mode, print the variable names being uploaded and confirm; document the `--push-env=false` escape hatch prominently.

**Verification:** CONFIRMED — `push-env` defaults to `true` (`cli/cmd/project/deploy.go:309`, `cli/cmd/project/connect_github.go:59`), and all three flows push every parsed `.env` variable with no listing or confirmation: `deploy.go:437-451`, `connect_github.go:195-210`, and the UI-driven `DeployProject` at `cli/pkg/local/server.go:448-462` (which has no opt-out at all).

**Manual Reproduction Steps (for QA):**
1. In a local project folder, create a `.env` containing `SOME_SECRET=supersecret`.
2. Log in (`rill login`) and run `rill deploy` for the project.
3. Observe the terminal output: no listing of the variable names being uploaded and no confirmation prompt appears.
4. Verify the secret was silently pushed: `rill env show --project <project-name>` shows `SOME_SECRET`.

### cli-13 [LOW] `PollServer` relies on a sleep to avoid a false-positive health check when a port is taken
- **Category:** product-improvement  **Confidence:** medium  **Location:** `cli/pkg/local/app.go:480-513`
- **Problem:** The comment admits the initial `250ms` sleep is "a slightly hacky way" to avoid opening the browser against a *different* Rill already bound to the port. A slow machine can still lose this race and open the browser onto an unrelated server; a fast crash can be missed.
- **Recommendation:** Detect the bind failure deterministically (the runtime server already fails to listen when the port is taken) and surface a clear "port already in use" error instead of racing a health check.

**Verification:** CONFIRMED — `PollServer` at `cli/pkg/local/app.go:480-529` sleeps 250ms before the first `/v1/ping` health check, with the in-code comment at `app.go:495-498` admitting it is "a slightly hacky way" to avoid a false positive against another Rill server already bound to the port. The race window is real: if ctx cancellation from the bind failure loses the 250ms race, the browser opens against the wrong server.

**Manual Reproduction Steps (for QA):**
1. In project folder A, run `rill start` (occupies port 9009).
2. In a different project folder B, run `rill start` again in a second terminal.
3. Expected symptom of the race: the second command may briefly pass the health check and open a browser tab pointing at the FIRST instance's UI (project A) before/while erroring about the port. On fast machines it usually errors cleanly; the hack mostly works, which is why this is a robustness note rather than a hard bug.

## New Feature Ideas

- **OS keychain credential storage (S/M).** Store the access token in the OS keyring (Keychain/Credential Manager/libsecret) with a plaintext-YAML fallback. Removes the world-readable-token exposure (cli-05) and matches user expectations from other dev CLIs. Value: security. Effort: M.
- **Named profiles / multi-account support (M).** Let `~/.rill` hold multiple named credential+org profiles (`rill --profile work deploy`), instead of a single global token that assume/unassume juggles via backup keys. Value: users working across orgs/tenants. Effort: M.
- **`rill deploy --dry-run` (S/M).** Print the resolved plan — target org/project, deploy mode (managed/github/archive), git remote/subpath, and the list of `.env` keys to push — without mutating anything. Value: confidence before a destructive/first deploy; also mitigates cli-01/cli-12 surprises. Effort: S–M.
- **Deterministic local-server port handling (S).** On `rill start`, if the port is in use, either fail fast with a clear message or auto-select the next free port and print the chosen URL, replacing the health-check race in `PollServer`. Value: fewer "opened the wrong app" confusions. Effort: S.
- **`rill login --token`/stdin for CI (S).** A first-class non-interactive login that accepts a service token via flag/stdin/env and validates it, instead of relying solely on `--api-token` per-invocation. Value: smoother CI/automation. Effort: S.

## Notes & Missing Test Coverage

- **No tests for the local PKCE HTTP handlers.** `authHandler`/`authCallbackHandler`/`logoutHandler` have no coverage; the open redirect (cli-04), the `state` validation bug (cli-08), and the map race (cli-02) would all be caught by handler tests, including a `-race` test that fires two `/auth` requests concurrently.
- **No coverage for the deploy name-collision path.** The retrier/suffix branch in `DeployProject`/`RedeployProject` is exercised by neither unit nor e2e tests, which is why cli-01 (wrong project for `UpdateProjectVariables`) went unnoticed. A test that forces a `"name already exists"` error and asserts the variables target the created (suffixed) project would lock this down.
- **No concurrency test for `Helper`.** `Client()`/`Telemetry()`/`ReloadAdminConfig()` are only invoked serially in CLI usage; there is no test that exercises them from concurrent goroutines the way the local server does (cli-03).
- **`deviceauth` has only a verifier-charset test.** `VerifyDevice`/`GetAccessTokenForDevice` polling, retry classification, and the `redirect` encoding (cli-10) are untested.
- **Could not fully verify** the admin-server side of the open-redirect (cli-04) and double-escaped redirect (cli-10) — the assessment is based on the CLI side only, since the `redirect` parameter is currently always empty from CLI callers and its server-side contract was out of scope.

## Verification Notes

All 13 original findings were independently re-verified against the code; every one survives and none were removed (`removed_count = 0`). One high-signal miss was added (cli-14). Details:

- **cli-01 (confirmed reachable).** The retrier's `nameConflictRetryErrClassifier` (`server.go:1017`) matches the substring `"name already exists"`, and the admin database returns `"a project with that name already exists in the org"` (`admin/database/postgres/postgres.go:4013`), so the suffix branch is genuinely reachable. `DeployProject` pushes variables to `r.Msg.ProjectName` (`server.go:456`) while the created project may be the suffixed `projResp.Project.Name`; the sibling `RedeployProject` correctly uses `projResp.Project.Name` (`server.go:576`). Scope note: the CLI `rill deploy` path (`cmd/project/deploy.go:424`) creates the project without a suffix retrier, so it is unaffected — the defect is specific to the UI-driven local server.
- **cli-02 (confirmed).** `pkceAuthenticators` is a plain map (`app.go:59`) written by `authHandler` (`server.go:825`) and read/deleted by `authCallbackHandler` (`server.go:845,852`); these run as independent `net/http` handler goroutines, so overlapping requests can trigger Go's fatal `concurrent map writes`.
- **cli-03 (confirmed).** `Client()`, `Telemetry()`, and `ReloadAdminConfig()` mutate `adminClient`/`activityClient`/`AdminTokenDefault`/`AdminURLDefault` with no mutex (only `gitHelperMu` exists, guarding `gitHelper`). Concurrent local-server RPC handlers (and the `trackingHandler` telemetry proxy) call these, so it is a real data race.
- **cli-06 (confirmed).** `ListMatchingProjects` (`server.go:733-745`) only special-cases `ErrInferProjectFailed`; any other error falls through to `return ... Projects: projects` with `projects == nil` and a nil error, silently swallowing the real failure. Contrast `GetCurrentProject` (`server.go:659`) and `GitStatus` (`git.go:38`), which `return nil, err`.
- **cli-07 (confirmed latent).** Every current caller passes `ch.Org`/`s.app.ch.Org`, or calls `SetOrg` immediately before (`server.go:297-303,491-496`), so `org == h.Org` always holds and the wrong-field construction (`helper.go:516`) has no present effect.
- **cli-09 (confirmed dead).** Grep found no references to the `deviceauth.Authenticator` interface anywhere in the tree; callers use `*DeviceAuthenticator` directly (`cli/cmd/auth/login.go:60,74`), and the concrete method signatures do not satisfy the interface.
- **cli-10 (confirmed, confidence raised to high).** The double-escape mechanism is certain; impact remains latent because all `Login`/`LoginWithTelemetry` callers pass an empty `redirectURL`.
- **cli-04 severity (kept medium).** The open redirect is real and the target fully attacker-controlled, but it is scoped to a running localhost instance and requires the victim to complete the full OAuth login; medium reflects that constraint.
- **cli-14 (added, missed by the reviewer).** `emitStartEvent` (`app.go:544-547`) returns the nil `err` from `parser.Parse` instead of `c.Err`, silently dropping the app-start telemetry event (and swallowing the error) whenever a connector fails analysis. Verified that `parser.AnalyzeConnectors` returns `[]*Connector` with an `Err error` field (`runtime/parser/connectors.go:17-24`).

Also examined but intentionally not flagged: `initLocalAdminService` registers the `local_admin` driver into the global `drivers.Drivers` map via `drivers.Register`, which panics on re-registration and is an unsynchronized global write — a genuine fragility for in-process double-init (e.g. tests running two apps), but not reachable in normal single-invocation CLI usage (`rill start` and `rill project validate` each construct one `App` per process), so it was left as a note rather than a finding.
