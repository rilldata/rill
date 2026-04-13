# Superuser Console Code Review Fixes

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address all 11 code review items for the superuser console PR #9083 in a single pass.

**Architecture:** The superuser console lives at `/-/superuser/` with 6 page files, shared layout, and selector modules. Changes span: proto definition (add `superuser` to `GetCurrentUserResponse`), Go handler, layout refactor (adopt `ContentContainer` + `LeftNav` pattern from Settings), dialog refactor (dedicated components + `AlertDialogGuardedConfirmation` for destructive actions), query key fixes, UI polish, and PR description update.

**Tech Stack:** Go, protobuf, Svelte 4, TypeScript, TanStack Query, Tailwind CSS, Orval-generated API clients

---

## Task 1: Add `superuser` field to `GetCurrentUserResponse` proto + handler

**Files:**
- Modify: `proto/rill/admin/v1/api.proto:2375-2378`
- Modify: `admin/server/users.go:104-128`

This replaces the `ListSuperusers` workaround. The frontend will check `$user.data?.superuser` instead of calling `ListSuperusers`.

- [ ] **Step 1: Add the field to the proto message**

In `proto/rill/admin/v1/api.proto`, add `bool superuser = 3;` to `GetCurrentUserResponse`:

```protobuf
message GetCurrentUserResponse {
  User user = 1;
  UserPreferences preferences = 2;
  bool superuser = 3;
}
```

- [ ] **Step 2: Run proto generation**

Run: `make proto.generate`
Expected: Clean generation with updated Go and TypeScript bindings.

- [ ] **Step 3: Update the Go handler**

In `admin/server/users.go`, update `GetCurrentUser` to set the new field using `claims.Superuser(ctx)`:

```go
func (s *Server) GetCurrentUser(ctx context.Context, req *adminv1.GetCurrentUserRequest) (*adminv1.GetCurrentUserResponse, error) {
	// Return an empty result if not authenticated.
	claims := auth.GetClaims(ctx)
	if claims.OwnerType() == auth.OwnerTypeAnon {
		return &adminv1.GetCurrentUserResponse{}, nil
	}

	// Error if authenticated as anything other than a user
	if claims.OwnerType() != auth.OwnerTypeUser {
		return nil, status.Error(codes.Unauthenticated, "not authenticated as a user")
	}

	// Owner is a user
	u, err := s.admin.DB.FindUser(ctx, claims.OwnerID())
	if err != nil {
		return nil, err
	}

	return &adminv1.GetCurrentUserResponse{
		User:       s.userToPB(u, true),
		Preferences: &adminv1.UserPreferences{
			TimeZone: &u.PreferenceTimeZone,
		},
		Superuser: claims.Superuser(ctx),
	}, nil
}
```

- [ ] **Step 4: Verify Go compiles**

Run: `go build ./admin/server/...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add proto/rill/admin/v1/api.proto admin/server/users.go
# Also add any generated files from proto.generate
git add web-admin/src/client/gen/ admin/
git commit -m "feat: add superuser field to GetCurrentUserResponse"
```

---

## Task 2: Update frontend to use `superuser` field instead of `ListSuperusers`

**Files:**
- Modify: `web-admin/src/routes/-/superuser/+layout.ts`
- Modify: `web-admin/src/features/authentication/AvatarButton.svelte`

- [ ] **Step 1: Simplify the layout guard**

Replace the `ListSuperusers` check in `web-admin/src/routes/-/superuser/+layout.ts` with a check on `GetCurrentUserResponse.superuser`. The exact field name in the generated types will be available after proto generation (likely `superuser`); verify before coding.

```typescript
import {
  adminServiceGetCurrentUser,
  getAdminServiceGetCurrentUserQueryKey,
  type V1GetCurrentUserResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async () => {
  let currentUserEmail: string | undefined;
  let isSuperuser = false;
  try {
    const userResp = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
      queryKey: getAdminServiceGetCurrentUserQueryKey(),
      queryFn: () => adminServiceGetCurrentUser(),
      staleTime: 5 * 60 * 1000,
    });
    currentUserEmail = userResp.user?.email;
    isSuperuser = userResp.superuser ?? false;
  } catch (e) {
    if (isAxiosError(e) && e.response?.status === 401) {
      redirectToLogin();
    }
    throw redirect(307, "/");
  }

  if (!currentUserEmail || !isSuperuser) {
    throw redirect(307, "/");
  }

  return { currentUserEmail };
};
```

- [ ] **Step 2: Simplify AvatarButton superuser check**

In `web-admin/src/features/authentication/AvatarButton.svelte`, remove the `createAdminServiceListSuperusers` query and derive `isSuperuser` from the existing `$user` query:

Remove these lines:
```svelte
const superusers = createAdminServiceListSuperusers({ ... });
$: isSuperuser = $superusers.isSuccess && ...;
```

Replace with:
```svelte
$: isSuperuser = $user.data?.superuser ?? false;
```

Remove the `createAdminServiceListSuperusers` import.

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd web-admin && npx svelte-check --tsconfig tsconfig.json` (or use `npm run check` if available)
Expected: No type errors in the modified files.

- [ ] **Step 4: Commit**

```bash
git add web-admin/src/routes/-/superuser/+layout.ts web-admin/src/features/authentication/AvatarButton.svelte
git commit -m "refactor: use GetCurrentUserResponse.superuser instead of ListSuperusers"
```

---

## Task 3: Refactor layout to use `ContentContainer` + `LeftNav`

**Files:**
- Modify: `web-admin/src/routes/-/superuser/+layout.svelte`
- Delete: `web-admin/src/features/superuser/layout/SuperuserSidebar.svelte`
- Delete: `web-admin/src/features/superuser/layout/SuperuserPageHeader.svelte`
- Modify: All 6 page files (remove `SuperuserPageHeader` usage)

The Settings pages use `ContentContainer` for the page title in the header area and `LeftNav` for the sidebar. We adopt the same pattern. "Superuser Console" becomes the `ContentContainer` title. Individual pages no longer need `SuperuserPageHeader` since `ContentContainer` provides the header.

- [ ] **Step 1: Rewrite the layout**

Replace `web-admin/src/routes/-/superuser/+layout.svelte` with the `ContentContainer` + `LeftNav` pattern, matching Settings layout at `web-admin/src/routes/[organization]/-/settings/+layout.svelte`:

```svelte
<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  const basePage = "/-/superuser";
  const baseRoute = "/-/superuser";

  const navItems = [
    { label: "Users", route: "" },
    { label: "Superusers", route: "/superusers" },
    { label: "Billing", route: "/billing" },
    { label: "Quotas", route: "/quotas" },
    { label: "Organizations", route: "/organizations" },
    { label: "Projects", route: "/projects" },
  ];
</script>

<svelte:head>
  <title>Superuser Console | Rill</title>
</svelte:head>

<ContentContainer title="Superuser Console" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav {basePage} {baseRoute} {navItems} minWidth="180px" />
    <div class="flex flex-col gap-y-6 w-full">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
```

Note: The `LeftNav` component does not support grouped headings natively. The Settings pattern uses a flat list. The sidebar groups ("People", "Billing & Plans", "Resources") are dropped in favor of a flat nav, matching the existing `LeftNav` API. The flat list is fine for 6 items.

- [ ] **Step 2: Remove `SuperuserPageHeader` from all pages**

In each of the 6 page files, remove the `SuperuserPageHeader` import and usage. The page title is now in the `ContentContainer` header. Keep the description text as a paragraph at the top of each page.

Files to update:
- `web-admin/src/routes/-/superuser/+page.svelte` — remove `SuperuserPageHeader`, add `<p class="text-sm text-fg-secondary">Search for users by email across all organizations.</p>`
- `web-admin/src/routes/-/superuser/superusers/+page.svelte` — same pattern
- `web-admin/src/routes/-/superuser/billing/+page.svelte` — same pattern
- `web-admin/src/routes/-/superuser/quotas/+page.svelte` — same pattern
- `web-admin/src/routes/-/superuser/organizations/+page.svelte` — same pattern
- `web-admin/src/routes/-/superuser/projects/+page.svelte` — same pattern

- [ ] **Step 3: Delete the old layout components**

Delete:
- `web-admin/src/features/superuser/layout/SuperuserSidebar.svelte`
- `web-admin/src/features/superuser/layout/SuperuserPageHeader.svelte`

If the `layout/` directory is now empty, delete it too.

- [ ] **Step 4: Verify no broken imports**

Run: `grep -r "SuperuserSidebar\|SuperuserPageHeader" web-admin/src/`
Expected: No results.

- [ ] **Step 5: Commit**

```bash
git add web-admin/src/routes/-/superuser/ web-admin/src/features/superuser/layout/
git commit -m "refactor: adopt ContentContainer + LeftNav layout for superuser console"
```

---

## Task 4: Replace generic dialog pattern with dedicated dialog components

**Files:**
- Create: `web-admin/src/features/superuser/dialogs/ConfirmActionDialog.svelte` (for non-destructive confirmations: assume user, hibernate, extend trial, save quotas)
- Create: `web-admin/src/features/superuser/dialogs/GuardedDeleteDialog.svelte` (for destructive actions: delete user, delete org, delete billing issue, redeploy)
- Modify: All 6 page files (replace dialog boilerplate with component usage)

The codebase pattern is dedicated dialog components with typed props. For destructive actions (Delete User, Delete Org), use `AlertDialogGuardedConfirmation` which requires typing to confirm.

- [ ] **Step 1: Create `ConfirmActionDialog.svelte`**

This component wraps a simple confirmation dialog for non-destructive actions (assume user, hibernate project, extend trial, save quotas):

```svelte
<!-- Non-destructive confirmation dialog for superuser actions -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let title: string;
  export let description: string;
  export let confirmLabel: string = "Confirm";
  export let loading = false;
  export let onConfirm: () => Promise<void>;

  let confirming = false;

  async function handleConfirm() {
    confirming = true;
    try {
      await onConfirm();
      open = false;
    } catch {
      // Keep dialog open for retry
    } finally {
      confirming = false;
    }
  }

  $: isLoading = loading || confirming;
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <AlertDialogDescription>{description}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button large class="font-normal" type="tertiary" onClick={() => (open = false)}>
        Cancel
      </Button>
      <Button
        large
        class="font-normal"
        type="primary"
        onClick={handleConfirm}
        loading={isLoading}
      >
        {confirmLabel}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
```

- [ ] **Step 2: Create `GuardedDeleteDialog.svelte`**

This wraps `AlertDialogGuardedConfirmation` for destructive actions:

```svelte
<!-- Type-to-confirm destructive dialog for superuser actions -->
<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { Button, type ButtonType } from "@rilldata/web-common/components/button";

  export let open = false;
  export let title: string;
  export let description: string;
  export let confirmText: string;
  export let confirmButtonText: string = "Delete";
  export let triggerLabel: string = "Delete";
  export let triggerType: ButtonType = "destructive";
  export let loading = false;
  export let error: string | undefined = undefined;
  export let onConfirm: () => Promise<void>;
</script>

<AlertDialogGuardedConfirmation
  bind:open
  {title}
  {description}
  {confirmText}
  {confirmButtonText}
  confirmButtonType="destructive"
  {loading}
  {error}
  {onConfirm}
>
  <Button type={triggerType} large class="font-normal">
    {triggerLabel}
  </Button>
</AlertDialogGuardedConfirmation>
```

- [ ] **Step 3: Refactor Users page (`+page.svelte`)**

Replace the 7 dialog state variables and generic `handleConfirm` with the two dialog components. Each action gets its own dialog instance with typed props.

Key changes:
- Remove: `dialogOpen`, `dialogTitle`, `dialogDescription`, `dialogDestructive`, `dialogAction`, `dialogLoading`, `handleConfirm`
- Add: `assumeDialogOpen`, `assumeEmail`, `deleteDialogOpen`, `deleteEmail`
- Use `ConfirmActionDialog` for assume; `GuardedDeleteDialog` for delete
- The `GuardedDeleteDialog` uses `AlertDialogGuardedConfirmation` which requires typing the user email to confirm deletion

Also fix empty-string fallbacks: disable buttons when `!user.email` instead of passing `user.email ?? ""`.

Also fix hardcoded Tailwind colors on the assumed-user banner: replace `bg-yellow-100 border-yellow-300 text-yellow-800` with semantic tokens or use the eventBus banner system (the `RepresentingBanner` already uses `eventBus.emit("add-banner", ...)` which handles theming). Since the in-page duplicate mirrors the `RepresentingBanner`, remove it and rely on `RepresentingBanner` alone (it's already mounted in the root layout and shows when `sessionStorage` has the assumed user).

- [ ] **Step 4: Refactor Organizations page**

Same pattern. Replace generic dialog with:
- `ConfirmActionDialog` for: Open as User, Hibernate Project
- `GuardedDeleteDialog` for: Delete Organization (requires typing org name), Redeploy Project
- Fix empty-string fallbacks: disable buttons when values are null

- [ ] **Step 5: Refactor Superusers page**

Replace the generic dialog with `ConfirmActionDialog` for Remove Superuser action (this is reversible so guarded confirmation is not needed). Fix empty-string fallback on `user.email`.

- [ ] **Step 6: Refactor Billing page**

Replace the generic dialog with:
- `ConfirmActionDialog` for: Extend Trial
- `GuardedDeleteDialog` for: Delete Billing Issue

- [ ] **Step 7: Refactor Quotas page**

Replace the generic dialog with `ConfirmActionDialog` for Save Quotas. This page is simpler since it only has one dialog action.

- [ ] **Step 8: Refactor Projects page**

Replace the generic dialog with:
- `ConfirmActionDialog` for: Hibernate Project
- `GuardedDeleteDialog` for: Redeploy Project (destructive/disruptive operation)
- The slots dialog is already a dedicated dialog pattern and can stay

- [ ] **Step 9: Verify no broken imports**

Run: `grep -r "dialogAction\|dialogDestructive\|dialogTitle.*=.*\"\"" web-admin/src/routes/-/superuser/`
Expected: No results (all generic dialog patterns removed).

- [ ] **Step 10: Commit**

```bash
git add web-admin/src/features/superuser/dialogs/ web-admin/src/routes/-/superuser/
git commit -m "refactor: replace generic dialog closures with dedicated dialog components"
```

---

## Task 5: Fix query invalidation to use generated key helpers

**Files:**
- Modify: `web-admin/src/routes/-/superuser/+page.svelte` (users)
- Modify: `web-admin/src/routes/-/superuser/superusers/+page.svelte`
- Modify: `web-admin/src/routes/-/superuser/billing/+page.svelte`
- Modify: `web-admin/src/routes/-/superuser/quotas/+page.svelte`

Replace all `predicate: (q) => q.queryKey[0] === "/v1/..."` patterns with Orval-generated key helpers.

- [ ] **Step 1: Fix Users page query invalidation**

Replace:
```typescript
await queryClient.invalidateQueries({
  predicate: (q) => q.queryKey[0] === "/v1/users/search",
});
```

With:
```typescript
import { getAdminServiceSearchUsersQueryKey } from "@rilldata/web-admin/client";

await queryClient.invalidateQueries({
  queryKey: getAdminServiceSearchUsersQueryKey(),
});
```

Note: `getAdminServiceSearchUsersQueryKey()` without params returns the base key prefix, which matches all `searchUsers` queries regardless of search params. Verify this by checking the generated function — it returns `["/v1/users/search", ...(params ? [params] : [])]`. Calling it without args returns `["/v1/users/search"]`, and TanStack Query does prefix matching by default, so this correctly invalidates all search queries.

- [ ] **Step 2: Fix Superusers page query invalidation**

Replace the two instances of:
```typescript
predicate: (q) => (q.queryKey[0] as string)?.includes("/v1/superuser")
```

With:
```typescript
import { getAdminServiceListSuperusersQueryKey } from "@rilldata/web-admin/client";

await queryClient.invalidateQueries({
  queryKey: getAdminServiceListSuperusersQueryKey(),
});
```

- [ ] **Step 3: Fix Billing page query invalidation**

Replace:
```typescript
predicate: (q) =>
  (q.queryKey[0] as string)?.includes("/v1/organizations") ||
  (q.queryKey[0] as string)?.includes("/v1/superuser/billing"),
```

With two targeted invalidations:
```typescript
import {
  getAdminServiceListOrganizationBillingIssuesQueryKey,
} from "@rilldata/web-admin/client";

await queryClient.invalidateQueries({
  queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(issuesOrg),
});
```

The broad `/v1/organizations` match was likely invalidating more than needed. Target just the billing issues for the specific org.

- [ ] **Step 4: Fix Quotas page query invalidation**

Replace:
```typescript
predicate: (q) =>
  (q.queryKey[0] as string)?.includes("/v1/superuser/quotas") ||
  (q.queryKey[0] as string)?.includes("/v1/organizations"),
```

With:
```typescript
import { getAdminServiceGetOrganizationQueryKey } from "@rilldata/web-admin/client";

await queryClient.invalidateQueries({
  queryKey: getAdminServiceGetOrganizationQueryKey(activeOrg),
});
```

The org query carries the quotas data (fetched via `getOrgForQuotas` which calls `createAdminServiceGetOrganization`). Invalidating the org query for the specific org is sufficient.

- [ ] **Step 5: Verify no string-based invalidation remains**

Run: `grep -n "queryKey\[0\]" web-admin/src/routes/-/superuser/`
Expected: No results.

- [ ] **Step 6: Commit**

```bash
git add web-admin/src/routes/-/superuser/
git commit -m "refactor: use Orval-generated query key helpers for cache invalidation"
```

---

## Task 6: Move "Superuser Console" above "View as" in avatar menu

**Files:**
- Modify: `web-admin/src/features/authentication/AvatarButton.svelte`

- [ ] **Step 1: Reorder the menu items**

In `AvatarButton.svelte`, move the superuser console block above the `ProjectAccessControls` block and give it its own separator:

```svelte
<DropdownMenu.Content>
  {#if isSuperuser}
    <DropdownMenu.Item href="/-/superuser">Superuser Console</DropdownMenu.Item>
    <DropdownMenu.Separator />
  {/if}

  {#if params.organization && params.project}
    <ProjectAccessControls ... >
      ...
    </ProjectAccessControls>
    {#if params.dashboard}
      ...
    {/if}
  {/if}

  <ThemeToggle />
  ...
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/authentication/AvatarButton.svelte
git commit -m "fix: move Superuser Console above View as in avatar menu"
```

---

## Task 7: Add org search caveat note to OrgPicker

**Files:**
- Modify: `web-admin/src/features/superuser/shared/OrgPicker.svelte`

- [ ] **Step 1: Add inline help text**

Add a brief note below the "No organizations found" message explaining the limitation:

In `OrgPicker.svelte`, update the "No organizations found" block:

```svelte
{:else if $orgNamesQuery.isSuccess}
  <div
    class="absolute z-10 left-0 right-0 mt-1 rounded-md border bg-surface-base shadow-md p-2"
  >
    <p class="text-sm text-fg-secondary">
      No organizations found. Note: orgs with zero projects won't appear in search.
    </p>
  </div>
{/if}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/superuser/shared/OrgPicker.svelte
git commit -m "fix: add caveat note about org search limitation in OrgPicker"
```

---

## Task 8: Fix hardcoded banner colors (already addressed in Task 4)

The in-page assumed-user banner in `+page.svelte` (Users page) uses hardcoded `bg-yellow-100 text-yellow-800` classes. This is addressed in Task 4, Step 3 by removing the duplicate banner and relying on the `RepresentingBanner` which uses the themed eventBus banner system.

If for some reason the in-page banner needs to stay (e.g., the eventBus banner doesn't show within the superuser console layout), replace hardcoded colors with semantic tokens:

```svelte
<div class="flex items-center gap-3 mb-4 px-4 py-2 rounded-md bg-warning-surface border border-warning-border text-warning-fg text-sm">
```

Verify the available semantic tokens in the Tailwind config. If `bg-warning-surface` doesn't exist, use `bg-surface-subtle` with a warning icon instead.

---

## Task 9: Move Delete User behind overflow menu (addressed via GuardedConfirmation in Task 4)

The reviewer suggested either overflow menu or `AlertDialogGuardedConfirmation`. Task 4 implements the latter (type-to-confirm for Delete User). This is the stronger safety mechanism; the overflow menu is optional polish.

If you want to also add the overflow menu: wrap the Delete button in a `DropdownMenu` with a three-dot trigger icon, placing "Delete User" as the only item. But given the guarded confirmation, this is not strictly necessary.

---

## Task 10: Update PR description

**Files:** None (GitHub API only)

- [ ] **Step 1: Update the PR body**

Use `gh pr edit` to update the description. The PR currently says `/-/admin/` (should be `/-/superuser/`) and lists 11 sections (only 6 shipped).

Run:
```bash
gh pr edit 9083 --body "$(cat <<'EOF'
Internal superuser console at `/-/superuser/`, giving CS and account managers a GUI for operations currently only available via `rill sudo` CLI commands.

- Route group `/-/superuser/` with auth guard (checks `GetCurrentUserResponse.superuser`)
- Uses `ContentContainer` + `LeftNav` layout pattern (matches Settings pages)
- 6 pages: Users (search/assume/delete), Superusers (list/add/remove), Billing (setup link/extend trial/issues), Quotas (org quotas), Organizations (lookup/members/projects), Projects (search/slots/hibernate/redeploy)
- Destructive actions (delete user, delete org) use type-to-confirm `AlertDialogGuardedConfirmation`
- "Superuser Console" link in avatar menu (above "View as"), visible only to superusers

**Checklist:**
- [ ] Covered by tests
- [ ] Ran it and it works as intended
- [ ] Reviewed the diff before requesting a review
- [ ] Checked for unhandled edge cases
- [ ] Linked the issues it closes
- [ ] Checked if the docs need to be updated. If so, create a separate Linear DOCS issue
- [ ] Intend to cherry-pick into the release branch
- [ ] I'm proud of this work!

---

*Developed in collaboration with Claude Code*
EOF
)"
```

- [ ] **Step 2: Commit** (no commit needed; this is a GitHub API operation)

---

## Task 11: Final verification

- [ ] **Step 1: Run Go tests for the handler change**

Run: `go test ./admin/server/...`
Expected: All tests pass.

- [ ] **Step 2: Run frontend lint/format**

Run: `npm run quality`
Expected: No lint or format errors. If there are formatting issues, fix them.

- [ ] **Step 3: Verify no unused imports**

Run: `grep -rn "SuperuserSidebar\|SuperuserPageHeader\|createAdminServiceListSuperusers" web-admin/src/`
Expected: Only the Orval-generated file and the superusers management page should reference `createAdminServiceListSuperusers`. No references to `SuperuserSidebar` or `SuperuserPageHeader`.

- [ ] **Step 4: Commit any formatting fixes**

```bash
git add -A
git commit -m "style: fix formatting from quality check"
```
