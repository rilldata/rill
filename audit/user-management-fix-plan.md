# User Management — Fix Plan

Companion to [user-management.md](./user-management.md). Findings are referenced by their `UM-nn` ids.

The work is grouped into seven PR-sized batches. Batch 0 is a small, self-contained fix for the user-reported "removed users stay on screen" bug and should ship first. Batches 1 and 2 are independent of each other and of everything else; 3–6 can proceed in parallel once 1 is merged. Nothing here requires a proto change except where explicitly noted (batch 5, optional).

---

## Batch 0 — Make cache invalidation actually match (frontend)

**Ships:** UM-50
**Risk:** low. Purely additive — no existing invalidation is removed, so the worst case is an extra no-op invalidation.
**Why first:** this is the reported bug, it affects every mutation on the org user pages, and the fix is mechanical.

### 0.1 Add a helper that invalidates both key shapes

Orval emits the infinite key as exactly `["infinite", ...plainKey]` (`web-admin/src/client/gen/default/default.ts:3358-3374`), so one helper covers every list endpoint without touching generated code. Put it somewhere shared — `web-admin/src/client/invalidate.ts` or alongside the existing query utilities in `web-common/src/lib/svelte-query/`:

```ts
import type { QueryClient } from "@tanstack/svelte-query";

/**
 * Invalidates both the plain and the infinite variant of a generated list query key.
 * Orval prefixes infinite keys with the literal "infinite", and TanStack matches
 * keys by positional prefix, so a plain key never matches an infinite query.
 */
export function invalidateListQuery(
  queryClient: QueryClient,
  key: readonly unknown[],
) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: key }),
    queryClient.invalidateQueries({ queryKey: ["infinite", ...key] }),
  ]);
}
```

### 0.2 Route the 43 call sites through it

Replace each `queryClient.invalidateQueries({ queryKey: getAdminService…QueryKey(...) })` with `invalidateListQuery(queryClient, getAdminService…QueryKey(...))`. The sites are in:

- `features/organizations/user-management/utils.ts:24-51` (`invalidateAfterUserDelete` — five keys)
- `features/organizations/user-management/table/users/UserRoleCell.svelte:71-78, 102-109`
- `features/organizations/user-management/dialogs/AddUsersDialog.svelte:62-69`
- `features/organizations/user-management/dialogs/CreateUserGroupDialog.svelte:117-124, 155-165`
- `features/organizations/user-management/dialogs/EditUserGroupDialog.svelte:117-124, 171-185`
- `features/projects/user-management/UserSetRole.svelte:58-70, 91-103`
- `features/projects/user-management/UserAndGroupInviteForm.svelte:73-97`
- `features/projects/user-management/{UsergroupSetRole,OrgUserGroupSetRole,ProjectUserGroupSetRole}.svelte`
- plus the remaining sites in `ConvertGuestToMemberDialog.svelte`, `GroupRoleCell.svelte`, and `DeleteUserGroupConfirmDialog.svelte`

Note that `AddUsersDialog.svelte:62-69` invalidates inside a per-email `Promise.all`; fold UM-49's fix in here (invalidate once after all adds settle) rather than doubling the count.

**Verify with:** `grep -rn "invalidateQueries" web-admin/src/features/organizations/user-management web-admin/src/features/projects/user-management` should return nothing after the sweep.

### 0.3 Sweep the rest of the repo

The mismatch is not specific to user management — any feature pairing an infinite list with a plain-key invalidation has the same silent failure. Find the exposed cases:

```bash
# every component that renders an infinite query
grep -rln "Infinite(" web-admin/src web-common/src
# every invalidation, none of which use an infinite key today
grep -rn "getAdminService.*InfiniteQueryKey" web-admin/src | grep invalidate
```

Fix any other confirmed pairs in the same PR if the count is small; otherwise file a follow-up. Consider an ESLint `no-restricted-syntax` rule banning `invalidateQueries({ queryKey: getAdminService…QueryKey(...) })` in favour of the helper, so the trap cannot be re-entered.

**Test:** a `web-admin` vitest asserting that `invalidateListQuery` marks an infinite query stale. Run it from the `web-admin/` directory so the `@rilldata/web-admin` alias resolves:

```ts
const qc = new QueryClient();
const key = getAdminServiceListOrganizationMemberUsersQueryKey("acme");
qc.setQueryData(["infinite", ...key, { pageSize: 50 }], { pages: [] });
await invalidateListQuery(qc, key);
expect(qc.getQueryState(["infinite", ...key, { pageSize: 50 }])?.isInvalidated).toBe(true);
```

Also worth a Playwright case: remove a user and assert the row disappears **without** a reload.

---

## Batch 1 — Security and integrity guards (backend)

**Ships:** UM-01, UM-02, UM-03, UM-05, UM-06, UM-14
**Risk:** UM-01 is behaviour-changing for any org that already stores a reserved-key attribute; see the migration note.

### 1.1 Reserve the identity attribute keys (UM-01)

`admin/database/postgres/postgres.go` — in `validateAttributes`, reject the reserved set and return a validation error so it maps to `InvalidArgument` (this also closes UM-15):

```go
var reservedAttributeKeys = map[string]bool{
    "name": true, "email": true, "domain": true, "groups": true, "admin": true,
}

// in validateAttributes, inside the key loop:
if reservedAttributeKeys[strings.ToLower(key)] {
    return nil, database.NewValidationError(fmt.Sprintf("attribute key %q is reserved", key))
}
```

Convert the three existing `fmt.Errorf` returns in the same function to `database.NewValidationError` as well.

`admin/server/server.go` — make the ordering defensive in both `jwtAttributesForUser` and `jwtAttributesForService`: seed the map from the custom attributes, then write the built-ins over them, rather than the reverse. Validation alone is enough for new writes; the ordering change is what protects against rows already in the database.

```go
attr := make(map[string]any, len(attributes)+5)
for k, v := range attributes {
    attr[k] = v
}
attr["name"] = user.DisplayName
attr["email"] = user.Email
attr["domain"] = user.Email[strings.LastIndex(user.Email, "@")+1:]
attr["groups"] = groupNames
attr["admin"] = projectPermissions.ManageProject
```

**Before merging**, run an audit query against production and notify any affected orgs, since their security policies may currently depend on the override:

```sql
SELECT o.name AS org, u.email, uor.attributes
FROM users_orgs_roles uor
JOIN orgs o ON o.id = uor.org_id
JOIN users u ON u.id = uor.user_id
WHERE uor.attributes ?| array['name','email','domain','groups','admin'];
```

Apply the same reservation to `service.attributes` (written by `rill service` commands) — `jwtAttributesForService` has the identical override shape.

**Tests:** extend `admin/server/rbac_test.go` (which already exercises `UpdateOrganizationMemberUserAttributes` at `:1561-1592`) with a case asserting that `admin` is rejected, and a unit test on `jwtAttributesForUser` asserting a stored `admin: true` attribute does not survive into the returned map.

### 1.2 Fix the nil dereference in `SetProjectMemberUsergroupRole` (UM-02)

`admin/server/usergroups.go:520-527` — stop tolerating `ErrNotFound`:

```go
if keepExistingRestrictions {
    ug, err := s.admin.DB.FindProjectMemberUsergroup(ctx, usergroup.ID, proj.ID)
    if err != nil {
        if errors.Is(err, database.ErrNotFound) {
            return nil, status.Error(codes.NotFound, "user group is not a member of the project")
        }
        return nil, err
    }
    restrictResources = ug.RestrictResources
    resources = ug.Resources
}
```

While in this function, also fix UM-22 (`usergroups.go:512`, "remove" → "assign") and drop the redundant second `FindProjectMemberUsergroupRole` call at `:529` when `req.Role == nil` — `role` already holds that value from `:502`.

**Test:** a handler test that calls `SetProjectMemberUsergroupRole` with only `role` set, for a group that has no project membership, and asserts `codes.NotFound`.

### 1.3 Guard account deletion (UM-03)

`admin/server/users.go:525-546` — before `DB.DeleteUser`, walk the user's org memberships and reject when the deletion would orphan an org:

```go
orgs, err := s.admin.DB.FindOrganizationsForUser(ctx, user.ID, ...)
for _, org := range orgs {
    _, isLastAdmin, err := s.admin.DB.FindOrganizationMemberUserAdminStatus(ctx, org.ID, user.ID)
    if err != nil { return nil, err }
    if isLastAdmin {
        return nil, status.Errorf(codes.FailedPrecondition,
            "you are the last admin of organization %q; transfer ownership or delete the organization first", org.Name)
    }
    if org.BillingEmail == user.Email {
        return nil, status.Errorf(codes.FailedPrecondition,
            "you are the billing contact for organization %q; update the billing email first", org.Name)
    }
}
```

Confirm the exact "list orgs for user" helper available on the `database.DB` interface and reuse it rather than adding one. Keep the superuser force-access path able to bypass this (a superuser cleaning up genuinely abandoned accounts needs it) — gate the bypass on `req.SuperuserForceAccess` explicitly.

**Test:** create an org with one admin, delete that user, assert `FailedPrecondition`; add a second admin, assert deletion succeeds.

### 1.4 Return `NotFound` for attribute updates on non-members (UM-05)

`admin/server/organizations.go:731` — use the returned bool:

```go
updated, err := s.admin.DB.UpdateOrganizationMemberUserAttributes(ctx, org.ID, user.ID, attributes)
if err != nil {
    return nil, err
}
if !updated {
    return nil, status.Error(codes.NotFound, "user is not a member of the organization")
}
```

### 1.5 Symmetrise `RemoveUsergroupMemberUser` (UM-06)

`admin/server/usergroups.go:733-765` — insert the admin-group guard copied from `AddUsergroupMemberUser:604-610`, and add the pending-invite branch:

```go
user, err := s.admin.DB.FindUserByEmail(ctx, req.Email)
if err != nil {
    if !errors.Is(err, database.ErrNotFound) {
        return nil, err
    }
    invite, err := s.admin.DB.FindOrganizationInvite(ctx, group.OrgID, req.Email)
    if err != nil {
        return nil, err
    }
    if i := slices.Index(invite.UsergroupIDs, group.ID); i >= 0 {
        invite.UsergroupIDs = slices.Delete(invite.UsergroupIDs, i, i+1)
        if err := s.admin.DB.UpdateOrganizationInviteUsergroups(ctx, invite.ID, invite.UsergroupIDs); err != nil {
            return nil, err
        }
    }
    return &adminv1.RemoveUsergroupMemberUserResponse{}, nil
}
```

**Test:** as an org editor, attempt to remove a member from a group holding the org admin role and assert `PermissionDenied`; add and then remove a pending invitee from a group and assert the invite's `usergroup_ids` is empty.

### 1.6 Authenticate `ListRoles` (UM-14)

`admin/server/roles.go:11` — add the standard anon check. Confirm no unauthenticated caller depends on it first (the web-admin login page does not; `grep` for `ListRoles` in `web-admin/src` and `cli/`).

---

## Batch 2 — Quota and lifecycle correctness (backend)

**Ships:** UM-07, UM-08, UM-09, UM-10, UM-11, UM-12, UM-13
**Risk:** UM-08 changes reported seat counts; coordinate with whoever owns billing reconciliation.

### 2.1 Scope the internal-domain exclusion to email only (UM-08)

`admin/database/postgres/postgres.go:2014-2038` — the `searchPattern`/`negateSearch` pair is doing two unrelated jobs. Split them: keep `searchPattern` matching email **or** display name for the interactive list, and add a separate `excludeEmailPattern string` parameter that matches `lower(u.email)` only. Update the two internal-domain callers (`organizations.go:504`, `admin/jobs/river/billing_reporter.go:39`) to use the new parameter, and drop `negateSearch`.

Before deploying, quantify the impact:

```sql
SELECT COUNT(*) FROM users
WHERE lower(display_name) LIKE '%@rilldata.com' AND lower(email) NOT LIKE '%@rilldata.com';
```

### 2.2 Enforce the seat quota on every membership-creating path (UM-07, UM-13)

Move the check out of the handler and into `admin.InsertOrganizationMemberUser` (`admin/user.go:19-39`), which is the single choke point for `users_orgs_roles` inserts — it is called by `AddOrganizationMemberUser`, by `InsertProjectMemberUser` (guest path), by invite acceptance, and by whitelist auto-join. Inside the existing transaction, before the insert:

```go
if !ifNotExists || !alreadyMember {
    org, err := s.DB.FindOrganization(ctx, orgID)
    ...
    seats, err := s.DB.CountOrganizationMemberUsers(ctx, orgID, "", internalEmailExcludePattern)
    if org.QuotaSeats >= 0 && seats >= org.QuotaSeats {
        return fmt.Errorf("%w: org %q is limited to %d seats", ErrQuotaExceeded, org.Name, org.QuotaSeats)
    }
}
```

Add an `ErrQuotaExceeded` sentinel and map it to `FailedPrecondition` in `mapGRPCError` so the message survives to the client. Then delete the now-duplicated check in `AddOrganizationMemberUser:503-510`.

Decide explicitly — and comment — whether invite acceptance may exceed the cap. Recommendation: **allow** it (the seat was notionally reserved when the invite was issued) but **enforce** it on whitelist auto-join, which is otherwise unbounded. Both branches live in `CreateOrUpdateUser` (`admin/user.go:197-234` vs `:242-257`), so pass a flag through `InsertOrganizationMemberUser`.

### 2.3 Only charge the invite quota when an invite is created (UM-09)

`admin/server/organizations.go:427-433` and `admin/server/projects.go:1159-1170` — move the `CountInvitesForOrganization` check inside the `errors.Is(err, database.ErrNotFound)` branch, immediately before the corresponding `Insert*Invite` call. In the project handler, the `FindOrganization` call at `:1164` is needed later for the email URLs, so leave it where it is and move only the count and the comparison.

### 2.4 Do not fail the request when only the email fails (UM-10)

Four sites: `organizations.go:486-496` and `:521-531`, `projects.go:1257-1268` and `:1294-1305`. Replace `return nil, err` with a warning log and a successful response:

```go
if err != nil {
    s.logger.Warn("failed to send org invite email",
        zap.String("org", org.Name), zap.String("email", req.Email), zap.Error(err))
}
```

This is the fix that makes UM-37's error message meaningful, so land it before the frontend batch.

### 2.5 Clean up orphaned guest memberships (UM-11)

`admin/server/projects.go:1367` — after `DeleteProjectMemberUser`, drop the org guest membership when nothing else justifies it. Do it transactionally in a new `admin.DeleteProjectMemberUser` service helper, mirroring `InsertProjectMemberUser`:

```go
// after deleting the project membership
role, err := s.DB.FindOrganizationMemberUserRole(ctx, orgID, userID)   // confirm exact helper name
if err == nil && role.Guest {
    remaining, err := s.DB.CountProjectMemberUsersForUserInOrganization(ctx, orgID, userID)
    if err == nil && remaining == 0 {
        err = s.DeleteOrganizationMemberUser(ctx, orgID, userID)
    }
}
```

`CountProjectMemberUsersForUserInOrganization` does not exist yet — add it, and make sure it counts group-derived project access too, or a user who reaches the project only through a usergroup will be dropped from the org incorrectly. If that turns out to be awkward, the safer interim fix is a scheduled reconciliation job that removes guest memberships with no project access; note that decision in the code.

**Test:** add a user to a project only (creating the guest row), remove them, assert they are no longer an org member; repeat with the user also in a usergroup that has project access and assert they are retained.

### 2.6 Allow re-affirming the last admin (UM-12)

`admin/server/organizations.go:659-661` — skip the `isLastAdmin` rejection when `role.Admin` is true.

---

## Batch 3 — Server-side search and filtering (frontend)

**Ships:** UM-30, UM-33, UM-34, UM-42
**Depends on:** nothing; can start immediately.

### 3.1 Push search and role filters into the query (UM-30)

`web-admin/src/features/organizations/user-management/selectors.ts` — extend `getOrgUserMembers` to take the search text and role filter and include them in the query params so TanStack keys on them:

```ts
export function getOrgUserMembers({ organization, guestOnly, searchPattern, role }: {...}) {
  return createAdminServiceListOrganizationMemberUsersInfinite(
    organization,
    {
      pageSize: INFINITE_PAGE_SIZE,
      includeCounts: true,
      ...(guestOnly ? { role: OrgUserRoles.Guest } : role && role !== "all" ? { role } : {}),
      ...(searchPattern ? { searchPattern: `%${searchPattern}%` } : {}),
    },
    { query: { getNextPageParam: ... } },
  );
}
```

In `users/+page.svelte` and `guests/+page.svelte`, debounce `searchText` (300ms, matching the group dialogs) and feed it into the selector. Reduce `filteredUsers` to the members/pending split only — that distinction is derivable from the loaded rows and has no server equivalent. Use `pages[0].totalCount` for any count label.

Note that `ListOrganizationInvites` accepts neither a search pattern nor a role, so invites still need client-side filtering. Keep that, and make it explicit in a comment; batch 5 covers adding the parameters if this proves insufficient.

`groups/+page.svelte` — the same treatment is not available (`ListOrganizationMemberUsergroups` has no search parameter), so at minimum switch `groups_total_count` to a server-provided total and add a comment recording that group search is loaded-pages-only until the endpoint gains a filter.

### 3.2 Filter invites by role on the guests and members pages (UM-33)

In `coerceInvitesToUsers`, or immediately after it, keep only `roleName === OrgUserRoles.Guest` on the guests page, and only `roleName !== OrgUserRoles.Guest` on the members page (which already excludes guest *members* at `users/+page.svelte:83`).

### 3.3 Detect invites structurally (UM-34)

Tag the rows rather than relying on an optional field:

```ts
function coerceInvitesToUsers(invites: V1OrganizationInvite[]) {
  return invites.map((invite) => ({
    ...invite,
    isInvite: true as const,
    userEmail: invite.email,
    roleName: invite.roleName,
  }));
}
```

Replace all six `"invitedBy" in user` checks (`users/+page.svelte:99,105`, `guests/+page.svelte:97,103`, `OrgUsersTable.svelte:60,125`) with `user.isInvite === true`, and widen the `OrgUser` interface accordingly.

### 3.4 Fix the group counts (UM-42)

`selectors.ts:146-171` — take `groupsCount` from the usergroups response's total rather than `.filter(...).length` over one page. This requires a managed/unmanaged distinction the endpoint does not currently make; until it does, either (a) request `pageSize` large enough to be safe and comment the limit, or (b) add an `exclude_managed` filter to `ListOrganizationMemberUsergroups` (batch 5). Prefer (b).

`getUserGroupsForUsersInOrg:21-76` — the `usersCount` lookup against a 50-row page should be replaced by requesting counts for the specific groups the user belongs to. If no such endpoint exists, drop the count from the cell rather than displaying a wrong `0`.

---

## Batch 4 — Dialog and permission correctness (frontend)

**Ships:** UM-31, UM-32, UM-35, UM-36, UM-37, UM-38, UM-43, UM-44, UM-45, UM-46, UM-47

### 4.1 Fix the group edit save order (UM-31)

`EditUserGroupDialog.svelte` — make `handleRename` rethrow, thread the effective name through, and only close on success:

```js
async onUpdate({ form }) {
  if (!form.valid) return;
  const newName = form.data.newName;
  try {
    await applyPendingChanges(groupName);          // members first, under the current name
    if (newName !== groupName) {
      await handleRename(groupName, newName);       // rename last
    }
    open = false;
  } catch (error) {
    // handleRename / applyPendingChanges already surfaced the toast
  }
}
```

Applying members first is preferable to renaming first because a failed rename then leaves the membership edits intact, which is the less surprising partial state. Pass the group name explicitly into `applyPendingChanges(usergroup)` instead of closing over the prop. Skip the rename entirely when the name is unchanged (this also removes the spurious "renamed" toast on every save).

### 4.2 Sync the edit dialog's form to the selected group (UM-32)

Add `$: if (open) { $form.newName = groupName; }` (guarded so it runs on open, not on every keystroke), and change `handleClose` to reset to `groupName` rather than the stale `initialValues.newName`. Alternatively, mount the dialog behind `{#if userGroupName}` on `groups/+page.svelte:113-117`, matching what `users/+page.svelte:208` already does — that is the smaller change and fixes both the empty-name and the stale-name symptoms.

### 4.3 Thread the real project-member permission (UM-35)

`ShareProjectForm.svelte` — add a `manageProjectMembers: boolean` prop, pass it from `ShareProjectPopover` (and from every other caller of the form), and replace the two hardcoded `manageProjectMembers={true}` literals at `:350` and `:360`.

`UserSetRole.svelte:125` — the caret is hidden for admin targets when the caller lacks `manageProjectAdmins`, but the menu still opens. Either suppress the whole dropdown in that case (render the read-only branch) or disable the individual items. Suppressing the dropdown is consistent with how `UserRoleCell` handles the org side via `canManageOrgUser`.

### 4.4 Centralise error message extraction (UM-36, UM-37)

Promote `web-admin/src/features/projects/user-management/errors.ts` to a shared helper — it is currently unused despite being correct:

```ts
export function getErrorMessage(error: unknown): string {
  if (isAxiosError<RpcStatus>(error)) {
    return error.response?.data?.message ?? error.message;
  }
  return error instanceof Error ? error.message : String(error);
}
```

Replace all twelve `error.response.data.message` sites with `getErrorMessage(error)`.

`AddUsersDialog.svelte` — capture the error alongside the email and group the summary by outcome. Only `AlreadyExists` should render `users_already_member`; everything else should render the server message. This depends on batch 2.4, without which a successful add that fails to email is still reported as a failure.

### 4.5 Align the group-name validator (UM-38)

`UserAndGroupInviteForm.svelte:151-158` — import `SLUG_REGEX` from `features/organizations/user-management/constants.ts` and use it in place of the local regex. Consider moving `SLUG_REGEX` somewhere neutral, since it is now used from both the organizations and projects feature directories.

### 4.6 Broaden invalidation after member removal (UM-43)

`utils.ts:20-52` — add invalidations for `ListProjectMemberUsers`, `ListProjectInvites`, `ListProjectMemberUsergroups`, and `ListOrganizationMemberUsergroups`, since `DeleteOrganizationMemberUser` cascades to all of them. Replace the three hardcoded `autogroup:*` literals with exported constants and reuse them at `ShareProjectForm.svelte:191, 213, 278, 404`.

### 4.7 Route billing-contact removal to its dialog (UM-44)

`UserRoleCell.svelte` — accept `onAttemptRemoveBillingContactUser` (the page already passes it to the table) and guard `handleRemove` the same way `handleSetRole:50-53` guards role changes.

### 4.8 Input handling in the invite box (UM-45, UM-46)

`SearchAndInviteInput.svelte`:
- Move `loading = true/false` inside the debounced function so the flag tracks the real request.
- Add a monotonically increasing request id; discard a response whose id is not the latest before assigning `searchResults`.
- In the Enter handler at `:226-244`, assign `error = processError` and leave `input` untouched when validation fails.

### 4.9 Navigate instead of mutating the URL (UM-47)

`groups/+page.svelte:44-55` — replace the `searchParams.delete` calls with `goto(copyWithAdditionalArguments($page.url, {}, { action: false, groupName: false }), { replaceState: true })`, matching `ShareProjectPopover.svelte:26-34`.

---

## Batch 5 — Endpoint gaps (proto, optional)

**Ships:** the server-side support that batch 3 works around.

Only take this on if batch 3's workarounds prove inadequate in practice. Each addition is backwards-compatible (new optional request fields), but requires `make proto.generate` and touches the CLI-facing API, so it carries the long-term compatibility obligation noted in `CLAUDE.md`.

1. `ListOrganizationInvitesRequest` — add `role` and `search_pattern`, mirroring `ListOrganizationMemberUsersRequest`. Removes the client-side invite filtering that batch 3.2 leaves in place.
2. `ListOrganizationMemberUsergroupsRequest` — add `search_pattern` and `exclude_managed`. Fixes the group search and the group count (UM-42) properly.
3. Consider a batch membership endpoint (`AddUsergroupMemberUsers` taking a list) to replace the sequential loops in UM-49 and give the dialogs a coherent partial-failure result.

---

## Batch 6 — Cleanup

**Ships:** UM-15 (if not already taken in 1.1), UM-16, UM-17, UM-18, UM-19, UM-20, UM-21, UM-48, UM-49

Low risk, no user-visible behaviour change except where noted. Reasonable to fold into whichever of the earlier batches touches the same file.

- **UM-16** — thread the target user id into `findUserAuthTokenFuzzy` when `SuperuserForceAccess` is set (`users.go:723-786`), or document the self-only limitation in the CLI help.
- **UM-17** — make `UpdateUserOptions` fields optional pointers and emit a partial `UPDATE` in `postgres.go:932-940`, removing the three read-modify-write copies. This touches the login path, so ship it on its own.
- **UM-18** — delete `CountOrganizationInvites` from `database.DB` and `postgres.go:2620-2626`.
- **UM-19** — generate `UnionOrgRoles`/`UnionProjectRoles` by reflection, and derive the two "org admin gets everything" literals from a single `allProjectPermissions()` helper so a new proto field cannot be silently omitted from five of six places.
- **UM-20** — log the `structpb.NewStruct` failure in `orgMemberUserToPB` instead of discarding it.
- **UM-21** — wrap `UpdateUsergroup`'s two writes in `NewTx`, or collapse them into one partial `UPDATE`.
- **UM-48** — delete `readPublicEmailDomains.ts`, the unused `AddUsersDialog` instance on the guests page, the `email`/`role`/`isSuperUser` props and their self-assignments, the `"guests"` union member and its dead branches, `allOrgMemberUsersRows`, and `accessType`'s overwritten initialiser. `errors.ts` is kept and promoted by batch 4.4.
- **UM-49** — replace the sequential membership loops with bounded concurrency and a single invalidation after settlement; move `AddUsersDialog`'s two `invalidateQueries` calls out of the per-email `Promise.all`.

---

## Suggested sequencing

| Order | Batch | Why here |
|---|---|---|
| 0 | Batch 0 | The user-reported bug. Small, low-risk, and independent of everything else. |
| 1 | Batch 1 | Contains the only privilege-escalation finding and the only crash. UM-01 needs a production data audit first, so start it early. |
| 2 | Batch 2 | Independent of batch 1; 2.4 unblocks the frontend error-message work. UM-08 needs billing sign-off, so raise it now. |
| 3 | Batch 3 | Highest user-visible impact — search is currently wrong for any org over 50 members. |
| 4 | Batch 4 | Largest batch; split into two PRs (dialogs / permissions+errors) if review load is a concern. |
| 5 | Batch 6 | Fold pieces into 1–4 opportunistically; the remainder ships on its own. |
| 6 | Batch 5 | Only if batch 3's workarounds are not good enough. |

## Test coverage to add

The existing suites — `admin/server/rbac_test.go`, `rbac_old_test.go`, `users_test.go`, `roles_test.go`, `admin/database/postgres/postgres_test.go` — cover the happy paths well and already exercise the attribute endpoint. The gaps this plan should close:

- **Attribute reservation** (UM-01): rejection of each reserved key, and an assertion that a pre-existing reserved-key row does not reach the JWT.
- **Group-role edge cases** (UM-02, UM-06): setting a project role for a non-member group; an editor removing a member from an admin-role group; adding and removing a pending invitee from a group.
- **Lifecycle invariants** (UM-03, UM-11, UM-12): last-admin account deletion; guest cleanup after project removal (both with and without group-derived access); re-affirming the last admin's role.
- **Quotas** (UM-07, UM-09, UM-13): seat enforcement via the project path; adding an existing user while at the invite cap; whitelist auto-join at the seat cap.
- **Frontend** (UM-31, UM-32, UM-34): a `web-admin` vitest for the group-edit save ordering, one for the dialog's form seeding, and one for invite detection when `invitedBy` is absent. Run these from the `web-admin/` directory so the `@rilldata/web-admin` alias resolves.
