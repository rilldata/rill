# Admin Auth & RBAC — Code Review

**Scope:** `admin/server/auth/{auth,claims,device_code,handlers,middleware,mcp_oauth,pkce}.go`, `admin/server/{roles,users,usergroups,magic_tokens,runtime_jwt,runtime_proxy,sudo,page_token}.go`, `admin/server/cookies/cookies.go`, `admin/{permissions,auth_token,auth_code,user}.go`, plus supporting `admin/server/server.go` (cookie/JWT helpers), `admin/pkg/authtoken`  |  **Health:** fair

## Overview

This area implements the cloud control plane's authentication (Auth0/OIDC login, device code, PKCE authorization code, refresh tokens, magic tokens), the token model (`authtoken`), session cookies, RBAC permission composition (org/project roles), superuser/sudo, runtime JWT minting, and the runtime reverse proxy. The overall structure is sound: permissions are composed by OR-union of roles, tokens are stored as SHA-256 hashes, redirect URLs are validated against an allow-list on the main login/logout paths, cookies are `HttpOnly`/`SameSite`, PKCE is S256-only with single-use codes that are deleted before validation, and the auth endpoints are rate-limited.

After adversarial verification, the most substantive issue is the magic-token attribute capture (admin-auth-rbac-03): a shareable/embeddable magic link inherits the creating admin's `admin: true` attribute and custom attributes and replays them into the runtime JWT, which can subvert attribute-based row/field security policies. The remaining findings are lower-severity correctness, defense-in-depth, and hardening items. Several findings that the prior reviewer rated medium are latent rather than presently exploitable (per-request claims are not cached across requests, and the shipped roles couple prod/dev), and have been downgraded accordingly.

## Findings

### admin-auth-rbac-03 [MEDIUM] Magic tokens capture the creator's `admin` flag and custom attributes and replay them into the runtime JWT
- **Category:** security  **Confidence:** high  **Location:** `admin/server/magic_tokens.go:90-100` (capture), `admin/server/server.go:380-386` (`jwtAttributesForUser` sets `"admin": projectPermissions.ManageProject`), `admin/server/runtime_jwt.go:113-119` (replay: `attr = mdl.Attributes`)
- **Problem:** When a project admin/editor issues a magic token, `jwtAttributesForUser` is invoked with the creator's live project permissions (`projPerms` from `magic_tokens.go:45`), so the stored attribute map contains `admin: true` (because `ManageProject` is true for admins/editors) plus all of the creator's custom org-member attributes. When the magic token is later used, `issueRuntimeToken` sets the runtime JWT attributes directly from `mdl.Attributes`, so the JWT carries `admin: true`. The resource allow-list in `securityRulesFromMagicAuthToken` restricts *which resources* are reachable but does not neutralize attribute-based row/field security policies. The code comments (`magic_tokens.go:93-94`) explicitly acknowledge this risk.
- **Failure scenario:** A metrics view uses a security policy gated on `{{ .user.admin }}` (a documented Rill pattern for row-level security) or on a custom attribute. An admin creates a shareable magic link (often embedded in an iframe or sent to external viewers). Anyone holding that link inherits `admin: true` (and the admin's custom attributes) and therefore sees unfiltered rows/fields the policy intended to hide from non-admins.
- **Recommendation:** Strip or force `admin: false` (and drop privilege-bearing custom attributes) when capturing attributes for a magic token, or compute magic-token attributes from an explicitly non-admin permission set rather than the creator's live permissions.

**Verification:** CONFIRMED — the capture→store→replay chain is intact: `magic_tokens.go:95` calls `jwtAttributesForUser(..., projPerms)`, which sets `"admin": projectPermissions.ManageProject` (`server.go:385`) plus all custom attributes; the map is persisted (`InsertMagicAuthToken` at `auth_token.go:207-221`) and replayed verbatim as `attr = mdl.Attributes` in `runtime_jwt.go:119`, with no admin-stripping anywhere. Note the creator must have `ManageProject=true` (project admin role) for `admin:true` to be captured.

**Manual Reproduction Steps (for QA):**
1. In a Rill Cloud project, add a metrics view with a security policy that gates row or field access on the user attribute `admin` (for example, only show a sensitive dimension/rows when `{{ .user.admin }}` is true).
2. Log in as a user with the project `admin` role and open a dashboard built on that metrics view. Confirm you see the admin-only rows/fields.
3. As that admin, create a shareable magic link for the dashboard (the "Share" / public URL / embed link feature that issues a magic auth token).
4. Open the magic link in a private/incognito window (no login), or hand it to a non-admin colleague.
5. Observe that the dashboard opened via the magic link shows the admin-only rows/fields that the security policy was meant to hide from non-admins — confirming the link inherited `admin: true`.

### admin-auth-rbac-01 [LOW] Handlers mutate the shared cached permission objects returned by `Claims`
- **Category:** security  **Confidence:** high  **Location:** `admin/server/auth/claims.go:201-242` (and `:246-279`); mutating callers at `admin/server/runtime_proxy.go:83-84`, `admin/server/projects.go:287-300`, `admin/server/organizations.go:75-77`, `admin/server/bookmarks.go:58-60,88`
- **Problem:** `authTokenClaims.ProjectPermissions` and `organizationPermissionsUnsafe` cache and return the *same* `*adminv1.ProjectPermissions` / `*adminv1.OrganizationPermissions` pointer they store in `projectPermissionsCache` / `orgPermissionsCache`. Several handlers take that pointer and write to it (`permissions.ReadProject = true`, `perms.Guest = true`, etc.) to apply public-project / superuser-force elevation. Every write persists into the claims cache, and the mutation happens without holding the claims mutex the getters otherwise use.
- **Failure scenario:** Not directly exploitable today: `newAuthTokenClaims` is constructed fresh per request in `middleware.go:201`, so the blast radius is a single request, and the flags mutated (public/force read + guest) are not re-read for an authorization decision later in the same handler. It is a latent privilege-escalation footgun (if the `Claims` wrapper is ever memoized across requests, an anonymous hit on a public project would leave `ReadProd`/`ReadProject` flipped true in the cached object) and a latent data race if a streaming handler ever touches claims from two goroutines.
- **Recommendation:** Have the getters return a defensive copy (e.g. `proto.Clone`) of the cached permissions, or construct the public/force elevation as a locally-owned permission object instead of writing through the returned pointer.

**Verification:** PLAUSIBLE — `ProjectPermissions` returns the same cached pointer (`claims.go:205-207,240-241`) and callers write through it (`runtime_proxy.go:83-84`, `projects.go:290-300`, `organizations.go:75`, `bookmarks.go:59-60`), but claims are built fresh per request (`middleware.go:201` `newAuthTokenClaims`), so it is a latent footgun/race rather than a present exploit as the finding itself concedes.

**Manual Reproduction Steps (for QA):**
Not user-reproducible today — the mutation is confined to a single request and no user-facing decision re-reads the flipped flag. Confirm in code: `admin/server/auth/claims.go:205-207` and `:240-241` return the cached `*adminv1.ProjectPermissions` pointer directly; then `admin/server/runtime_proxy.go:83-84` and `admin/server/projects.go:290-300` write `permissions.ReadProd = true` etc. through that shared pointer. Verify no `proto.Clone`/copy sits between the getter and the writes.

### admin-auth-rbac-02 [LOW] Revoked / logged-out / rotated tokens stay valid for up to 10s (cache is never evicted on revoke)
- **Category:** security  **Confidence:** high  **Location:** `admin/auth_token.go:274-312` (cache), `admin/auth_token.go:249-267` (`RevokeAuthToken`); revoke call sites in `admin/server/users.go:335,363,456`, `admin/server/magic_tokens.go:252`, `admin/server/auth/handlers.go:366,426,435,537,569`, `admin/server/auth/pkce.go:371`
- **Problem:** `ValidateAuthToken` caches successful validations (keyed by SHA-256 of the token) for `_authCacheTTL = 10s`. `PurgeAuthTokenCache` exists to flush that cache but is called only from tests (`admin/server/users_test.go:39,51`), never from any production revocation path — every revoke path (`RevokeAuthToken`, `DeleteUserAuthToken`, `DeleteAllUserAuthTokens`, magic-token revoke, logout, refresh-token rotation) only deletes the DB row. That the tests must manually purge the cache to observe revocation confirms the gap.
- **Failure scenario:** A user (or admin) revokes a leaked token, logs out, or calls "revoke all sessions" after a compromise. The stolen token continues to authenticate successfully for up to 10 seconds on any admin replica that has it cached, because validation short-circuits on the cache before hitting the DB. Impact is bounded (≤10s, per-replica, and only where the token was recently used) but the revocation control is not immediate.
- **Recommendation:** On any revoke/delete, evict the specific token from `s.authCache` (add a `Remove(hash)` helper keyed like `ValidateAuthToken`) or call `PurgeAuthTokenCache`. At minimum, do not serve a cache hit for a token deleted in the same flow (e.g. refresh-token rotation).

**Verification:** CONFIRMED — `ValidateAuthToken` caches successful validations for `_authCacheTTL` (`auth_token.go:288-312`); `PurgeAuthTokenCache` (`auth_token.go:269-272`) is called only from `users_test.go:39,51` (grep confirms zero production call sites), and `RevokeAuthToken` (`auth_token.go:249-267`) only deletes the DB row without touching the cache.

**Manual Reproduction Steps (for QA):**
1. Create a Rill service token (or personal access token) via CLI/UI and note the token string.
2. Make an authenticated admin API call with the token (e.g. `curl -H "Authorization: Bearer <token>" <admin-host>/v1/users/me` or any endpoint) and confirm it succeeds — this populates the 10s validation cache.
3. Immediately revoke/delete that same token (CLI `rill token revoke`, or the UI "Delete token" / "Revoke" action).
4. Within ~10 seconds, repeat the same authenticated API call with the revoked token against the same admin replica.
5. Observe the call still succeeds (200) for up to 10 seconds after revocation, instead of returning 401 immediately.

### admin-auth-rbac-04 [LOW] Runtime proxy gates every deployment (prod, dev, branch) on `ReadProd`, and public projects expose branch deployments anonymously
- **Category:** security  **Confidence:** low  **Location:** `admin/server/runtime_proxy.go:41-93`; branch route at `admin/server/server.go:221`
- **Problem:** The proxy resolves the primary deployment or a branch deployment (`FindDeploymentsForProject(..., branch)`) and then gates access solely on `permissions.ReadProd` (after forcing `ReadProd = true` for `proj.Public`). It never consults `ReadDev` even when the resolved deployment runs in the `dev` environment. The issued runtime JWT always grants the base read permissions (`ReadAPI/ReadMetrics/ReadObjects/UseAI`) regardless of environment (only the *elevated* `ReadInstance`/`Manage*` permissions in `runtime_jwt.go:178-187` are environment-aware).
- **Failure scenario:** For a public project, the `/branch/{branch}/runtime/...` route forces `ReadProd = true` for anonymous callers and then serves the branch (potentially in-development) deployment's data anonymously — broader than the "public prod dashboard" intent. Note: the parallel "a custom role granting `ReadProd` but not `ReadDev` reaches dev" concern is **not reachable today** — the three shipped project roles (admin/collaborator/viewer, `migrations/0003.sql:40-42`) all grant `read_prod` and `read_dev` together, and roles are not user-customizable — so this reduces to the public-project branch-exposure question, which is partly a design choice (a public project arguably exposes all its deployments).
- **Recommendation:** Select the permission to check based on `depl.Environment` (`ReadProd` for prod, `ReadDev` for dev/branch), and decide explicitly whether public projects should expose non-primary/branch deployments at all.

**Verification:** CONFIRMED — the proxy resolves a branch deployment (`runtime_proxy.go:52-61`) and gates solely on `permissions.ReadProd` after forcing `ReadProd=true` for `proj.Public` (`runtime_proxy.go:82-93`), never consulting `ReadDev`/environment; the branch route exists at `server.go:221`. The "custom role" escalation is correctly ruled out — `migrations/0003.sql:40-42` shows all three built-in roles couple `read_prod` and `read_dev`.

**Manual Reproduction Steps (for QA):**
1. Create a Rill Cloud project and mark it Public.
2. Push a non-primary branch that produces a separate (e.g. dev/in-development) branch deployment for that project.
3. In an incognito window with no authentication, request the branch runtime proxy route: `curl <admin-host>/v1/orgs/{org}/projects/{project}/branch/{branch}/runtime/v1/instances/.../<some-read-endpoint>`.
4. Observe the anonymous request is served the branch deployment's data (200) rather than being restricted to the primary/prod deployment.

### admin-auth-rbac-13 [LOW] `CookieRefreshMiddleware` logs the user out on any token-validation error, including transient ones
- **Category:** product-improvement  **Confidence:** high  **Location:** `admin/server/auth/middleware.go:102-112`
- **Problem:** On every request carrying a session cookie, `CookieRefreshMiddleware` calls `ValidateAuthToken` and, on *any* error, deletes `cookieFieldAccessToken` from the session and saves the cookie. The comment acknowledges the error "can be due to expiry or transient context issues," yet treats all of them as a reason to clear the session. Combined with admin-auth-rbac-07 (transient DB errors are negatively cached for up to 10s), a brief admin-DB blip evicts a valid browser session and forces a re-login.
- **Failure scenario:** The admin DB has a transient hiccup while a logged-in user is browsing. `ValidateAuthToken` returns a non-context error, the middleware wipes the auth cookie, and the user is silently logged out even though their token is valid.
- **Recommendation:** Only clear the cookie when the token is definitively invalid (expired, not found, hash mismatch). On transient/infrastructure errors, leave the cookie intact and continue.

**Verification:** CONFIRMED — `CookieRefreshMiddleware` deletes `cookieFieldAccessToken` on *any* non-nil error from `ValidateAuthToken` (`middleware.go:102-107`), with no distinction between definitive-invalid and transient errors; the comment at `:104` acknowledges the error "can be due to expiry or transient context issues."

**Manual Reproduction Steps (for QA):**
1. Log in to Rill Cloud in a browser so a valid session cookie is set.
2. In a controlled/staging environment, induce a transient admin-DB failure (e.g. briefly stop or firewall the admin Postgres, or use a fault-injection proxy) while the session is still valid.
3. Reload any authenticated Rill Cloud page during the outage so the request passes through `CookieRefreshMiddleware`.
4. Restore the DB and reload again.
5. Observe you have been logged out (redirected to login) even though the token had not expired — confirming a transient error wiped the cookie. (If DB fault injection is not feasible, confirm in code at `admin/server/auth/middleware.go:102-107`.)

### admin-auth-rbac-07 [LOW] `ValidateAuthToken` negatively caches transient errors, rejecting valid tokens for up to 10s
- **Category:** correctness  **Confidence:** high  **Location:** `admin/auth_token.go:298-311`
- **Problem:** After a cache miss, any non-context error from `validateAuthTokenUncached` is cached alongside the (nil) token for `_authCacheTTL`. Only errors matching `ctx.Err()` bypass caching; a transient DB error with a live context (`err != nil && errors.Is(err, nil)` is false) falls through and is memoized for 10s.
- **Failure scenario:** The admin DB briefly becomes unavailable while validating a token. The failure is cached under that token's hash; even after the DB recovers, every request presenting that exact token is rejected for up to 10 seconds instead of being re-validated. (This is the root cause behind admin-auth-rbac-13's transient logout.)
- **Recommendation:** Only cache definitive negative results (malformed token, not found, expired, hash mismatch). Do not cache infrastructure/transient errors.

**Verification:** CONFIRMED — after a cache miss, `ValidateAuthToken` only bypasses caching when `errors.Is(err, ctx.Err())` (`auth_token.go:300`); every other error, including transient DB errors, is stored via `s.authCache.Add(cacheKey, authCacheEntry{token, err, ...})` (`auth_token.go:304-309`) and returned for the remaining TTL.

**Manual Reproduction Steps (for QA):**
1. In a controlled/staging environment, obtain a valid token and prime a failure: induce a transient admin-DB error (stop/firewall Postgres or use a fault-injection proxy).
2. Make an authenticated API call with the valid token during the outage so the failed validation is cached: `curl -H "Authorization: Bearer <token>" <admin-host>/v1/users/me` (expect an error).
3. Immediately restore the DB.
4. Within ~10 seconds, repeat the same call with the same valid token against the same replica.
5. Observe it is still rejected for up to 10 seconds even though the DB has recovered and the token is valid. (If fault injection is not feasible, confirm in code at `admin/auth_token.go:298-311`.)

### admin-auth-rbac-05 [LOW] Missing `return` after redirecting anonymous users in `handleAuthorizeRequest`
- **Category:** bug  **Confidence:** high  **Location:** `admin/server/auth/handlers.go:657-666`
- **Problem:** When `claims.OwnerType() == OwnerTypeAnon`, the handler calls `http.Redirect(...)` but does not `return`. Execution falls through to `if claims.OwnerType() != OwnerTypeUser { http.Error(...) }` (anon is not a user), which writes a second response header after the redirect already committed the status.
- **Failure scenario:** An unauthenticated OAuth `/auth/oauth/authorize` request triggers the 307 redirect to `/auth/login`, then also runs `http.Error(w, "only users can be authorized", 400)`, producing a superfluous-`WriteHeader` log line and appending an error body to the redirect. Mostly benign because browsers follow the redirect, but it is a clear control-flow bug that would misbehave with a buffering `ResponseWriter`.
- **Recommendation:** Add `return` immediately after the `http.Redirect` call.

**Verification:** CONFIRMED — in `handleAuthorizeRequest`, the anonymous branch calls `http.Redirect(...)` (`handlers.go:661`) with no `return`, so execution falls through to `if claims.OwnerType() != OwnerTypeUser { http.Error(w, "only users can be authorized", 400) }` (`handlers.go:663-666`), writing a second response after the redirect.

**Manual Reproduction Steps (for QA):**
1. In a browser or via curl with no Rill session, hit the OAuth authorize endpoint: `curl -sv "<admin-host>/auth/oauth/authorize?client_id=x&redirect_uri=y&response_type=code"` (with cookies disabled / no auth).
2. Inspect the raw response and the admin service logs.
3. Observe both a 307 redirect to `/auth/login` and the follow-on `only users can be authorized` error body, plus a superfluous-`WriteHeader` warning in the server logs — confirming the missing `return`.

### admin-auth-rbac-09 [LOW] `DeleteUser` looks up the account before the authorization check, enabling email enumeration
- **Category:** security  **Confidence:** medium  **Location:** `admin/server/users.go:525-538` (`DeleteUser`); minor: `admin/server/users.go:723-747` (`findUserAuthTokenFuzzy` exact-ID path)
- **Problem:** `DeleteUser` calls `FindUserByEmail` before evaluating the permission gate. A non-existent email returns the raw `ErrNotFound`, which `mapGRPCError` (`server.go:441-456`) maps to `codes.NotFound`, while an existing email that the caller may not delete returns `codes.PermissionDenied`. The two distinct codes form an enumeration oracle for registered emails, reachable even by anonymous callers since the lookup precedes any auth check.
- **Failure scenario:** Any caller enumerates which emails are registered Rill users by observing `NotFound` vs `PermissionDenied`. (The analogous `findUserAuthTokenFuzzy` exact-ID lookup is far weaker because token IDs are 128-bit UUIDs and not guessable, so it is noted only for completeness.)
- **Recommendation:** Perform the coarse authorization check before the lookup where possible, or normalize the response so "missing" and "forbidden" return the same code.

**Verification:** CONFIRMED — `DeleteUser` calls `FindUserByEmail` (`users.go:528`) and returns its raw error before the permission gate at `users.go:533-538`; a missing email yields `ErrNotFound`→`codes.NotFound` (`mapGRPCError`, `server.go:454-456`) while an existing-but-forbidden email yields `codes.PermissionDenied`, giving distinct responses.

**Manual Reproduction Steps (for QA):**
1. As a non-superuser (or unauthenticated) caller, invoke the admin `DeleteUser` API with an email you know is NOT registered: e.g. `grpcurl`/REST call to DeleteUser with `email=doesnotexist@example.com`. Note the returned gRPC code is `NotFound`.
2. Repeat with an email you know IS registered but that you are not allowed to delete (any other user's email). Note the returned code is `PermissionDenied`.
3. Observe the two different codes let a caller distinguish registered from unregistered emails (an enumeration oracle) without any account privileges.

### admin-auth-rbac-06 [LOW] Token secret comparison is not constant-time
- **Category:** security  **Confidence:** high  **Location:** `admin/auth_token.go:331,352,370,387` (`bytes.Equal(model.SecretHash, parsed.SecretHash())`)
- **Problem:** All four token types validate the presented secret with `bytes.Equal` over the SHA-256 hashes, which short-circuits on the first differing byte and is not timing-safe.
- **Failure scenario:** Exploitability is minimal because the compared values are SHA-256 hashes (a timing leak reveals hash bytes, not the secret, and SHA-256 is not invertible), but this is security-critical credential verification where constant-time comparison is the expected norm and cheap to adopt.
- **Recommendation:** Use `crypto/subtle.ConstantTimeCompare` (or `hmac.Equal`) for the secret-hash comparison.

**Verification:** CONFIRMED — all four token types compare secrets with `bytes.Equal(<model>.SecretHash, parsed.SecretHash())` (`auth_token.go:331,352,370,387`), which short-circuits on the first differing byte and is not constant-time. Exploitability is minimal since the compared values are non-invertible SHA-256 hashes, as the finding notes.

**Manual Reproduction Steps (for QA):**
Not user-reproducible through the product (a timing side-channel over SHA-256 hashes is not observable via normal UI/API use and leaks no usable secret). Confirm in code: `admin/auth_token.go:331,352,370,387` use `bytes.Equal` for secret-hash comparison instead of `crypto/subtle.ConstantTimeCompare`/`hmac.Equal`.

### admin-auth-rbac-12 [LOW] `validPageSize` imposes no maximum, allowing unbounded list queries
- **Category:** performance  **Confidence:** high  **Location:** `admin/server/page_token.go:70-75`
- **Problem:** `validPageSize` substitutes a default when `pageSize == 0` but otherwise returns `int(pageSize)` with no upper bound. It is the only size gate for ~17 list endpoints (`ListMagicAuthTokens`, `ListOrganizationMemberUsers`, `ListProjectsForOrganization`, `ListUsergroupMemberUsers`, etc.).
- **Failure scenario:** An authenticated caller passes `page_size = 4294967295`; the endpoint issues a single query with that `LIMIT`, defeating pagination and, for a large tenant (thousands of members/tokens/projects), loading the entire table into memory in one response.
- **Recommendation:** Clamp `validPageSize` to a sane maximum (e.g. 1000) in addition to the default.

**Verification:** CONFIRMED — `validPageSize` substitutes `_defaultPageSize` only when `pageSize == 0` and otherwise returns `int(pageSize)` with no upper clamp (`page_token.go:70-75`).

**Manual Reproduction Steps (for QA):**
1. Authenticate against an org with many members/tokens/projects (or seed a test org with thousands of rows in one of the paged tables).
2. Call a paged list endpoint with a huge page size, e.g. REST/grpcurl to `ListOrganizationMemberUsers` (or `ListProjectsForOrganization`, `ListMagicAuthTokens`) with `page_size = 4294967295`.
3. Observe the endpoint returns the entire table in a single response (no cap applied), rather than clamping to a reasonable maximum — confirming pagination can be bypassed.

### admin-auth-rbac-08 [LOW] Device-code request embeds an unvalidated user-supplied `redirect` into the verification URL
- **Category:** security  **Confidence:** low  **Location:** `admin/server/auth/device_code.go:79-90`
- **Problem:** `handleDeviceCodeRequest` copies `values.Get("redirect")` straight into the `verification_uri_complete` query without running it through `validateRedirectURL`, unlike `authStart` (`handlers.go:192`) and `authLogoutProvider` (`handlers.go:597`), which validate against the external/frontend host or a registered custom domain.
- **Failure scenario:** If the device-verification frontend navigates to this `redirect` after the user approves the code, a device flow initiated with `redirect=https://evil.com` yields an open redirect off a trusted Rill URL. Exploitability is low: the completed URL is returned to the flow initiator (typically the CLI) rather than an unrelated victim, and an attacker who can influence the frontend URL could craft the verification URL directly regardless of this endpoint. It remains an inconsistency worth closing for defense-in-depth.
- **Recommendation:** Validate the `redirect` parameter server-side with the same `validateRedirectURL`/`IsSafeRedirectURL` logic before embedding it.

**Verification:** CONFIRMED — `handleDeviceCodeRequest` copies `values.Get("redirect")` directly into the `verification_uri_complete` query (`device_code.go:80-90`) with no `validateRedirectURL` call, and the device verify frontend navigates to it unvalidated (`web-admin/src/routes/-/auth/device/+page.svelte:30-31` does `window.location.href = decodeURIComponent(redirectURL)`). Exploitability is low as the finding notes (URL returned to the flow initiator, typically the CLI).

**Manual Reproduction Steps (for QA):**
1. Initiate a device-code request with an attacker-controlled redirect: `curl -X POST "<admin-host>/auth/deviceauth/authorize" -d "client_id=<id>&scope=full_account&redirect=https://evil.example.com"`.
2. Inspect the JSON response `verification_uri_complete` field and confirm it embeds `redirect=https://evil.example.com` unchanged (no host allow-list validation).
3. Open that verification URL in a browser, log in, and approve the device code; observe the device page navigates to `https://evil.example.com` — an open redirect off a trusted Rill URL.

### admin-auth-rbac-10 [LOW] Org member-managers can grant project-admin indirectly via usergroup membership
- **Category:** product-improvement  **Confidence:** high  **Location:** `admin/server/usergroups.go:595-654` (`AddUsergroupMemberUser`, acknowledgement at `:611-612`)
- **Problem:** Adding a user to a usergroup only requires `ManageOrgMembers` on the org (plus `ManageOrgAdmins` if the group holds an *org*-admin role). A usergroup may hold an admin *project* role on a project the actor does not administer, and that case is not checked. The code comment explicitly accepts this ("not a big leak").
- **Failure scenario:** An org member-manager who is deliberately not a project admin adds themselves (or a colleague) to a usergroup that is project-admin on a sensitive project, gaining admin-level access to it.
- **Recommendation:** When adding members to a group, verify the actor holds the group's effective project roles (or at least `ManageProjectAdmins` for any admin-bearing project), or surface the group's project roles in the UI so the escalation is explicit.

**Verification:** CONFIRMED — `AddUsergroupMemberUser` only checks `ManageOrgMembers` (`usergroups.go:596`) plus `ManageOrgAdmins` when the group holds an org-admin role (`usergroups.go:608-609`); it never checks the group's *project* roles, and the code comment at `usergroups.go:611-612` explicitly accepts this ("not a big leak of permissions").

**Manual Reproduction Steps (for QA):**
1. Set up an org with a usergroup that has been granted the `admin` role on a specific sensitive project, but which holds no org-admin role.
2. Log in as a user who has org member-management rights (`ManageOrgMembers`) but is deliberately NOT an admin on that sensitive project.
3. As that user, add yourself (or a colleague) to the project-admin usergroup via the "Add member to group" UI or the `AddUsergroupMemberUser` API.
4. Confirm the add succeeds, and then open the sensitive project — observe you now have admin-level access to a project you could not previously administer.

### admin-auth-rbac-11 [LOW] First-user-becomes-superuser bootstrap can race under concurrent first sign-ups
- **Category:** concurrency  **Confidence:** low  **Location:** `admin/user.go:168-192` (`CheckUsersEmpty` then `InsertUser` with `Superuser: isFirstUser`)
- **Problem:** The first user ever created is granted superuser based on `CheckUsersEmpty` inside a transaction opened with `NewTx(ctx, false)`. Depending on the DB isolation level, two concurrent first-time sign-ups could both observe an empty users table and both be inserted as superusers.
- **Failure scenario:** On a freshly provisioned deployment, two users complete OIDC login simultaneously and both are granted superuser. The window is extremely narrow (only the very first sign-ups on a brand-new install, typically the operator themselves), so practical impact is low, but the consequence (unintended superuser) is high.
- **Recommendation:** Enforce the invariant at the DB level (partial unique constraint / advisory lock), or gate superuser bootstrap on an explicit configuration signal rather than emptiness.

**Verification:** PLAUSIBLE — `CheckUsersEmpty` then `InsertUser` with `Superuser: isFirstUser` runs inside `NewTx(ctx, false)` (`user.go:168-189`), which does a plain `BeginTxx(ctx, nil)` with no explicit isolation level or locking (`postgres/tx.go:22`); under the default Read Committed isolation two concurrent first sign-ups could both observe an empty table. The window is extremely narrow (only the very first sign-ups on a brand-new install), so it stays PLAUSIBLE.

**Manual Reproduction Steps (for QA):**
Not reliably user-reproducible — it requires two OIDC logins to race within the same instant on a freshly provisioned, empty-users deployment. Confirm in code: `admin/user.go:174-185` (`CheckUsersEmpty` gating `Superuser: isFirstUser`) inside `admin/database/postgres/tx.go:22` (`BeginTxx(ctx, nil)`, default isolation, no `SELECT ... FOR UPDATE`/advisory lock). Verify there is no unique constraint restricting superusers to one row.

## Product Improvements

- **admin-auth-rbac-10** (above) — make indirect project-admin escalation through usergroups explicit or blocked.
- **Token revocation is effectively delayed** (admin-auth-rbac-02) — from a product standpoint, "log out everywhere" / "revoke token" should take effect immediately; wiring cache eviction makes the security control trustworthy.
- **Transient DB blips log users out** (admin-auth-rbac-13/07) — a brief admin-DB hiccup should not evict valid browser sessions.
- **Magic-token secrets are stored recoverable in the DB** (`admin/auth_token.go:210` stores `Secret`, re-materialized in `magic_tokens.go:288-302`) — unlike user/service tokens (hash-only). This is required to re-display shareable links, but it means anyone with `ManageMagicAuthTokens` or DB read access recovers working tokens. Consider surfacing this clearly in the UI and offering a "hide secret after creation" mode.

## New Feature Ideas

- **Immediate session/token revocation with cache eviction + audit** — add a `Remove(tokenHash)` on the auth cache and a "Sessions & Tokens" view listing active tokens (client, last-used, expiry) with one-click revoke that takes effect instantly. Value: trustworthy account security. Effort: M.
- **Scoped / least-privilege magic tokens** — let issuers pick an explicit attribute profile (never `admin`, choose which custom attributes carry) when creating a magic link, closing admin-auth-rbac-03 and giving embedders predictable data access. Value: safer embeds. Effort: M.
- **Constant-time token verification + token-format hardening pass** — adopt `subtle.ConstantTimeCompare` and add fuzz/property tests for `authtoken.FromString`/`String` round-trips and truncation. Value: defense-in-depth. Effort: S.
- **Superuser action audit log** — persist every `Superuser`/`SuperuserForceAccess`/assume-user action (who, whom, when, why) surfaced in an admin UI. Value: compliance and incident response for a support-heavy control plane. Effort: M.
- **Redirect allow-list as a first-class, centrally-tested policy** — route every redirect (`authStart`, logout, device code, PKCE) through one validated helper with a shared test matrix, eliminating gaps like admin-auth-rbac-08. Value: systematic open-redirect prevention. Effort: S.
- **Environment-aware runtime proxy access** — gate the proxy on `ReadProd`/`ReadDev` by `depl.Environment` (admin-auth-rbac-04) and add an org setting for whether public projects may serve non-prod/branch deployments. Value: correct exposure of in-development data. Effort: M.

## Notes & Missing Test Coverage

- `admin/auth_token.go` has no production-path eviction on revoke; `users_test.go` has to call `PurgeAuthTokenCache()` manually to observe revocation, which pins the gap in admin-auth-rbac-02. A test asserting that a revoked token is rejected within the TTL *without* a manual purge would catch it.
- The permission-composition layer (`admin/permissions.go`) and the public-project/force-access elevation paths lack tests asserting that the returned permission objects are not aliased/mutated (admin-auth-rbac-01). `roles_test.go` only covers DTO conversion.
- No test covers magic-token attribute capture/replay (admin-auth-rbac-03); a test asserting an admin-created magic token does not yield `admin: true` in the issued JWT would pin the intended behaviour.
- The runtime proxy (`runtime_proxy.go`) has no unit test around the branch route / environment-based authorization (admin-auth-rbac-04) or the direct-JWT-passthrough vs mint decision.
- Header forwarding in the proxy copies only the first value of each header (`runtime_proxy.go:135-137`, `req.Header.Add(k, v[0])`); harmless today but untested and would silently drop multi-valued request headers.

## Verification Notes

All 11 original findings were re-checked against the code; each survived at least mechanically, and 2 new in-scope findings were added. Net removed: 0. Key adjustments:

- **Reordered by significance.** admin-auth-rbac-03 (magic-token admin-flag replay) is the strongest issue and is now listed first; it is retained at MEDIUM/high after confirming the full capture→store→replay chain (`magic_tokens.go:95` → `IssueMagicAuthToken`/`InsertMagicAuthToken` → `runtime_jwt.go:119`).
- **admin-auth-rbac-01:** confirmed the getters return the cached pointer and that `runtime_proxy.go`, `projects.go`, `organizations.go`, `bookmarks.go` mutate it. Downgraded MEDIUM→LOW because claims are constructed fresh per request (`middleware.go:201`) and the mutated flags are not re-read for authorization within the same handler, so it is latent, not exploitable today (the reviewer's own scenario conceded this).
- **admin-auth-rbac-02:** confirmed via `grep` that `PurgeAuthTokenCache` is called only from `users_test.go`, never in production. Corrected the claim "never called anywhere" (it exists and is test-only). Downgraded MEDIUM→LOW given the bounded, per-replica ≤10s window.
- **admin-auth-rbac-04:** partially refuted. The "custom role with ReadProd-but-not-ReadDev" escalation is not reachable — `migrations/0003.sql:40-42` shows all three built-in project roles couple `read_prod` and `read_dev`, and roles are not user-customizable. Rewritten to focus on the one reachable concern (anonymous exposure of branch deployments on public projects) and downgraded MEDIUM→LOW, confidence medium→low.
- **admin-auth-rbac-07:** traced the `errors.Is(err, ctx.Err())` guard and confirmed non-context errors are cached; bumped confidence medium→high.
- **admin-auth-rbac-09:** confirmed the enumeration oracle by verifying `mapGRPCError` maps `ErrNotFound`→`codes.NotFound` (`server.go:441-456`). Narrowed the finding to `DeleteUser`'s email enumeration (the token-ID path is weak because IDs are UUIDs).
- **admin-auth-rbac-05, -06, -08, -10, -11:** confirmed as written; severities/categories unchanged (05, 06, 08 verified at the cited lines; 10 and 11 confirmed including the code's own acknowledgements). 08 kept at low confidence with the exploitability caveat spelled out.
- **Added admin-auth-rbac-12** (`page_token.go:70-75`): `validPageSize` has no maximum, so ~17 list endpoints accept an unbounded `page_size`. LOW/performance, confirmed no cap exists elsewhere.
- **Added admin-auth-rbac-13** (`middleware.go:102-112`): `CookieRefreshMiddleware` clears the session cookie on any `ValidateAuthToken` error, so a transient DB error (compounded by admin-auth-rbac-07) silently logs valid users out. LOW/product-improvement.
- PKCE (`pkce.go`), device-code issuance (`auth_code.go`), sudo (`sudo.go`), cookies (`cookies.go`), roles (`roles.go`), and MCP OAuth/DCR (`mcp_oauth.go`) were reviewed and found sound (single-use codes deleted before validation, S256-only, superuser-gated sudo, PKCE redirect exact-match against registered URIs); no additional findings.
