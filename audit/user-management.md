# User Management — Code Review

**Date:** 2026-08-10
**Scope (backend):** `admin/server/{users,usergroups,organizations,projects,roles}.go`, `admin/{user,permissions}.go`, `admin/server/auth/claims.go`, `admin/server/server.go` (`jwtAttributesForUser`, error mapping), `admin/database/postgres/postgres.go` (member/invite/usergroup/attribute queries), `admin/database/postgres/migrations/{0003,0061,0064}.sql`, `cli/cmd/user/set_attributes.go`
**Scope (frontend):** `web-admin/src/routes/[organization]/-/users/**`, `web-admin/src/features/organizations/user-management/**`, `web-admin/src/features/projects/user-management/**`, `web-common/src/features/users/roles.ts`
**Method:** Static review. Every cited line was opened; role/permission claims were checked against the seeded role rows in the migrations rather than against comments. No builds or tests were run.
**Health:** needs-attention

## Overview

User management spans four layers: the gRPC handlers in `admin/server`, the transactional service helpers in `admin/user.go`, the SQL in `admin/database/postgres`, and two fairly independent frontends (the org-level `/[organization]/-/users` pages and the project-level share dialog). The permission model itself is sound in shape — roles are OR-unioned, permission resolution fails closed, and the `ManageOrgAdmins` / `ManageProjectAdmins` split correctly prevents non-admins from creating or demoting admins in most paths.

The problems cluster in three places.

**First, per-member custom attributes are unvalidated and win over the built-in identity attributes.** `jwtAttributesForUser` applies user-supplied attributes *after* `email`, `domain`, `groups`, and `admin`, and the attribute key validator only checks the character set. Anyone with `ManageOrgMembers` — which migration `0064.sql` grants to the **editor** role, not just admins — can set `admin: true` or a forged `email` on any org member, and that value flows straight into the runtime JWT that backs row- and field-level security policies (UM-01). This is the same class of issue as the already-tracked magic-token finding ([admin-auth-rbac-03](./admin-auth-rbac.md), MF-13), but reached through a documented CLI command (`rill user set-attributes`).

**Second, several membership paths are missing guards that their siblings have.** `DeleteUser` — the self-service account-deletion endpoint — has no last-admin check even though `RemoveOrganizationMemberUser` and `LeaveOrganization` both do, and `users_orgs_roles` cascades on delete, so a user can leave an organization permanently adminless (UM-03). `RemoveUsergroupMemberUser` omits the admin-group guard that `AddUsergroupMemberUser` performs three functions earlier (UM-06). `LeaveOrganization` gates on `ManageOrgMembers`, so the viewers and guests who most need it cannot leave (UM-04). And `SetProjectMemberUsergroupRole` swallows an `ErrNotFound` and then dereferences the nil result, panicking the handler (UM-02).

**Third, the frontend's cache invalidation never matches the queries it renders.** Orval prefixes infinite query keys with the literal `"infinite"`, TanStack matches keys by positional prefix, and all 43 invalidation sites in the two user-management feature directories use the plain key — so every invalidation resolves to zero queries. With `refetchOnMount`, `refetchOnWindowFocus`, and `refetchOnReconnect` all disabled globally, nothing refreshes the list until a page load. This is the reported "removed users stay on screen" bug, and it applies equally to role changes, additions, and group edits (UM-50).

**Fourth, the frontend does server-side work on the client.** The org users, guests, and groups pages all fetch a paginated infinite query and then apply search, role filters, and managed-group filtering in JavaScript over whatever pages happen to be loaded. The backend already accepts `searchPattern` and `role` on `ListOrganizationMemberUsers`; the list page ignores both. Past 50 members, search silently returns wrong results, and a filter that matches nothing on page one renders an empty table with nothing to scroll and therefore no way to load page two (UM-30). Separately, the group edit dialog renames the group before applying member changes and then addresses those changes to the old name, so renaming and re-membering in one save always loses the membership edits (UM-31).

Counts: **6 High · 20 Medium · 17 Low** across 43 findings.

---

## Backend findings

### UM-01 [HIGH] Org member attributes override the reserved identity attributes in the runtime JWT

- **Category:** security  **Confidence:** high
- **Location:** `admin/server/server.go:380-390`; `admin/server/organizations.go:703-737`; `admin/database/postgres/postgres.go:4096-4130`; `admin/database/postgres/migrations/0064.sql:9-10`; CLI surface at `cli/cmd/user/set_attributes.go`
- **Problem:** `jwtAttributesForUser` builds the base attribute map (`name`, `email`, `domain`, `groups`, `admin`) and then copies the member's stored custom attributes over it:

  ```go
  attr := map[string]any{
      "name":   user.DisplayName,
      "email":  user.Email,
      "domain": user.Email[strings.LastIndex(user.Email, "@")+1:],
      "groups": groupNames,
      "admin":  projectPermissions.ManageProject,
  }
  for k, v := range attributes {
      attr[k] = v            // user-supplied keys win
  }
  ```

  The only validation is `validateAttributes` (≤50 keys, alphanumeric+underscore keys, ≤256-char string values) — there is no reserved-key list. `UpdateOrganizationMemberUserAttributes` requires only `ManageOrgMembers`, and migration `0064.sql:10` grants that permission to the **editor** org role (`UPDATE org_roles SET manage_org_members = true WHERE name = 'editor'`), while `manage_org_admins` stays false for editors (`0064.sql:17`). So the escalation does not require an org admin.
- **Failure scenario:** A metrics view uses the documented row-level pattern `{{ .user.admin }}`, or filters on `{{ .user.email }}` / `{{ .user.domain }}`. An org editor runs `rill user set-attributes --email victim@corp.com --attribute admin=true` (the CLI's `parseValue` coerces `"true"` to a real boolean, so the policy comparison succeeds). The next runtime JWT minted for that user carries `admin: true` and every `admin`-gated row and field opens up. Setting `email` or `domain` instead lets an editor impersonate another tenant's identity inside security policies.
- **Recommendation:** Reject the reserved keys (`name`, `email`, `domain`, `groups`, `admin`) in `validateAttributes`, and additionally apply custom attributes *before* the built-ins in `jwtAttributesForUser` so the identity keys always win as defence in depth. Apply the same treatment to `jwtAttributesForService` (`server.go:401-409`), which has the identical override ordering.

**Verification:** CONFIRMED — the override loop is at `server.go:388-390`; `validateAttributes` (`postgres.go:4096-4117`) and `isValidAttributeKey` (`postgres.go:4120-4130`) contain no reserved-name check and accept `admin`/`email`/`domain` unchanged; `organizations.go:715` gates only on `ManageOrgMembers`; `migrations/0064.sql:9-10` grants that to `editor`.

**Manual reproduction (QA):**
1. In a Rill Cloud project, add a metrics view with a security policy gated on `{{ .user.admin }}` so a non-admin sees a restricted subset of rows.
2. Log in as user B (org role `viewer`) and confirm the restricted view.
3. As user A with the org role **editor** (not admin), run `rill user set-attributes --org <org> --email <B's email> --attribute admin=true`.
4. Have user B reload the dashboard.
5. Observe B now sees the admin-only rows.

---

### UM-02 [HIGH] `SetProjectMemberUsergroupRole` panics on a nil result when the group is not yet a project member

- **Category:** bug  **Confidence:** high
- **Location:** `admin/server/usergroups.go:516-527`; `admin/database/postgres/postgres.go:2494-2507`
- **Problem:**

  ```go
  if keepExistingRestrictions {
      ug, err := s.admin.DB.FindProjectMemberUsergroup(ctx, usergroup.ID, proj.ID)
      if err != nil && !errors.Is(err, database.ErrNotFound) {
          return nil, err
      }
      restrictResources = ug.RestrictResources   // ug is nil on ErrNotFound
      resources = ug.Resources
  }
  ```

  `FindProjectMemberUsergroup` returns `nil, parseErr(...)` when the row is missing (`postgres.go:2503-2505`), so tolerating `ErrNotFound` and then dereferencing `ug` is a guaranteed nil-pointer panic. The `req.Role == nil` branch is protected by an early return at `usergroups.go:502-505`, so the crash is reachable only when a role *is* supplied and neither `restrict_resources` nor `resources` is — which is exactly what the CLI's `rill usergroup set-role` sends.
- **Failure scenario:** An admin calls `SetProjectMemberUsergroupRole` for a group that has an org role but has never been added to the project (a plausible mistake, and the natural CLI flow if you assume "set role" is idempotent). The handler panics. Because the gRPC server is served through `vanguardgrpc` as an `http.Handler`, `net/http` recovers the panic and drops the connection, so the client sees a transport error rather than a useful status, and the panic is logged on every occurrence.
- **Recommendation:** Return `codes.NotFound` ("user group is not a member of the project") when `FindProjectMemberUsergroup` reports `ErrNotFound`, matching the `err != nil` handling in `AddProjectMemberUsergroup` (`usergroups.go:451-458`).

**Verification:** CONFIRMED — nil return on missing row at `postgres.go:2503-2505`; swallowed `ErrNotFound` then dereferenced at `usergroups.go:522-526`. The `SetProjectMemberUserRole` sibling (`projects.go:1391-1393`) *does* validate that at least one field is set, and `SetProjectMemberUsergroupRole` has no equivalent guard.

**Manual reproduction (QA):**
1. Create an org user group `analysts` and give it an org role, but do **not** add it to project `P`.
2. Run `rill usergroup set-role --org <org> --project P --group analysts --role viewer` (or call `SetProjectMemberUsergroupRole` with only `role` set).
3. Observe a transport-level failure rather than a `NotFound`, and a panic stack in the admin server logs.

---

### UM-03 [HIGH] `DeleteUser` has no last-admin or billing-contact guard, so an org can be left with zero admins

- **Category:** data-loss  **Confidence:** high
- **Location:** `admin/server/users.go:525-546`; `admin/database/postgres/postgres.go:927-930`; `admin/database/postgres/migrations/0003.sql:44-48`
- **Problem:** `DeleteUser` checks only that the caller is the user themselves or a superuser, then calls `DB.DeleteUser`, which is a bare `DELETE FROM users WHERE id=$1`. `users_orgs_roles.user_id` is declared `REFERENCES users (id) ON DELETE CASCADE`, so every org membership disappears silently. The two sibling paths that remove a user from an org both refuse this: `RemoveOrganizationMemberUser` rejects the last admin (`organizations.go:596-598`) and the org's billing-email holder (`organizations.go:584-586`), and `LeaveOrganization` repeats the same two checks (`organizations.go:764-775`). Account deletion bypasses both.
- **Failure scenario:** The sole admin of an organization deletes their Rill account. The org survives with projects, deployments, and a billing subscription, but has no member holding `ManageOrgMembers` or `ManageOrgAdmins`, so nobody can add one back through the API — recovery requires a superuser. If that user was also the billing contact, `orgs.billing_email` now points at a nonexistent user.
- **Recommendation:** Before deleting, enumerate the user's org memberships and reject (or require an explicit force flag plus org transfer) when the user is the last admin of any org, or is any org's billing email. Reuse `FindOrganizationMemberUserAdminStatus`, which already returns `(isAdmin, isLastAdmin)`.

**Verification:** CONFIRMED — no membership inspection anywhere in `users.go:525-546`; `DeleteUser` SQL at `postgres.go:928`; cascade declared at `migrations/0003.sql:44-48`; the guards that exist elsewhere are at `organizations.go:584-598` and `:764-775`.

**Manual reproduction (QA):**
1. Create org `O` with exactly one member, user A (role `admin`).
2. As A, delete the account (`rill user delete --email <A>` / the account-deletion flow).
3. As a superuser, inspect `O`: it still exists, and `ListOrganizationMemberUsers` returns no admin.
4. Confirm no non-superuser can add a member back.

---

### UM-04 [MEDIUM] `LeaveOrganization` requires `ManageOrgMembers`, so viewers and guests cannot leave

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/organizations.go:755-757`; role rows at `admin/database/postgres/migrations/0003.sql:32-35`, `0061.sql:19-20`, `0064.sql:9-10`
- **Problem:** The handler checks `if !claims.OrganizationPermissions(ctx, org.ID).ManageOrgMembers { PermissionDenied }`. Per the seeded roles, `manage_org_members` is true only for `admin` and (since `0064.sql`) `editor`. The `viewer` and `guest` roles have it false, so the members who are most likely to want to leave an org get "not allowed to remove org members".
- **Failure scenario:** A viewer invited to an org they no longer work with tries to leave and is told they lack permission. (They can in practice self-remove via `RemoveOrganizationMemberUser`, which allows `isSelf` at `organizations.go:579-582` — which also makes `LeaveOrganization`'s stricter check clearly unintentional.)
- **Recommendation:** Replace the permission check with a membership check — confirm the caller is a member of the org (the `FindOrganizationMemberUserAdminStatus` call two lines below already implies it) and keep the existing billing-contact and last-admin guards.

**Verification:** CONFIRMED — the check is at `organizations.go:755`; `viewer` is seeded with `read_org_members`/`manage_org_members` false (`0003.sql:35`), `guest` likewise (`0061.sql:19-20`), and only `editor` was upgraded (`0064.sql:10`). The `isSelf` escape hatch in the parallel handler is at `organizations.go:579-582`.

---

### UM-05 [MEDIUM] Updating attributes for a non-member silently succeeds

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/organizations.go:731`; `admin/database/postgres/postgres.go:2148-2169`
- **Problem:** `UpdateOrganizationMemberUserAttributes` returns `(bool, error)` where the bool is `rows != 0` — deliberately signalling whether a membership row was updated. The handler discards it: `_, err = s.admin.DB.UpdateOrganizationMemberUserAttributes(...)`. If the email belongs to a Rill user who is not a member of this org, zero rows update and the API returns success.
- **Failure scenario:** An admin sets attributes on a user they believe is in the org (typo, or the user was removed). The CLI prints "Successfully updated attributes", the security policy never picks the attributes up, and the misconfiguration is invisible until someone debugs a dashboard.
- **Recommendation:** Return `codes.NotFound` ("user is not a member of the organization") when the bool is false.

**Verification:** CONFIRMED — the DB method computes and returns `rows != 0` (`postgres.go:2160-2168`); the handler discards it at `organizations.go:731`.

---

### UM-06 [MEDIUM] `RemoveUsergroupMemberUser` skips the admin-group guard and ignores pending invites

- **Category:** security  **Confidence:** high
- **Location:** `admin/server/usergroups.go:733-765`; compare `admin/server/usergroups.go:583-654`
- **Problem:** `AddUsergroupMemberUser` refuses to touch a group that holds an admin org role unless the caller has `ManageOrgAdmins`:

  ```go
  if currentRole != nil && currentRole.Admin && !claims.OrganizationPermissions(ctx, group.OrgID).ManageOrgAdmins {
      return nil, status.Error(codes.PermissionDenied, "as a non-admin you are not allowed to edit a group that has an admin role")
  }
  ```

  `RemoveUsergroupMemberUser` performs no such check — it verifies `ManageOrgMembers` and `!group.Managed` and then deletes. It also does not mirror `Add`'s invite handling: `Add` appends the group to a pending `org_invites.usergroup_ids` row when the email has no user yet (`usergroups.go:614-638`), while `Remove` calls `FindUserByEmail` and returns `NotFound` for an invitee.
- **Failure scenario:** (a) An org **editor** (has `ManageOrgMembers`, lacks `ManageOrgAdmins`) removes members from a group that carries the org `admin` role — the exact operation the `Add` guard was written to prevent. (b) An admin adds `newhire@corp.com` to group `analysts` while the invite is pending, then changes their mind; the removal returns `NotFound`, and the user still lands in `analysts` when they sign up.
- **Recommendation:** Copy the `FindOrganizationMemberUsergroupRole` + `ManageOrgAdmins` guard from `Add` into `Remove`, and add the mirror-image invite branch that strips the group id from `org_invites.usergroup_ids` via `UpdateOrganizationInviteUsergroups`.

**Verification:** CONFIRMED — guard present at `usergroups.go:604-610`, absent from `usergroups.go:740-759`; invite branch present at `usergroups.go:614-638`, absent at `usergroups.go:754-757`.

---

### UM-07 [MEDIUM] Seat quota can be bypassed by adding users to a project instead of the org

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/projects.go:1135-1310` (no seat check); `admin/server/organizations.go:503-510` (the check that exists); `admin/user.go:44-70`; `admin/database/postgres/postgres.go:2014-2038`
- **Problem:** `AddOrganizationMemberUser` enforces `QuotaSeats` against `CountOrganizationMemberUsers`, which counts every row in `users_orgs_roles` regardless of role — including guests. `AddProjectMemberUser` performs no seat check at all, yet `InsertProjectMemberUser` (`admin/user.go:62-67`) adds any non-member to the org as a `guest`, creating exactly the row the quota counts.
- **Failure scenario:** An org at its seat cap cannot add another org member, but can add unlimited users to any project; each one becomes a billable org guest. The `billing_reporter` job (`admin/jobs/river/billing_reporter.go:39`) uses the same count, so the reported seat usage exceeds the enforced quota indefinitely.
- **Recommendation:** Apply the same `QuotaSeats` check in `AddProjectMemberUser` before `InsertProjectMemberUser`, or move the check inside `admin.InsertOrganizationMemberUser` so every path that creates a membership row is covered.

**Verification:** CONFIRMED — grep for `QuotaSeats` in `admin/server` yields only `organizations.go:508`; the guest insert is at `admin/user.go:64`; the count has no role filter when `filterRoleID == ""` (`postgres.go:2019-2022`).

---

### UM-08 [MEDIUM] The internal-domain seat exclusion matches `display_name`, so any user can exclude themselves from the billable count

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/database/postgres/postgres.go:2024-2031`; callers `admin/server/organizations.go:504`, `admin/jobs/river/billing_reporter.go:39`; constant at `admin/billing/biller.go:16-17`
- **Problem:** Both callers pass `searchPattern = "%@" + billing.InternalEmailDomain` with `negateSearch = true` to exclude Rill staff from billable seats. The generated predicate is:

  ```sql
  AND NOT (lower(u.email) ILIKE $n OR lower(u.display_name) ILIKE $n)
  ```

  `display_name` is populated from the identity provider on every login (`admin/user.go:141-151` writes `name` straight through), so it is fully user-controlled.
- **Failure scenario:** A user sets their IdP display name to `me@rilldata.com`. On next login their name is synced, and they are excluded from both the enforced seat quota and the seat count reported to the biller.
- **Recommendation:** Add a dedicated parameter for the internal-domain exclusion that matches `u.email` only, and keep `display_name` matching for the interactive `searchPattern` case.

**Verification:** CONFIRMED — the `OR lower(u.display_name) ILIKE` clause shares the same negation and placeholder (`postgres.go:2029`); the display name write-through is at `admin/user.go:142-143`.

---

### UM-09 [MEDIUM] The outstanding-invite quota is enforced even when no invite will be created

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/organizations.go:427-433`; `admin/server/projects.go:1159-1170`
- **Problem:** Both handlers count outstanding invites and reject the request *before* looking up whether the email belongs to an existing Rill user. When the user exists, the handler takes the membership branch and never creates an invite, so the quota is irrelevant to the operation being performed.
- **Failure scenario:** An org sits at its invite cap because of stale pending invites. An admin tries to add an existing colleague — a pure membership write — and gets `quota exceeded: org can at most have N outstanding invitations`, with no indication that clearing unrelated pending invites would unblock it.
- **Recommendation:** Move the quota check into the `errors.Is(err, database.ErrNotFound)` branch, immediately before `InsertOrganizationInvite` / `InsertProjectInvite`.

**Verification:** CONFIRMED — the check precedes `FindUserByEmail` at `organizations.go:427` vs `:455`, and `projects.go:1160` vs `:1196`.

---

### UM-10 [MEDIUM] Email delivery failure is reported as if the membership write failed

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/organizations.go:485-496, 521-531`; `admin/server/projects.go:1256-1268, 1294-1305`
- **Problem:** In all four paths, the DB write commits and then the notification email is sent; a send error is returned directly to the caller. The client cannot distinguish "the user was not added" from "the user was added but we could not email them."
- **Failure scenario:** The email provider is briefly unavailable. `AddOrganizationMemberUser` returns an error; the UI shows a failure (and `AddUsersDialog` labels it "already a member" — see UM-37). The admin retries; the second attempt hits `ErrNotUnique` and returns `AlreadyExists`, which reads as a contradiction of the first error.
- **Recommendation:** Log the send failure and return success, or enqueue the email as a job. If a hard failure is genuinely wanted, use a distinct code and message ("added, but the notification email could not be sent").

**Verification:** CONFIRMED — the writes at `organizations.go:462-483`/`:513` and `projects.go:1229-1254`/`:1276` precede the sends, and each `SendOrganizationInvite`/`SendOrganizationAddition`/`SendProjectInvite`/`SendProjectAddition` error is returned unwrapped.

---

### UM-11 [MEDIUM] Removing a project member leaves an orphaned org `guest` membership

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/projects.go:1367`; contrast `admin/user.go:75-100`
- **Problem:** `AddProjectMemberUser` creates an org `guest` row for users who are not yet org members (`admin/user.go:62-67`). `RemoveProjectMemberUser` calls only `DB.DeleteProjectMemberUser` — there is no reverse cleanup. `DeleteOrganizationMemberUser` cascades org → projects, but nothing cascades project → org guest.
- **Failure scenario:** A contractor is given access to one project, then removed. They remain an org guest forever: they still appear on the Guests page, still count toward `QuotaSeats` and the reported billing seats (see UM-07), and still resolve `ReadOrg` through the guest role.
- **Recommendation:** After deleting the project membership, check whether the user holds any remaining project role or non-guest org role in that org, and drop the guest org membership when none remain — transactionally, mirroring `InsertProjectMemberUser`.

**Verification:** CONFIRMED — `projects.go:1367` is the only delete; the guest insert is at `admin/user.go:64`; `guest` is a real row in `users_orgs_roles` per `migrations/0061.sql:19-30`.

---

### UM-12 [LOW] Re-affirming the last admin's role fails with "cannot remove the last admin member"

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/organizations.go:652-661`
- **Problem:** The `isLastAdmin` check runs unconditionally, without comparing the requested role to the current one. Setting the last admin's role to `admin` — a no-op — is rejected.
- **Failure scenario:** A script or UI that reconciles roles idempotently ("ensure alice is admin") errors on the one org where alice is the sole admin.
- **Recommendation:** Skip the `isLastAdmin` rejection when `role.Admin` is true; the invariant being protected is "at least one admin remains."

**Verification:** CONFIRMED — `organizations.go:659-661` has no `role.Admin` condition, unlike the demotion case it is meant to catch.

---

### UM-13 [LOW] Invite acceptance and domain-whitelist auto-join ignore both quotas

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/user.go:194-301`
- **Problem:** `CreateOrUpdateUser` processes org invites, whitelisted org domains, project invites, and whitelisted project domains, calling `InsertOrganizationMemberUser` / `InsertProjectMemberUser` for each with no seat-quota check. The whitelist path in particular has no upper bound: every future signup from a whitelisted domain joins automatically.
- **Failure scenario:** An org whitelists `corp.com` and sets `QuotaSeats`. Every new `@corp.com` signup joins the org regardless of the cap; the enforced quota only ever blocks explicit adds.
- **Recommendation:** Enforce `QuotaSeats` in the whitelist loops at minimum, and decide explicitly whether an accepted invite may exceed the cap (arguably yes — the seat was reserved when the invite was created — but that should be a stated decision, and the invite quota is not decremented on acceptance either).

**Verification:** CONFIRMED — no quota reads anywhere in `admin/user.go`; the four insertion loops are at `:197-234`, `:242-257`, `:262-279`, `:286-301`.

---

### UM-14 [LOW] `ListRoles` performs no authentication check

- **Category:** security  **Confidence:** high
- **Location:** `admin/server/roles.go:11-36`
- **Problem:** The handler never calls `auth.GetClaims`. Every other handler in the package begins with a claims check; this one returns the full org and project role catalogue with expanded permission bits to anonymous callers.
- **Failure scenario:** Low impact — the role names and their permissions are public documentation. It is nonetheless an unauthenticated endpoint that leaks the exact permission matrix, including any future private or experimental role.
- **Recommendation:** Require an authenticated owner (`claims.OwnerType() != auth.OwnerTypeAnon`), consistent with the rest of the package.

**Verification:** CONFIRMED — no `auth.GetClaims` reference in `admin/server/roles.go`.

---

### UM-15 [LOW] Attribute validation errors surface as HTTP 500

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/database/postgres/postgres.go:4096-4117`; `admin/server/server.go:440-466`
- **Problem:** `validateAttributes` returns plain `fmt.Errorf` values ("too many attributes…", "invalid attribute key…", "attribute value … too long"). `mapGRPCError` recognises `database.ErrValidation` and maps it to `InvalidArgument`, but a bare error falls through to the final `status.Error(codes.Internal, ...)`.
- **Failure scenario:** A user passes an attribute key containing a hyphen and gets a 500 with an internal-looking error, and the failure is recorded as a server error in monitoring.
- **Recommendation:** Wrap these in `database.NewValidationError`, as the rest of the package does.

**Verification:** CONFIRMED — the three returns at `postgres.go:4102, 4108, 4113` use `fmt.Errorf`; `mapGRPCError`'s `ErrValidation` branch is at `server.go:459-461` and the `Internal` fallthrough at `:466`.

---

### UM-16 [LOW] Superuser force-access cannot revoke another user's token by prefix

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/users.go:723-786`, called from `:322`
- **Problem:** `findUserAuthTokenFuzzy` resolves exact IDs and full token strings globally, but the prefix-matching fallback lists tokens with `FindUserAuthTokens(ctx, userID, ...)` where `userID = claims.OwnerID()` — always the caller. `RevokeUserAuthToken` then applies the `SuperuserForceAccess` check to whatever was found, so the force-access path works for full IDs and dies with `NotFound` for prefixes.
- **Failure scenario:** A superuser responding to a leaked token has only its prefix (which is all `ListUserAuthTokens` displays) and cannot revoke it.
- **Recommendation:** Thread the target user id into `findUserAuthTokenFuzzy` when `SuperuserForceAccess` is set, or document that prefix revocation is self-only.

**Verification:** CONFIRMED — `userID := claims.OwnerID()` at `users.go:726`, used at `:755`; the force-access check happens afterwards at `:327-331`.

---

### UM-17 [LOW] Full-row read-modify-write on `users` lets concurrent updates clobber each other

- **Category:** concurrency  **Confidence:** medium
- **Location:** `admin/server/users.go:149-168` and `:562-578`; `admin/user.go:139-151`; `admin/database/postgres/postgres.go:932-940`
- **Problem:** `UpdateUser` takes a complete `UpdateUserOptions` and writes every column. Three callers therefore read the row, copy the fields they are not changing, and write it all back: preferences, sudo quotas, and login sync. Nothing serialises them.
- **Failure scenario:** A GitHub OAuth token refresh writes `github_token`/`github_token_expires_on` while the user is saving a timezone preference. The preference update was built from a pre-refresh snapshot and writes the stale GitHub token back, invalidating the connection.
- **Recommendation:** Make `UpdateUserOptions` fields optional pointers and emit a partial `UPDATE`, or take a row lock for the read-modify-write.

**Verification:** PLAUSIBLE — the pattern is unambiguous (`users.go:155-165` copies seven unrelated fields verbatim) and the SQL writes all ten columns (`postgres.go:938`); the interleaving is timing-dependent, hence medium confidence.

---

### UM-18 [LOW] `CountOrganizationInvites` is a dead duplicate of `CountInvitesForOrganization`

- **Category:** tech-debt  **Confidence:** high
- **Location:** `admin/database/postgres/postgres.go:2620-2626` and `:2682-2691`; interface entries in `admin/database/database.go`
- **Problem:** Both run `SELECT COUNT(*) FROM org_invites WHERE org_id = $1`. Only `CountInvitesForOrganization` has callers (`organizations.go:387,427`, `projects.go:1160`).
- **Recommendation:** Delete `CountOrganizationInvites` from the interface and the implementation.

**Verification:** CONFIRMED — identical SQL; grep for `CountOrganizationInvites` outside `postgres.go`/`database.go` returns nothing.

---

### UM-19 [LOW] Permission unions are hand-maintained in six places

- **Category:** tech-debt  **Confidence:** high
- **Location:** `admin/permissions.go:102-146, 150-197, 199-236, 238-270, 272-286, 288-314`
- **Problem:** Every permission field is listed by hand in `UnionOrgRoles`, `UnionProjectRoles`, the two "org admin gets everything" literals, `ProjectPermissionsForDeployment`, and `ProjectPermissionsForMagicAuthToken`. Adding a field to the proto requires six correct edits, and omitting one fails silently as a permission that is never granted (or, in the union, never inherited from a role).
- **Recommendation:** Generate the union with reflection over the proto message, or derive the "full admin" literals from a single `allProjectPermissions()` helper so new fields default consistently.

**Verification:** CONFIRMED — six exhaustive field lists over the same message types.

---

### UM-20 [LOW] Malformed member attributes are silently dropped from the API response

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/users.go:656-676`
- **Problem:** `orgMemberUserToPB` does `if s, err := structpb.NewStruct(m.Attributes); err == nil { attributes = s }` — a conversion failure yields a nil `Attributes` field indistinguishable from "no attributes."
- **Failure scenario:** A stored attribute value that `structpb` cannot represent makes the member look attribute-free in the UI and CLI, so `rill user set-attributes` (which warns before overwriting) skips its warning and silently replaces them.
- **Recommendation:** Log or propagate the error rather than discarding it.

**Verification:** CONFIRMED — the swallowed error is at `users.go:659-661`; the CLI's overwrite warning keys on `Attributes` being non-empty (`cli/cmd/user/set_attributes.go:47`).

---

### UM-21 [LOW] `UpdateUsergroup` applies name and description as two independent writes

- **Category:** correctness  **Confidence:** high
- **Location:** `admin/server/usergroups.go:137-149`
- **Problem:** `UpdateUsergroupName` and `UpdateUsergroupDescription` are separate statements with no transaction. If the second fails (for example the description exceeds a column limit), the rename has already committed.
- **Failure scenario:** A caller renaming and re-describing a group in one request sees an error but finds the group renamed — and, if they retry with the original name in the request, gets `NotFound` for the old name.
- **Recommendation:** Wrap both in `NewTx`, or replace them with a single partial `UPDATE`.

**Verification:** CONFIRMED — no `NewTx` in the handler; two sequential mutations at `usergroups.go:138` and `:145`.

---

### UM-22 [LOW] Wrong error message when assigning an admin project role to a group

- **Category:** readability  **Confidence:** high
- **Location:** `admin/server/usergroups.go:511-513`
- **Problem:** The assign path returns "as a non-admin you are not allowed to **remove** an admin role"; the correct wording ("assign") is used in the equivalent check in `AddProjectMemberUsergroup` (`usergroups.go:434`).
- **Recommendation:** Change to "assign".

**Verification:** CONFIRMED.

---

## Frontend findings

### UM-50 [HIGH] No mutation in the org user-management UI invalidates the queries the pages actually render

- **Category:** bug  **Confidence:** high
- **Location:** every invalidation site in `web-admin/src/features/organizations/user-management/**` and `web-admin/src/features/projects/user-management/**` (43 calls); key generators at `web-admin/src/client/gen/default/default.ts:3358-3374`, `:3136-3146`, `:10958-10968`; query-client defaults at `web-common/src/lib/svelte-query/globalQueryClient.ts:7-22`
- **Problem:** Orval generates two query-key shapes per list endpoint, and the infinite one is **prefixed**, not suffixed:

  ```ts
  getAdminServiceListOrganizationMemberUsersInfiniteQueryKey = (org, params) =>
    ["infinite", `/v1/orgs/${org}/members`, ...(params ? [params] : [])];

  getAdminServiceListOrganizationMemberUsersQueryKey = (org, params) =>
    [`/v1/orgs/${org}/members`, ...(params ? [params] : [])];
  ```

  The user, guest, and group pages all render **infinite** queries (`selectors.ts:87, 108, 128, 174`, `groups/+page.svelte:20`, `CreateUserGroupDialog.svelte:57`, `ShareProjectForm.svelte:52`). Every mutation invalidates with the **plain** key — `invalidateAfterUserDelete` (`utils.ts:24-51`), `UserRoleCell.handleSetRole/handleUpgrade` (`:71-78, 102-109`), `UserSetRole` (`:58-70, 91-103`), `AddUsersDialog.handleCreate` (`:62-69`), `UserAndGroupInviteForm.processInvitations` (`:73-97`), and the group dialogs. A `grep` across both feature directories returns 43 invalidation call sites and **zero** that use an infinite key.

  TanStack Query v5 matches `queryKey` by *positional* prefix (`partialMatchKey` compares element 0 against element 0). `["/v1/orgs/acme/members"]` therefore does not match `["infinite", "/v1/orgs/acme/members", {…}]` — index 0 differs immediately — so the invalidation resolves to zero queries and silently succeeds.

  Nothing else covers the gap: the global query client sets `refetchOnMount: false`, `refetchOnReconnect: false`, and `refetchOnWindowFocus: false` (`globalQueryClient.ts:10-12`), so there is no background refetch. Only a full page load re-runs the query.
- **Failure scenario:** This is the reported symptom. An admin removes a user from `/<org>/-/users`. The `RemoveOrganizationMemberUser` call succeeds, the "user removed" toast fires, `invalidateAfterUserDelete` runs and matches nothing, and the row stays in the table until the page is reloaded. The same holds for every other mutation on these pages: role changes, guest upgrades, adding users, creating and editing groups, and the member/group counts in the left nav. Re-removing the stale row produces a confusing `NotFound`.

  The project share dialog is partly insulated by accident — its queries set `refetchOnMount: true` and `refetchOnWindowFocus: true` explicitly (`ShareProjectForm.svelte:91-92, 103-104, 148-149, 160-161, 186-187, 196-197`), so closing and reopening the popover remounts and refetches. That is also why the bug reads as "org users page only" in practice.
- **Recommendation:** Add a shared helper that invalidates both shapes, and route all 43 sites through it. Because the infinite key is exactly `["infinite", ...plainKey]`, one helper covers every endpoint without touching the generated code:

  ```ts
  export function invalidateListQuery(queryClient: QueryClient, key: readonly unknown[]) {
    return Promise.all([
      queryClient.invalidateQueries({ queryKey: key }),
      queryClient.invalidateQueries({ queryKey: ["infinite", ...key] }),
    ]);
  }
  ```

  This is a repo-wide hazard, not a user-management one — the same prefix mismatch will silently break any other feature that pairs an infinite list with a plain-key invalidation. Worth a follow-up sweep and a lint rule or a wrapper that makes the plain-key-only call hard to write.

**Verification:** CONFIRMED — the two key shapes are at `default.ts:3358-3367` (infinite, `"infinite"` at index 0) and `:3369-3374` (plain); the same pattern holds for invites (`:3136-3146`) and usergroups (`:10958-10968`). All list UIs on the affected pages use the infinite constructors, and a grep for `getAdminService…QueryKey` in the two feature directories returns 43 hits, none of them `…InfiniteQueryKey`. TanStack `~5.76.0` per `web-admin/package.json:26`; `partialMatchKey` compares array elements positionally from index 0. All three `refetchOn*` defaults are `false` (`globalQueryClient.ts:10-12`), so no background refetch masks the miss.

**Manual reproduction (QA):**
1. Open `/<org>/-/users` as an org admin.
2. Remove a user via the role dropdown → Remove.
3. Observe the success toast and that the row remains in the table.
4. Reload the page; the row is gone, confirming the server-side removal succeeded.
5. Repeat for a role change and for creating a group on the Groups tab — same behaviour.

---

### UM-30 [HIGH] Search and filtering run client-side over paginated data on all three org user pages

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/routes/[organization]/-/users/+page.svelte:51-122`; `.../users/guests/+page.svelte:52-107`; `.../users/groups/+page.svelte:20-31, 96-102`; `web-admin/src/features/organizations/user-management/selectors.ts:80-105, 127-144, 173-191`
- **Problem:** Each page creates an infinite query with `pageSize: 50`, flattens whatever pages are loaded, and then filters in JavaScript by search text, membership type, and role. The backend already supports the right filters — `ListOrganizationMemberUsers` accepts `searchPattern` and `role` (`organizations.go:324-333`), and `selectors.ts:91` uses `role` for the guest split — but the list page passes neither and re-implements both in `filteredUsers`.

  Three consequences follow. Search only sees loaded pages. The rendered result count and `groups_total_count` (`groups/+page.svelte:99`) describe the filtered subset of loaded rows, not the total, while `ListOrganizationMemberUsers` already returns an authoritative `totalCount`. And loading more pages depends on scrolling the table (`OrgUsersTable.svelte:137-144` fires `handleLoadMore` from the scroll container), so when a filter matches nothing on page one there is nothing to scroll and the remaining pages are never fetched.
- **Failure scenario:** An org has 300 members. An admin types a colleague's name into the search box. Because only the first 50 members (ordered by email) are loaded, the table shows "no users" for someone who plainly exists — and no amount of waiting or scrolling fixes it, because the empty table has no scrollable content.
- **Recommendation:** Push `searchPattern` (debounced) and `role` into the query parameters, key the infinite query on them, and use the server's `totalCount` for the count labels. Keep client-side work to the members/pending split, which is derivable from the loaded rows.

**Verification:** CONFIRMED — `getOrgUserMembers` (`selectors.ts:80-105`) passes only `pageSize`, `role` (guest split), and `includeCounts`; `filteredUsers` at `+page.svelte:81-116` does the search and role filtering; the load-more trigger is scroll-driven (`OrgUsersTable.svelte:137-144, 157-168`).

**Manual reproduction (QA):**
1. Use an org with more than 50 members.
2. Open `/<org>/-/users` without scrolling.
3. Search for a member whose email sorts after the first 50 alphabetically.
4. Observe an empty table; scrolling does not help because there are no rows to scroll.

---

### UM-31 [HIGH] Renaming a group and editing its members in one save loses the membership edits

- **Category:** bug  **Confidence:** high
- **Location:** `web-admin/src/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte:107-135, 151-201, 228-241`
- **Problem:** The submit handler is:

  ```js
  await handleRename(groupName, values.newName);
  await applyPendingChanges();
  open = false;
  ```

  `applyPendingChanges` addresses every add/remove as `usergroup: groupName` — the **pre-rename** name, since `groupName` is the component's prop and is not updated by the rename. After a successful rename that name no longer resolves, so every membership mutation returns `NotFound`.

  Two aggravating factors: `handleRename` catches its own error and does not rethrow, so `applyPendingChanges` runs even when the rename failed; and `open = false` executes unconditionally, so the dialog closes as though the save succeeded regardless of what happened. `pendingAdditions`/`pendingRemovals` are only cleared on the success path (`:187-188`), so a retry re-sends already-applied additions.
- **Failure scenario:** An admin opens `analysts`, renames it to `data-analysts`, adds two people, and saves. The group is renamed, an error toast flashes, the dialog closes, and neither person is in the group.
- **Recommendation:** Apply member changes first, then rename; or capture the effective name (`values.newName` after a successful rename) and pass it to `applyPendingChanges`. Rethrow from `handleRename`, and only close the dialog when both steps succeed.

**Verification:** CONFIRMED — `usergroup: groupName` at `:156` and `:165`; `handleRename`'s catch at `:127-134` has no rethrow; the unconditional close at `:235`.

**Manual reproduction (QA):**
1. Open `/<org>/-/users/groups` and edit a group with at least one member.
2. Change the name **and** add a user in the same dialog.
3. Save.
4. Observe the rename applied, an error toast, and the added user missing from the group.

---

### UM-32 [MEDIUM] The edit-group dialog opens with an empty name on the groups page

- **Category:** bug  **Confidence:** high
- **Location:** `EditUserGroupDialog.svelte:205-207, 222-241, 266-277`; `web-admin/src/routes/[organization]/-/users/groups/+page.svelte:113-117`
- **Problem:** `const initialValues = { newName: groupName }` is evaluated once when the component is constructed, and `superForm(defaults(initialValues, schema))` seeds the form from it. The groups page mounts `EditUserGroupDialog` unconditionally with `groupName={userGroupName}`, which starts as `""`, so the form's `newName` is `""` and never re-syncs when a group is selected. `handleClose` then "restores" `$form.newName = initialValues.newName` — still `""`.

  The users page avoids this by mounting the dialog behind `{#if editingUserGroupName}` (`users/+page.svelte:208-214`), so its `initialValues` capture is correct.
- **Failure scenario:** On the Groups tab, clicking Edit on `analysts` shows an empty Name field with "Name is required", and the save button is blocked until the admin retypes the name from memory. After editing a second group, the field shows the first group's name.
- **Recommendation:** Sync the form when the dialog opens (`$: if (open) $form.newName = groupName`), or mount the dialog behind an `{#if}` on the selected group name as the users page does.

**Verification:** CONFIRMED — the one-time capture at `:205-207`; the unconditional mount with an initially empty name at `groups/+page.svelte:113-117`; the conditional mount that works at `users/+page.svelte:208`.

---

### UM-33 [MEDIUM] The Guests page lists all pending org invites, not just guest invites

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/routes/[organization]/-/users/guests/+page.svelte:52-78`; `admin/server/organizations.go:360-407`
- **Problem:** The member query is correctly scoped (`getOrgUserMembers({ guestOnly: true })` → `role: "guest"`), but the invite query is `getOrgUserInvites(organization)`, which has no role filter — `ListOrganizationInvitesRequest` does not offer one. Both are merged into `combinedRows` and rendered.
- **Failure scenario:** An admin invites a new editor. That invite appears on both the Members tab and the Guests tab, and the Guests tab's count and the Members tab's count now double-count them.
- **Recommendation:** Filter the merged invites by `roleName === OrgUserRoles.Guest` on the guests page (and by `!== guest` on the members page, which currently drops guest *members* at `+page.svelte:83` but not guest *invites*). Longer term, add a role filter to `ListOrganizationInvites`.

**Verification:** CONFIRMED — `guestOnly: true` at `guests/+page.svelte:54` vs the unfiltered `getOrgUserInvites` at `:56`; `ListOrganizationInvites` takes only org/page params (`organizations.go:360-382`).

---

### UM-34 [MEDIUM] Invites with no inviter are rendered as accepted members

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/routes/[organization]/-/users/+page.svelte:99-105`; `.../guests/+page.svelte:97-103`; `.../table/users/OrgUsersTable.svelte:60, 125`; `admin/server/users.go:704-710`; `admin/database/postgres/postgres.go:2604-2618`
- **Problem:** The frontend distinguishes a pending invite from a member with `"invitedBy" in user`. On the backend, `orgInviteToPB` sets `InvitedBy: safeStr(i.InvitedBy)`, and `FindOrganizationInvites` produces that value from `LEFT JOIN users u ON uoi.invited_by_user_id = u.id` — so it is `NULL` → `""` whenever the invite was created without an inviter (a service token, where `invitedByUserID` stays empty per `organizations.go:443-453`) or the inviting user has since been deleted. The admin API is served through the vanguard transcoder with default protojson semantics, which omit empty strings, so the key is absent from the JSON and `"invitedBy" in user` is false.
- **Failure scenario:** An invite sent via a service token, or one whose inviter has left the company, shows up in the members list with no "pending" chip, is excluded from the "Pending invites" filter, and is presented as if the user has already joined.
- **Recommendation:** Detect invites structurally rather than by an optional field — tag the rows in `coerceInvitesToUsers` (e.g. `isInvite: true`) and branch on that everywhere. Independently, consider making `InvitedBy` an explicitly optional proto field so absence is meaningful.

**Verification:** CONFIRMED — `LEFT JOIN` at `postgres.go:2610`; `safeStr` at `users.go:708`; the four `"invitedBy" in user` checks at `users/+page.svelte:99,105`, `guests/+page.svelte:97,103`, plus `OrgUsersTable.svelte:60,125`.

---

### UM-35 [MEDIUM] The project share dialog offers role and remove controls to users who cannot use them

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/features/projects/user-management/ShareProjectForm.svelte:343-362`; `UserSetRole.svelte:117-196`
- **Problem:** `ShareProjectForm` passes `manageProjectMembers={true}` as a literal for every member and invite row, while correctly threading the real `manageProjectAdmins` prop. `UserSetRole` gates its entire dropdown on `manageProjectMembers && !isCurrentUser`, so every viewer sees editable role dropdowns and a red "Remove" item for every other member.

  A second layer of the same problem: when the viewer *does* have `manageProjectMembers` but not `manageProjectAdmins` and the target is an admin, `UserSetRole.svelte:125` only hides the caret icon — the dropdown still opens and still offers Editor, Viewer, and Remove for that admin.
- **Failure scenario:** A project viewer opens Share, sees full role controls, demotes someone, and gets `PermissionDenied` rendered as a generic "error updating role" toast.
- **Recommendation:** Pass the real `manageProjectMembers` permission down from the route (it is already available alongside `manageProjectAdmins`), and disable the dropdown items — not just the caret — when the target is an admin and the caller lacks `manageProjectAdmins`.

**Verification:** CONFIRMED — the hardcoded literal at `ShareProjectForm.svelte:350` and `:360`; the gate at `UserSetRole.svelte:117`; the caret-only guard at `:125`.

---

### UM-36 [MEDIUM] Twelve catch blocks dereference `error.response.data.message` unguarded

- **Category:** bug  **Confidence:** high
- **Location:** `EditUserGroupDialog.svelte:130,196`; `CreateUserGroupDialog.svelte:136,173`; `ProjectUserGroupSetRole.svelte:65,96,124`; `OrgUserGroupSetRole.svelte:60,88,113`; `UsergroupSetRole.svelte:53,81`
- **Problem:** Each is `catch (error) { eventBus.emit("notification", { message: m.common_error_message({ message: error.response.data.message }), ... }) }`. A network failure, a timeout, or any non-Axios throw has no `response`, so the catch block itself throws a `TypeError`, the notification is never emitted, and the rejection escapes as an unhandled promise rejection.
- **Failure scenario:** The user goes offline mid-save. Instead of "network error", they get complete silence — the dialog appears to do nothing.
- **Recommendation:** Route all of these through the existing helper. `web-admin/src/features/projects/user-management/errors.ts` already exports a `parseError` that does `error.response?.data?.message ?? error.message`; it is currently unused (see UM-48). Generalise it and adopt it in all twelve places.

**Verification:** CONFIRMED — grep for `error.response.data` in the two user-management feature directories returns exactly these twelve sites; `errors.ts:5` contains the correct optional-chaining form.

---

### UM-37 [MEDIUM] Every failed add in the Add Users dialog is reported as "already a member"

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/features/organizations/user-management/dialogs/AddUsersDialog.svelte:112-153, 266-273`
- **Problem:** The per-email catch discards the error and records only the address. The template then renders `m.users_already_member({ emails, count })` for whatever landed in `failedInvites`, with no inspection of the actual status.
- **Failure scenario:** The org hits `QuotaOutstandingInvites` or `QuotaSeats` (`FailedPrecondition`), the caller lacks `ManageOrgAdmins` for the chosen role (`PermissionDenied`), or the notification email fails (UM-10) — and in all four cases the admin is told the user is already a member, which is both wrong and actively misleading about the remedy.
- **Recommendation:** Capture the error per email and group the message by status code, falling back to the server's message. `AlreadyExists` is the only case that warrants the current wording.

**Verification:** CONFIRMED — `catch (error) { console.error(...); return { index, email, success: false } }` at `:117-120`; the unconditional message at `:268`.

---

### UM-38 [MEDIUM] The project invite box rejects group names the backend accepts

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/features/projects/user-management/UserAndGroupInviteForm.svelte:151-158`; `admin/database/validate.go:57`; `web-admin/src/features/organizations/user-management/constants.ts:13`
- **Problem:** The validator treats a non-email entry as a group name and tests it against `/^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$/`, which forbids underscores and leading digits-plus-underscore forms. The backend slug rule is `^[_a-zA-Z0-9][-_a-zA-Z0-9]*$` — underscores allowed anywhere — and the org-side UI already mirrors it correctly as `SLUG_REGEX` in `constants.ts`, which `CreateUserGroupDialog` and `EditUserGroupDialog` use to *create* such groups.
- **Failure scenario:** An admin creates the group `data_analysts` on the Groups tab (accepted), then cannot type it into the project share box — it is rejected as an invalid email or group name.
- **Recommendation:** Import `SLUG_REGEX` here instead of the local regex.

**Verification:** CONFIRMED — local regex at `UserAndGroupInviteForm.svelte:155`; `SLUG_REGEX` at `constants.ts:13`, used at `CreateUserGroupDialog.svelte:30` and `EditUserGroupDialog.svelte:216`.

---

### UM-39 [MEDIUM] The share dialog runs duplicate paginated and non-paginated queries for the same data

- **Category:** performance  **Confidence:** high
- **Location:** `ShareProjectForm.svelte:96-121, 123-150, 152-163, 165-188, 223-238`
- **Problem:** Members are fetched twice — `createAdminServiceListProjectMemberUsers` **and** `projectMembersInfiniteQuery` — and invites likewise. The non-infinite results are only ever used as a fallback in `infiniteMembers ?? $listProjectMemberUsers?.data?.members ?? []`, which is reachable only while the infinite query has no pages at all. The non-infinite member query also carries a `select` that sorts the current user to the top — work on data that is essentially never rendered. All four have `refetchOnMount` and `refetchOnWindowFocus` enabled, so opening the popover issues four requests and every window focus issues four more.
- **Recommendation:** Delete the two non-infinite queries and the redundant `select`; keep the infinite queries and the sort that already exists at `:233-237`.

**Verification:** CONFIRMED — the four query declarations and the `??` fallback chain at `:228-230`.

---

### UM-40 [MEDIUM] Org members never appear in the share dialog's local suggestion list

- **Category:** correctness  **Confidence:** high
- **Location:** `ShareProjectForm.svelte:202, 266-273`; `web-admin/src/features/projects/user-management/helpers.ts:20-45`
- **Problem:** `$: allOrgMemberUsersRows = [];` — a reactive statement that assigns a literal empty array — is passed as the first argument to `buildSearchList`, whose entire first loop (the one that emits org member users, with name, photo, and org role) therefore never produces anything. Only org invites and org groups reach the local list. The remote `handleSearch` partially compensates but returns at most five results and only for the typed prefix.
- **Failure scenario:** Opening Share and clicking the input with no query shows groups and pending invitees but no existing org members, so the fastest path to "add a teammate" is missing.
- **Recommendation:** Either wire a real org-members query into `allOrgMemberUsersRows` or delete the dead parameter and the corresponding loop in `buildSearchList` so the remote search is unambiguously the single source.

**Verification:** CONFIRMED — the literal at `:202`; the dead loop at `helpers.ts:29-43`.

---

### UM-41 [MEDIUM] Remote user search is prefix-only and does not exclude existing project members

- **Category:** correctness  **Confidence:** high
- **Location:** `UserAndGroupInviteForm.svelte:160-195`; `admin/database/postgres/postgres.go:1992-1996`
- **Problem:** `handleSearch` sends `searchPattern: \`${query}%\``. The backend applies it as `lower(u.email) ILIKE $n OR lower(u.display_name) ILIKE $n`, so only prefixes match — typing `smith` finds nothing for `john.smith@acme.com`. The same function excludes only the current user; unlike the local `buildSearchList` path it does not filter against `projectMemberEmailSet` or `projectInviteEmailSet`.
- **Failure scenario:** An admin searching by surname sees no results and concludes the person has no account; or they pick a suggestion for someone already on the project and re-send an invite.
- **Recommendation:** Use `%${query}%`, and filter the remote results against the project member and invite sets before returning them.

**Verification:** CONFIRMED — the prefix pattern at `:167`; the ILIKE predicate at `postgres.go:1993`; the self-only filter at `:181-183`.

---

### UM-42 [MEDIUM] Group counts and per-user group counts only reflect the first 50 groups

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/features/organizations/user-management/selectors.ts:21-76, 146-171`
- **Problem:** `getUserGroupsForUsersInOrg` fetches org usergroups with a single non-paginated `pageSize: 50` query and looks up each of the user's groups in that page to read `usersCount`; groups beyond the first 50 resolve to `count: 0`. `getUserCounts` computes `groupsCount` as `orgUsergroupsResp.data?.members?.filter(g => !g.groupManaged).length` over the same 50-row page, so the Groups tab label caps at 50 minus the managed groups. It also derives `membersCount` as `allUsersCounts - guestUsersCounts` from two independently-fetched `totalCount` values.
- **Failure scenario:** An org with 80 groups shows "Groups (47)" permanently, and members of the 51st-onwards groups show a 0 user count in the Groups cell.
- **Recommendation:** Use the server `totalCount` for the tab label, add a managed/unmanaged filter to `ListOrganizationMemberUsergroups` so the count can be computed server-side, and fetch the specific groups a user belongs to rather than searching a page.

**Verification:** CONFIRMED — `PAGE_SIZE = 50` non-infinite query at `:30-35` and `:193-198`; the client-side `.filter(...).length` at `:160-162`.

---

### UM-43 [MEDIUM] Removing an org member does not invalidate project member lists

- **Category:** correctness  **Confidence:** high
- **Location:** `web-admin/src/features/organizations/user-management/utils.ts:20-52`; `admin/user.go:75-100`
- **Problem:** `DeleteOrganizationMemberUser` transactionally removes the user from the org, from every usergroup in the org, and from **every project in the org**. `invalidateAfterUserDelete` invalidates the org member list, the org invite list, and the three hardcoded autogroup member lists — but not `ListProjectMemberUsers`, `ListProjectMemberUsergroups`, or `ListOrganizationMemberUsergroups` (group user counts).
- **Failure scenario:** An admin removes a user on the Users tab, then opens a project's Share dialog and still sees them listed until the cache expires or the window is refocused.
- **Recommendation:** Add invalidations for the project member/invite lists and the usergroup list. Replace the hardcoded `"autogroup:users"` / `"autogroup:members"` / `"autogroup:guests"` literals with shared constants — they duplicate `admin/database/database.go:760-762` in five frontend locations (also `ShareProjectForm.svelte:191, 213, 278, 404`).

**Verification:** CONFIRMED — the cascade at `admin/user.go:88-97`; the five invalidations at `utils.ts:24-51`; the backend constants at `database.go:760-762`.

---

### UM-44 [LOW] Removing the billing contact from the main users table shows a generic error

- **Category:** correctness  **Confidence:** high
- **Location:** `.../table/users/UserRoleCell.svelte:49-54, 123-142, 205-213`; `.../table/users/OrgUsersTable.svelte:114-135`
- **Problem:** `UserRoleCell` guards the *role change* path with `onAttemptChangeBillingContactUserRole` (`:50-53`) but calls `handleRemove` directly for removals. The corresponding `onAttemptRemoveBillingContactUser` guard lives only in `UserActionsCell`, and `OrgUsersTable` includes that column exclusively in the `guestOnly` layout (`:133-135`). On the main Users tab, removing the billing contact therefore hits the backend's `FailedPrecondition` (`organizations.go:584-586`) and renders `users_error_removing`.
- **Failure scenario:** An admin tries to remove the billing contact and is told only "error removing user", with no hint that the billing email must be reassigned first — even though a dedicated `RemovingBillingContactDialog` exists and is already mounted on the page.
- **Recommendation:** Apply the `isBillingContact` guard in `UserRoleCell.handleRemove` and route it to `onAttemptRemoveBillingContactUser`, which the page already supplies.

**Verification:** CONFIRMED — the asymmetric guards at `UserRoleCell.svelte:50` vs `:207-210`; the guest-only column list at `OrgUsersTable.svelte:133-135`; the mounted-but-unreachable dialog at `users/+page.svelte:192-195`.

---

### UM-45 [LOW] The invite input's loading state is a no-op and its search results can arrive out of order

- **Category:** bug  **Confidence:** high
- **Location:** `web-admin/src/features/projects/user-management/SearchAndInviteInput.svelte:59-90, 129-131`; `web-common/src/lib/create-debouncer.ts`
- **Problem:** `debounce` returns `(...args) => void` — it schedules a `setTimeout` and returns `undefined`. So `loading = true; await debouncedSearch(input); loading = false;` awaits `undefined` and flips `loading` back in the same microtask, before the search has even been scheduled. Separately, `debouncedSearch` performs an `await onSearch(query)` and then assigns `searchResults = merged` with no sequence guard, so a slow response for an earlier query can overwrite a faster response for a later one.
- **Recommendation:** Track loading inside the debounced function, and discard responses whose request id is not the latest.

**Verification:** CONFIRMED — `debounce` returns void (`create-debouncer.ts`); the await-and-reset at `SearchAndInviteInput.svelte:129-131`; the unguarded assignment at `:85`.

---

### UM-46 [LOW] Enter on invalid comma-separated input silently discards everything typed

- **Category:** bug  **Confidence:** high
- **Location:** `SearchAndInviteInput.svelte:226-244`
- **Problem:** The Enter handler calls `processCommaSeparatedInput`, applies the parsed entries only `if (!processError)`, and then clears `input = ""` **unconditionally** without ever assigning `processError` to the displayed `error`. The `handleInput` path (`:112-127`) does surface the error correctly; the Enter path does not.
- **Failure scenario:** A user pastes `alice@corp.com, bob@corp, carol@corp.com` and presses Enter. `bob@corp` fails validation, so nothing is added, the field is cleared, and no error is shown — all three addresses are gone with no explanation.
- **Recommendation:** Assign `error = processError` and preserve the input when validation fails, matching `handleInput`.

**Verification:** CONFIRMED — `processError` is destructured at `:228` and never read; the unconditional clear at `:241`.

---

### UM-47 [LOW] The groups page mutates `$page.url` instead of navigating

- **Category:** bug  **Confidence:** high
- **Location:** `web-admin/src/routes/[organization]/-/users/groups/+page.svelte:44-55`
- **Problem:** A reactive block reads `?action=open-edit-user-group-dialog`, opens the dialog, and then calls `$page.url.searchParams.delete(...)` on the store's URL object. That mutates a SvelteKit-owned object without notifying the router, so the address bar keeps the parameters. `ShareProjectPopover.svelte:26-34` does this correctly with `goto(..., { replaceState: true })`.
- **Failure scenario:** After the dialog is dismissed, a page reload (or restoring the tab) reopens it, and the URL remains unshareable.
- **Recommendation:** Use `goto` with `replaceState: true`, as the share popover does.

**Verification:** CONFIRMED — direct mutation at `:53-54`; the correct pattern at `ShareProjectPopover.svelte:28-33`.

---

### UM-48 [LOW] Dead code across the user-management surface

- **Category:** tech-debt  **Confidence:** high
- **Location:** as listed
- **Problem:**
  - `web-admin/src/features/projects/user-management/errors.ts` — `parseError` has no callers, while twelve sites hand-roll a worse version of it (UM-36).
  - `web-admin/src/features/projects/user-management/readPublicEmailDomains.ts` — imports Node's `fs` and reads a path relative to `__dirname`; no callers anywhere in `web-admin/src`.
  - `AddUsersDialog`'s `email`, `role`, and `isSuperUser` props are passed as permanently-empty/false values from both pages, and `superuserForceAccess: isSuperUser` is therefore always `false`; `handleCreate` also assigns to these props (`:71-73`), where the `isSuperUser` write targets a shadowing parameter.
  - `guests/+page.svelte:175-180` mounts an `AddUsersDialog` whose `open` flag is never set true.
  - `filterSelection`'s `"guests"` union member is unreachable — `OrgUsersFilters` never offers it — but both pages branch on it, yielding a silently empty table if it is ever set.
  - `ShareProjectForm.svelte:202` `allOrgMemberUsersRows` (UM-40) and `:46` `accessType`'s initialiser, which is immediately overwritten by the reactive assignment at `:281`.
- **Recommendation:** Delete, or wire up where the intent was real (`parseError` in particular).

**Verification:** CONFIRMED — grep for `parseError`/`readPublicEmailDomains` across `web-admin/src` returns only unrelated `parseErrors` (project status); the prop and dialog sites are as cited.

---

### UM-49 [LOW] Group membership changes are applied one request at a time

- **Category:** performance  **Confidence:** high
- **Location:** `EditUserGroupDialog.svelte:151-168`; `CreateUserGroupDialog.svelte:143-153`; `AddUsersDialog.svelte:53-69`
- **Problem:** Both group dialogs `await` inside `for` loops over `pendingAdditions` and `pendingRemovals`, so N membership changes cost N sequential round-trips. `AddUsersDialog.handleCreate` runs its two `invalidateQueries` calls inside the per-email `Promise.all`, so adding ten users triggers twenty invalidations and the resulting refetch storm.
- **Recommendation:** Issue the membership mutations concurrently (bounded), and invalidate once after all of them settle. A batch endpoint would be better still, since the current design cannot report partial failure coherently.

**Verification:** CONFIRMED — sequential loops at `EditUserGroupDialog.svelte:153-168` and `CreateUserGroupDialog.svelte:146-153`; the in-loop invalidations at `AddUsersDialog.svelte:62-69`.

---

## Notes on things that are *not* problems

Recorded so a future reviewer does not re-open them:

- **Permission resolution fails closed.** `authTokenClaims` returns empty permission objects when the DB lookup errors (`claims.go:233-238, 270-275`), so a transient failure denies rather than grants.
- **The `ManageOrgAdmins` / `ManageProjectAdmins` split is applied consistently in the org-member and project-member handlers** — assignment and demotion of admins are both guarded (`organizations.go:439, 630, 656`, `projects.go:1176, 1362, 1427, 1453`). The two gaps are UM-06 (usergroup member removal) and UM-35 (the UI presenting actions it cannot perform).
- **Nav gating on `manageOrgMembers` in `users/+layout.svelte:19` is correct**, not overly strict: migration `0064.sql:10` grants that permission to `editor`, and `viewer`/`guest` genuinely lack `read_org_members`.
- **`AddProjectMemberUsergroup` intentionally permits managed groups** (unlike the org-level handlers, which reject them) — that is how "everyone in the org gets viewer on this project" is implemented via `autogroup:members`.
- **`getNextPageParam`'s `lastPage.nextPageToken !== ""` idiom** returns `undefined` when the field is absent, which TanStack correctly treats as "no more pages." It reads like a bug but is not one.
