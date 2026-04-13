# Admin Console Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a super admin console within `web-admin` that provides non-technical staff (CS, account managers) a GUI for all 51 sudo CLI commands, deployed in phased rollouts behind a sidebar navigation layout.

**Architecture:** New SvelteKit route group at `/-/admin/` in `web-admin` with its own layout featuring a persistent sidebar. Superuser access is gated by calling `ListSuperusers` on layout load and checking if the current user is in the list. Each sidebar section maps to a sudo command group and uses the existing Orval-generated API clients with TanStack Query.

**Tech Stack:** Svelte 4, TypeScript, TanStack Query, Tailwind CSS, Orval-generated admin API clients

---

## Design Decisions

### Persona
Non-technical staff (CS reps, account managers). Engineers have access but aren't the design target. All actions should use plain language labels, confirmation dialogs for destructive operations, and inline success/error feedback.

### Layout
Option A: Sidebar Navigation. Persistent left sidebar listing all command groups. Content area on the right. Classic admin panel pattern (Stripe Dashboard, Django Admin).

### Phasing Strategy
- **Phase 1 (MVP):** Layout shell + Users + Billing + Nav link (highest CS frequency; nav link ensures discoverability from day one)
- **Phase 2:** Organizations + Quotas
- **Phase 3:** Projects + Whitelist
- **Phase 4:** Superuser Management + Annotations
- **Phase 5:** Virtual Files + Runtime + Clone (engineer-leaning)

### Auth Model
- Call `adminServiceListSuperusers()` in the `/-/admin/+layout.ts` load function
- Compare current user's email against the superuser list
- If not a superuser, redirect to `/` with an error
- All sudo API calls will also fail server-side with 403 if not a superuser; this is defense in depth

### Routing Convention
```
/-/admin/                    → Dashboard home (quick stats, recent actions)
/-/admin/users/              → User search, lookup, list
/-/admin/billing/            → Trial extensions, billing setup, repair
/-/admin/organizations/      → Org lookup, settings, custom domains
/-/admin/quotas/             → View and adjust quotas
/-/admin/projects/           → Search, edit, hibernate, reset
/-/admin/whitelist/          → Domain whitelist management
/-/admin/superusers/         → Add/remove superusers
/-/admin/annotations/        → Project annotations
/-/admin/virtual-files/      → Virtual file management
/-/admin/runtime/            → Runtime instances, manager tokens
```

### Component Naming Convention
All admin console components live under `web-admin/src/features/admin/`. Shared UI primitives (buttons, inputs, tables) come from `web-common`. Admin-specific components are scoped to their feature directory.

---

## File Structure

### New Files

```
web-admin/src/routes/-/admin/
  +layout.ts                           # Auth guard: superuser check
  +layout.svelte                       # Sidebar + content layout
  +page.svelte                         # Dashboard home

  users/+page.svelte                   # User management page
  billing/+page.svelte                 # Billing operations page
  organizations/+page.svelte           # Org management page
  quotas/+page.svelte                  # Quota management page
  projects/+page.svelte                # Project operations page
  whitelist/+page.svelte               # Whitelist management page
  superusers/+page.svelte              # Superuser management page
  annotations/+page.svelte             # Annotations page
  virtual-files/+page.svelte           # Virtual files page
  runtime/+page.svelte                 # Runtime page

web-admin/src/features/admin/
  layout/
    AdminSidebar.svelte                # Sidebar navigation component
    AdminPageHeader.svelte             # Consistent page header with title + description

  users/
    UserSearchForm.svelte              # Email/name search input
    UserResultCard.svelte              # User details display card
    UserActionsMenu.svelte             # Assume, open, delete actions
    selectors.ts                       # TanStack Query wrappers for user APIs

  billing/
    ExtendTrialForm.svelte             # Trial extension form
    BillingSetupForm.svelte            # Billing setup form
    BillingRepairButton.svelte         # Trigger billing repair
    BillingIssuesList.svelte           # List and delete billing issues
    selectors.ts                       # TanStack Query wrappers for billing APIs

  organizations/
    OrgLookupForm.svelte               # Org search/lookup
    OrgDetailsCard.svelte              # Org info display
    OrgActionsPanel.svelte             # Set custom domain, internal plan, join org
    selectors.ts                       # TanStack Query wrappers for org APIs

  quotas/
    QuotaLookupForm.svelte             # Org or user quota lookup
    QuotaEditor.svelte                 # Editable quota fields
    selectors.ts                       # TanStack Query wrappers for quota APIs

  projects/
    ProjectSearchForm.svelte           # Project search with filters
    ProjectResultsTable.svelte         # Search results table
    ProjectActionsPanel.svelte         # Edit, reset, hibernate actions
    selectors.ts                       # TanStack Query wrappers for project APIs

  whitelist/
    WhitelistForm.svelte               # Add/remove domain whitelist
    selectors.ts                       # TanStack Query wrappers

  superusers/
    SuperuserList.svelte               # List all superusers
    SuperuserForm.svelte               # Add/remove superuser
    selectors.ts                       # TanStack Query wrappers

  shared/
    ConfirmDialog.svelte               # Reusable confirmation dialog for destructive actions
    StatusBadge.svelte                 # Success/error/pending status indicator
    ActionResultBanner.svelte          # Inline success/error feedback after operations
    SearchInput.svelte                 # Reusable search input with debounce
```

---

## Phase 1: Layout Shell + Users + Billing

### Task 1: Admin Layout Shell — Superuser Auth Guard

**Files:**
- Create: `web-admin/src/routes/-/admin/+layout.ts`

- [ ] **Step 1: Write the layout load function with superuser check**

```typescript
// web-admin/src/routes/-/admin/+layout.ts
import {
  adminServiceGetCurrentUser,
  adminServiceListSuperusers,
  getAdminServiceGetCurrentUserQueryKey,
  getAdminServiceListSuperusersQueryKey,
  type V1GetCurrentUserResponse,
  type V1ListSuperusersResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async () => {
  // Get current user
  let currentUserEmail: string | undefined;
  try {
    const userResp = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
      queryKey: getAdminServiceGetCurrentUserQueryKey(),
      queryFn: () => adminServiceGetCurrentUser(),
      staleTime: 5 * 60 * 1000,
    });
    currentUserEmail = userResp.user?.email;
  } catch (e) {
    if (isAxiosError(e) && e.response?.status === 401) {
      // redirectToLogin() throws a SvelteKit redirect internally;
      // call it outside the catch to avoid swallowing the redirect exception
    } else {
      throw redirect(307, "/");
    }
    redirectToLogin();
  }

  if (!currentUserEmail) {
    throw redirect(307, "/");
  }

  // Check if current user is a superuser
  try {
    const superusersResp =
      await queryClient.fetchQuery<V1ListSuperusersResponse>({
        queryKey: getAdminServiceListSuperusersQueryKey(),
        queryFn: () => adminServiceListSuperusers(),
        staleTime: 5 * 60 * 1000,
      });

    const isSuperuser = superusersResp.users?.some(
      (u) => u.email === currentUserEmail,
    );

    if (!isSuperuser) {
      throw redirect(307, "/");
    }
  } catch (e) {
    // ListSuperusers itself will 403 if not a superuser
    if (isAxiosError(e) && e.response?.status === 403) {
      throw redirect(307, "/");
    }
    // Re-throw SvelteKit redirects
    throw e;
  }

  return { currentUserEmail };
};
```

- [ ] **Step 2: Verify the file compiles**

Run: `cd /Users/eokuma/rill && npx svelte-check --workspace web-admin 2>&1 | head -30`
Expected: No errors in the new file

- [ ] **Step 3: Commit**

```bash
git add web-admin/src/routes/-/admin/+layout.ts
git commit -m "feat(admin-console): add superuser auth guard for admin layout"
```

---

### Task 2: Admin Sidebar Component

**Files:**
- Create: `web-admin/src/features/admin/layout/AdminSidebar.svelte`

- [ ] **Step 1: Create the sidebar component**

```svelte
<!-- web-admin/src/features/admin/layout/AdminSidebar.svelte -->
<script lang="ts">
  import { page } from "$app/stores";

  const navGroups = [
    {
      heading: "Overview",
      items: [{ label: "Dashboard", href: "/-/admin" }],
    },
    {
      heading: "People",
      items: [
        { label: "Users", href: "/-/admin/users" },
        { label: "Superusers", href: "/-/admin/superusers" },
      ],
    },
    {
      heading: "Billing & Plans",
      items: [
        { label: "Billing", href: "/-/admin/billing" },
        { label: "Quotas", href: "/-/admin/quotas" },
      ],
    },
    {
      heading: "Resources",
      items: [
        { label: "Organizations", href: "/-/admin/organizations" },
        { label: "Projects", href: "/-/admin/projects" },
        { label: "Whitelist", href: "/-/admin/whitelist" },
      ],
    },
    {
      heading: "Advanced",
      items: [
        { label: "Annotations", href: "/-/admin/annotations" },
        { label: "Virtual Files", href: "/-/admin/virtual-files" },
        { label: "Runtime", href: "/-/admin/runtime" },
      ],
    },
  ];

  function isActive(href: string, pathname: string): boolean {
    if (href === "/-/admin") return pathname === "/-/admin";
    return pathname.startsWith(href);
  }
</script>

<nav class="sidebar">
  <div class="sidebar-header">
    <span class="logo-text">Admin Console</span>
  </div>

  <div class="sidebar-content">
    {#each navGroups as group}
      <div class="nav-group">
        <span class="group-heading">{group.heading}</span>
        {#each group.items as item}
          <a
            href={item.href}
            class="nav-item"
            class:active={isActive(item.href, $page.url.pathname)}
          >
            {item.label}
          </a>
        {/each}
      </div>
    {/each}
  </div>
</nav>

<style lang="postcss">
  .sidebar {
    @apply w-56 flex-shrink-0 border-r border-slate-200 dark:border-slate-700
      bg-white dark:bg-slate-900 flex flex-col h-full;
  }

  .sidebar-header {
    @apply px-4 py-4 border-b border-slate-200 dark:border-slate-700;
  }

  .logo-text {
    @apply text-sm font-semibold text-slate-900 dark:text-slate-100;
  }

  .sidebar-content {
    @apply flex-1 overflow-y-auto py-3 px-3;
  }

  .nav-group {
    @apply mb-4;
  }

  .group-heading {
    @apply text-[11px] font-semibold uppercase tracking-wider
      text-slate-400 dark:text-slate-500 px-2 mb-1 block;
  }

  .nav-item {
    @apply block px-2 py-1.5 text-sm rounded-md
      text-slate-600 dark:text-slate-300
      hover:bg-slate-100 dark:hover:bg-slate-800
      transition-colors;
  }

  .nav-item.active {
    @apply bg-slate-100 dark:bg-slate-800
      text-slate-900 dark:text-slate-100 font-medium;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/layout/AdminSidebar.svelte
git commit -m "feat(admin-console): add sidebar navigation component"
```

---

### Task 3: Admin Page Header Component

**Files:**
- Create: `web-admin/src/features/admin/layout/AdminPageHeader.svelte`

- [ ] **Step 1: Create the page header component**

```svelte
<!-- web-admin/src/features/admin/layout/AdminPageHeader.svelte -->
<script lang="ts">
  export let title: string;
  export let description: string = "";
</script>

<div class="page-header">
  <h1 class="page-title">{title}</h1>
  {#if description}
    <p class="page-description">{description}</p>
  {/if}
</div>

<style lang="postcss">
  .page-header {
    @apply mb-6;
  }

  .page-title {
    @apply text-xl font-semibold text-slate-900 dark:text-slate-100;
  }

  .page-description {
    @apply text-sm text-slate-500 dark:text-slate-400 mt-1;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/layout/AdminPageHeader.svelte
git commit -m "feat(admin-console): add page header component"
```

---

### Task 4: Shared Components — ConfirmDialog, StatusBadge, ActionResultBanner, SearchInput

**Files:**
- Create: `web-admin/src/features/admin/shared/ConfirmDialog.svelte`
- Create: `web-admin/src/features/admin/shared/StatusBadge.svelte`
- Create: `web-admin/src/features/admin/shared/ActionResultBanner.svelte`
- Create: `web-admin/src/features/admin/shared/SearchInput.svelte`

- [ ] **Step 1: Create ConfirmDialog**

```svelte
<!-- web-admin/src/features/admin/shared/ConfirmDialog.svelte -->
<script lang="ts">
  export let open = false;
  export let title: string;
  export let description: string = "";
  export let confirmLabel: string = "Confirm";
  export let cancelLabel: string = "Cancel";
  export let destructive: boolean = false;
  export let onConfirm: () => void | Promise<void>;

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await onConfirm();
      open = false;
    } finally {
      loading = false;
    }
  }

  function handleCancel() {
    open = false;
  }
</script>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="overlay" on:click={handleCancel}>
    <div class="dialog" on:click|stopPropagation>
      <h2 class="dialog-title">{title}</h2>
      {#if description}
        <p class="dialog-description">{description}</p>
      {/if}
      <div class="dialog-actions">
        <button class="btn-cancel" on:click={handleCancel} disabled={loading}>
          {cancelLabel}
        </button>
        <button
          class="btn-confirm"
          class:destructive
          on:click={handleConfirm}
          disabled={loading}
        >
          {#if loading}Working...{:else}{confirmLabel}{/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .overlay {
    @apply fixed inset-0 bg-black/50 flex items-center justify-center z-50;
  }

  .dialog {
    @apply bg-white dark:bg-slate-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl;
  }

  .dialog-title {
    @apply text-lg font-semibold text-slate-900 dark:text-slate-100;
  }

  .dialog-description {
    @apply text-sm text-slate-500 dark:text-slate-400 mt-2;
  }

  .dialog-actions {
    @apply flex justify-end gap-3 mt-6;
  }

  .btn-cancel {
    @apply px-4 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700;
  }

  .btn-confirm {
    @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700;
  }

  .btn-confirm.destructive {
    @apply bg-red-600 hover:bg-red-700;
  }

  button:disabled {
    @apply opacity-50 cursor-not-allowed;
  }
</style>
```

- [ ] **Step 2: Create StatusBadge**

```svelte
<!-- web-admin/src/features/admin/shared/StatusBadge.svelte -->
<script lang="ts">
  export let status: "success" | "error" | "pending" | "info" = "info";
  export let label: string;
</script>

<span class="badge" class:success={status === "success"}
  class:error={status === "error"} class:pending={status === "pending"}
  class:info={status === "info"}>
  {label}
</span>

<style lang="postcss">
  .badge {
    @apply inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full;
  }

  .success {
    @apply bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400;
  }

  .error {
    @apply bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400;
  }

  .pending {
    @apply bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400;
  }

  .info {
    @apply bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400;
  }
</style>
```

- [ ] **Step 3: Create ActionResultBanner**

```svelte
<!-- web-admin/src/features/admin/shared/ActionResultBanner.svelte -->
<script lang="ts">
  export let type: "success" | "error" | "" = "";
  export let message: string = "";

  export function show(newType: "success" | "error", newMessage: string) {
    type = newType;
    message = newMessage;
    setTimeout(() => {
      type = "";
      message = "";
    }, 5000);
  }
</script>

{#if type && message}
  <div
    class="banner"
    class:success={type === "success"}
    class:error={type === "error"}
  >
    <span>{message}</span>
    <button
      class="close-btn"
      on:click={() => {
        type = "";
        message = "";
      }}>x</button
    >
  </div>
{/if}

<style lang="postcss">
  .banner {
    @apply flex items-center justify-between px-4 py-3 rounded-md text-sm mb-4;
  }

  .success {
    @apply bg-green-50 text-green-800 dark:bg-green-900/20 dark:text-green-300;
  }

  .error {
    @apply bg-red-50 text-red-800 dark:bg-red-900/20 dark:text-red-300;
  }

  .close-btn {
    @apply ml-4 text-current opacity-50 hover:opacity-100;
  }
</style>
```

- [ ] **Step 4: Create SearchInput**

```svelte
<!-- web-admin/src/features/admin/shared/SearchInput.svelte -->
<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let placeholder: string = "Search...";
  export let value: string = "";
  export let debounceMs: number = 300;

  const dispatch = createEventDispatcher<{ search: string }>();

  let timeout: ReturnType<typeof setTimeout>;

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    value = target.value;
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      dispatch("search", value);
    }, debounceMs);
  }

  function handleSubmit() {
    clearTimeout(timeout);
    dispatch("search", value);
  }
</script>

<div class="search-container">
  <input
    type="text"
    class="search-input"
    {placeholder}
    {value}
    on:input={handleInput}
    on:keydown={(e) => e.key === "Enter" && handleSubmit()}
  />
</div>

<style lang="postcss">
  .search-container {
    @apply relative;
  }

  .search-input {
    @apply w-full px-3 py-2 text-sm rounded-md border border-slate-300
      dark:border-slate-600 bg-white dark:bg-slate-800
      text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 dark:placeholder:text-slate-500
      focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }
</style>
```

- [ ] **Step 5: Commit all shared components**

```bash
git add web-admin/src/features/admin/shared/
git commit -m "feat(admin-console): add shared UI components (ConfirmDialog, StatusBadge, ActionResultBanner, SearchInput)"
```

---

### Task 5: Admin Layout Svelte Component

**Files:**
- Create: `web-admin/src/routes/-/admin/+layout.svelte`

- [ ] **Step 1: Create the layout component with sidebar + content area**

```svelte
<!-- web-admin/src/routes/-/admin/+layout.svelte -->
<script lang="ts">
  import AdminSidebar from "@rilldata/web-admin/features/admin/layout/AdminSidebar.svelte";
</script>

<svelte:head>
  <title>Admin Console | Rill</title>
</svelte:head>

<div class="admin-layout">
  <AdminSidebar />
  <div class="admin-content">
    <slot />
  </div>
</div>

<style lang="postcss">
  .admin-layout {
    @apply flex h-screen overflow-hidden;
  }

  .admin-content {
    @apply flex-1 overflow-y-auto p-8;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/+layout.svelte
git commit -m "feat(admin-console): add admin layout with sidebar + content area"
```

---

### Task 6: Admin Dashboard Home Page

**Files:**
- Create: `web-admin/src/routes/-/admin/+page.svelte`

- [ ] **Step 1: Create the dashboard home page**

```svelte
<!-- web-admin/src/routes/-/admin/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
</script>

<AdminPageHeader
  title="Admin Console"
  description="Internal tools for managing users, billing, organizations, and more."
/>

<div class="quick-actions">
  <a href="/-/admin/users" class="action-card">
    <span class="action-title">Users</span>
    <span class="action-desc">Search, lookup, and manage user accounts</span>
  </a>
  <a href="/-/admin/billing" class="action-card">
    <span class="action-title">Billing</span>
    <span class="action-desc">Extend trials, repair billing, manage subscriptions</span>
  </a>
  <a href="/-/admin/organizations" class="action-card">
    <span class="action-title">Organizations</span>
    <span class="action-desc">Lookup orgs, set custom domains, manage plans</span>
  </a>
  <a href="/-/admin/projects" class="action-card">
    <span class="action-title">Projects</span>
    <span class="action-desc">Search projects, edit settings, hibernate or reset</span>
  </a>
  <a href="/-/admin/quotas" class="action-card">
    <span class="action-title">Quotas</span>
    <span class="action-desc">View and adjust organization and user quotas</span>
  </a>
  <a href="/-/admin/superusers" class="action-card">
    <span class="action-title">Superusers</span>
    <span class="action-desc">Manage superuser access</span>
  </a>
</div>

<style lang="postcss">
  .quick-actions {
    @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4;
  }

  .action-card {
    @apply block p-4 rounded-lg border border-slate-200 dark:border-slate-700
      hover:border-blue-300 dark:hover:border-blue-600
      hover:shadow-sm transition-all;
  }

  .action-title {
    @apply block text-sm font-semibold text-slate-900 dark:text-slate-100 mb-1;
  }

  .action-desc {
    @apply block text-xs text-slate-500 dark:text-slate-400;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/+page.svelte
git commit -m "feat(admin-console): add dashboard home page with quick action cards"
```

---

### Task 7: User Management — Selectors (API Layer)

**Files:**
- Create: `web-admin/src/features/admin/users/selectors.ts`

**Reference:** The generated API clients live in `web-admin/src/client/gen/default/default.ts`. Key functions:
- `createAdminServiceSearchUsers` — search users by email pattern
- `createAdminServiceGetUser` — get user by email
- `createAdminServiceIssueRepresentativeAuthToken` — assume user identity
- `createAdminServiceRevokeCurrentRepresentativeAuthToken` — unassume
- `createAdminServiceDeleteUser` — delete user (requires superuserForceAccess)

- [ ] **Step 1: Create the selectors file**

```typescript
// web-admin/src/features/admin/users/selectors.ts
import {
  createAdminServiceSearchUsers,
  createAdminServiceIssueRepresentativeAuthToken,
  createAdminServiceRevokeRepresentativeAuthTokens,
  createAdminServiceDeleteUser,
} from "@rilldata/web-admin/client";

export function searchUsers(emailQuery: string) {
  return createAdminServiceSearchUsers(
    { emailQuery },
    { query: { enabled: emailQuery.length >= 2 } },
  );
}

export function createAssumeUserMutation() {
  return createAdminServiceIssueRepresentativeAuthToken();
}

export function createUnassumeUserMutation() {
  return createAdminServiceRevokeRepresentativeAuthTokens();
}

export function createDeleteUserMutation() {
  return createAdminServiceDeleteUser();
}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/users/selectors.ts
git commit -m "feat(admin-console): add user management API selectors"
```

---

### Task 8: User Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/users/+page.svelte`

- [ ] **Step 1: Create the users page with search, results table, and action buttons**

```svelte
<!-- web-admin/src/routes/-/admin/users/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    searchUsers,
    createAssumeUserMutation,
    createDeleteUserMutation,
  } from "@rilldata/web-admin/features/admin/users/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let searchQuery = "";
  let bannerRef: ActionResultBanner;
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const assumeUser = createAssumeUserMutation();
  const deleteUser = createDeleteUserMutation();

  $: usersQuery = searchUsers(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleAssume(email: string) {
    confirmTitle = "Assume User Identity";
    confirmDescription = `You will browse Rill Cloud as ${email}. Use "Unassume" to return to your own identity.`;
    confirmDestructive = false;
    confirmAction = async () => {
      try {
        await $assumeUser.mutateAsync({ data: { email } });
        bannerRef.show("success", `Now browsing as ${email}`);
      } catch (err) {
        bannerRef.show("error", `Failed to assume user: ${err}`);
      }
    };
    confirmOpen = true;
  }

  function handleDelete(email: string) {
    confirmTitle = "Delete User";
    confirmDescription = `This will permanently delete the user ${email}. This action cannot be undone.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $deleteUser.mutateAsync({
          email,
          superuserForceAccess: true,
        });
        bannerRef.show("success", `User ${email} deleted`);
        await queryClient.invalidateQueries({
          predicate: (q) => q.queryKey[0] === "/v1/users/search",
        });
      } catch (err) {
        bannerRef.show("error", `Failed to delete user: ${err}`);
      }
    };
    confirmOpen = true;
  }

  async function handleOpenAsUser(email: string) {
    // Assume the user's identity first, then open the main page
    try {
      await $assumeUser.mutateAsync({ data: { email } });
      window.open("/", "_blank");
      bannerRef.show("success", `Opened as ${email} in a new tab. Remember to unassume when done.`);
    } catch (err) {
      bannerRef.show("error", `Failed to assume user: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Users"
  description="Search for users by email, assume their identity for debugging, or manage their accounts."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search by email (min 2 characters)..."
    on:search={handleSearch}
  />
</div>

{#if $usersQuery.isLoading && searchQuery.length >= 2}
  <p class="text-sm text-slate-500">Searching...</p>
{:else if $usersQuery.data?.users?.length}
  <div class="results-table">
    <table class="w-full">
      <thead>
        <tr>
          <th>Email</th>
          <th>Display Name</th>
          <th>Created</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each $usersQuery.data.users as user}
          <tr>
            <td class="font-mono text-xs">{user.email}</td>
            <td>{user.displayName ?? "-"}</td>
            <td class="text-xs text-slate-500">
              {user.createdOn
                ? new Date(user.createdOn).toLocaleDateString()
                : "-"}
            </td>
            <td>
              <div class="flex gap-2">
                <button
                  class="action-btn"
                  on:click={() => handleAssume(user.email ?? "")}
                >
                  Assume
                </button>
                <button
                  class="action-btn"
                  on:click={() => handleOpenAsUser(user.email ?? "")}
                >
                  Open as User
                </button>
                <button
                  class="action-btn destructive"
                  on:click={() => handleDelete(user.email ?? "")}
                >
                  Delete
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else if searchQuery.length >= 2 && $usersQuery.isSuccess}
  <p class="text-sm text-slate-500">No users found for "{searchQuery}"</p>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  table {
    @apply border-collapse;
  }

  th {
    @apply text-left text-xs font-medium text-slate-500 dark:text-slate-400
      uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700;
  }

  td {
    @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300
      border-b border-slate-100 dark:border-slate-800;
  }

  tr:hover td {
    @apply bg-slate-50 dark:bg-slate-800/50;
  }

  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700;
  }

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/users/+page.svelte
git commit -m "feat(admin-console): add user management page with search, assume, and delete"
```

---

### Task 9: Billing Management — Selectors (API Layer)

**Files:**
- Create: `web-admin/src/features/admin/billing/selectors.ts`

**Reference:** Key generated API functions:
- `createAdminServiceSudoExtendTrial` — extend trial period
- `createAdminServiceSudoSetOrganizationBillingCustomerId` — set billing customer
- `createAdminServiceSudoTriggerBillingRepair` — trigger billing repair
- `createAdminServiceSudoDeleteOrganizationBillingIssue` — delete billing issue
- `createAdminServiceGetBillingSubscription` — get current subscription

- [ ] **Step 1: Create the billing selectors**

```typescript
// web-admin/src/features/admin/billing/selectors.ts
import {
  createAdminServiceSudoExtendTrial,
  createAdminServiceSudoTriggerBillingRepair,
  createAdminServiceSudoDeleteOrganizationBillingIssue,
  createAdminServiceSudoUpdateOrganizationBillingCustomer,
  createAdminServiceListOrganizationBillingIssues,
} from "@rilldata/web-admin/client";

export function createExtendTrialMutation() {
  return createAdminServiceSudoExtendTrial();
}

export function createBillingRepairMutation() {
  return createAdminServiceSudoTriggerBillingRepair();
}

export function createDeleteBillingIssueMutation() {
  return createAdminServiceSudoDeleteOrganizationBillingIssue();
}

export function createSetBillingCustomerMutation() {
  return createAdminServiceSudoUpdateOrganizationBillingCustomer();
}

export function getBillingIssues(org: string) {
  return createAdminServiceListOrganizationBillingIssues(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/billing/selectors.ts
git commit -m "feat(admin-console): add billing management API selectors"
```

---

### Task 10: Billing Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/billing/+page.svelte`

- [ ] **Step 1: Create the billing page with trial extension, repair, and issue management**

```svelte
<!-- web-admin/src/routes/-/admin/billing/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    createExtendTrialMutation,
    createBillingRepairMutation,
    createDeleteBillingIssueMutation,
    createSetBillingCustomerMutation,
    getBillingIssues,
  } from "@rilldata/web-admin/features/admin/billing/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmAction: () => Promise<void> = async () => {};

  // Form state
  let trialOrg = "";
  let trialDays = 14;
  let repairOrg = "";
  let customerIdOrg = "";
  let customerId = "";
  let issuesOrg = "";

  const queryClient = useQueryClient();
  const extendTrial = createExtendTrialMutation();
  const billingRepair = createBillingRepairMutation();
  const deleteBillingIssue = createDeleteBillingIssueMutation();
  const setCustomer = createSetBillingCustomerMutation();

  $: billingIssuesQuery = getBillingIssues(issuesOrg);

  async function handleExtendTrial() {
    if (!trialOrg) return;
    try {
      await $extendTrial.mutateAsync({
        data: { org: trialOrg, days: trialDays },
      });
      bannerRef.show("success", `Trial extended by ${trialDays} days for ${trialOrg}`);
      trialOrg = "";
    } catch (err) {
      bannerRef.show("error", `Failed to extend trial: ${err}`);
    }
  }

  async function handleBillingRepair() {
    if (!repairOrg) return;
    confirmTitle = "Trigger Billing Repair";
    confirmDescription = `This will trigger a billing repair for organization "${repairOrg}". This recalculates billing state.`;
    confirmAction = async () => {
      try {
        await $billingRepair.mutateAsync({ data: { org: repairOrg } });
        bannerRef.show("success", `Billing repair triggered for ${repairOrg}`);
        repairOrg = "";
      } catch (err) {
        bannerRef.show("error", `Failed to trigger billing repair: ${err}`);
      }
    };
    confirmOpen = true;
  }

  async function handleSetCustomerId() {
    if (!customerIdOrg || !customerId) return;
    try {
      await $setCustomer.mutateAsync({
        data: { org: customerIdOrg, billingCustomerId: customerId },
      });
      bannerRef.show("success", `Billing customer ID set for ${customerIdOrg}`);
      customerIdOrg = "";
      customerId = "";
    } catch (err) {
      bannerRef.show("error", `Failed to set billing customer ID: ${err}`);
    }
  }

  async function handleDeleteIssue(org: string, type: string) {
    try {
      await $deleteBillingIssue.mutateAsync({ org, type });
      bannerRef.show("success", `Billing issue "${type}" deleted for ${org}`);
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed to delete billing issue: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Billing"
  description="Extend trials, repair billing state, manage billing customer IDs, and resolve billing issues."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <!-- Extend Trial -->
  <section class="card">
    <h2 class="card-title">Extend Trial</h2>
    <p class="card-desc">Add days to an organization's trial period.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={trialOrg}
      />
      <input
        type="number"
        class="input w-24"
        min="1"
        max="365"
        bind:value={trialDays}
      />
      <button class="btn-primary" on:click={handleExtendTrial}>
        Extend Trial
      </button>
    </div>
  </section>

  <!-- Set Billing Customer ID -->
  <section class="card">
    <h2 class="card-title">Set Billing Customer ID</h2>
    <p class="card-desc">Associate a Stripe customer ID with an organization.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={customerIdOrg}
      />
      <input
        type="text"
        class="input"
        placeholder="Stripe customer ID (cus_...)"
        bind:value={customerId}
      />
      <button class="btn-primary" on:click={handleSetCustomerId}>
        Set Customer ID
      </button>
    </div>
  </section>

  <!-- Billing Repair -->
  <section class="card">
    <h2 class="card-title">Billing Repair</h2>
    <p class="card-desc">Trigger a billing state recalculation for an organization.</p>
    <div class="form-row">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={repairOrg}
      />
      <button class="btn-primary" on:click={handleBillingRepair}>
        Trigger Repair
      </button>
    </div>
  </section>

  <!-- Billing Issues -->
  <section class="card">
    <h2 class="card-title">Billing Issues</h2>
    <p class="card-desc">View and resolve billing issues for an organization.</p>
    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={issuesOrg}
      />
    </div>
    {#if $billingIssuesQuery.data?.issues?.length}
      <div class="issues-list">
        {#each $billingIssuesQuery.data.issues as issue}
          <div class="issue-row">
            <div>
              <span class="issue-type">{issue.type}</span>
              <span class="issue-meta">{issue.metadata ?? ""}</span>
            </div>
            <button
              class="action-btn destructive"
              on:click={() => handleDeleteIssue(issuesOrg, issue.type ?? "")}
            >
              Delete Issue
            </button>
          </div>
        {/each}
      </div>
    {:else if issuesOrg && $billingIssuesQuery.isSuccess}
      <p class="text-sm text-slate-500">No billing issues found.</p>
    {/if}
  </section>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  .sections {
    @apply flex flex-col gap-6;
  }

  .card {
    @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700;
  }

  .card-title {
    @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-1;
  }

  .card-desc {
    @apply text-xs text-slate-500 dark:text-slate-400 mb-4;
  }

  .form-row {
    @apply flex gap-3 items-center flex-wrap;
  }

  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300
      dark:border-slate-600 bg-white dark:bg-slate-800
      text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 dark:placeholder:text-slate-500
      focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }

  .btn-primary {
    @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white
      hover:bg-blue-700 whitespace-nowrap;
  }

  .issues-list {
    @apply flex flex-col gap-2;
  }

  .issue-row {
    @apply flex items-center justify-between px-3 py-2 rounded
      bg-slate-50 dark:bg-slate-800;
  }

  .issue-type {
    @apply text-sm font-mono text-slate-700 dark:text-slate-300;
  }

  .issue-meta {
    @apply text-xs text-slate-500 ml-2;
  }

  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300;
  }

  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/billing/+page.svelte
git commit -m "feat(admin-console): add billing management page with trial extension, repair, and issue management"
```

---

### Task 11: Navigation Link from Main App

**Files:**
- Modify: `web-admin/src/features/organizations/OrgHeader.svelte` (or equivalent top nav)

This task adds a small "Admin" link in the top navigation bar that is only visible to superusers, ensuring the admin console is discoverable from day one.

- [ ] **Step 1: Add superuser check and admin link**

In the `OrgHeader.svelte` component (or the user menu dropdown), add a conditional admin console link. The superuser check uses the same `ListSuperusers` query pattern from the admin layout guard:

```svelte
<script lang="ts">
  import {
    createAdminServiceListSuperusers,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";

  const currentUser = createAdminServiceGetCurrentUser();
  const superusersQuery = createAdminServiceListSuperusers();

  $: isSuperuser = $superusersQuery.data?.users?.some(
    (u) => u.email === $currentUser.data?.user?.email,
  ) ?? false;
</script>

<!-- Add near the user menu / top-right nav area -->
{#if isSuperuser}
  <a
    href="/-/admin"
    class="text-xs font-medium text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200"
  >
    Admin
  </a>
{/if}
```

Note: The `ListSuperusers` call will 403 for non-superusers. Wrap in a try/catch or use the `query.error` state to silently hide the link for non-superusers.

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/organizations/OrgHeader.svelte
git commit -m "feat(admin-console): add admin link in top nav for superusers"
```

---

## Phase 2: Organizations + Quotas

### Task 12: Organization Management — Selectors

**Files:**
- Create: `web-admin/src/features/admin/organizations/selectors.ts`

- [ ] **Step 1: Create org selectors**

```typescript
// web-admin/src/features/admin/organizations/selectors.ts
import {
  createAdminServiceGetOrganization,
  createAdminServiceSudoUpdateOrganizationCustomDomain,
  createAdminServiceListOrganizationMemberUsers,
  createAdminServiceAddOrganizationMemberUser,
} from "@rilldata/web-admin/client";

export function getOrganization(org: string) {
  return createAdminServiceGetOrganization(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function getOrgAdmins(org: string) {
  return createAdminServiceListOrganizationMemberUsers(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createSetCustomDomainMutation() {
  return createAdminServiceSudoUpdateOrganizationCustomDomain();
}

export function createJoinOrgMutation() {
  return createAdminServiceAddOrganizationMemberUser();
}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/organizations/selectors.ts
git commit -m "feat(admin-console): add organization management selectors"
```

---

### Task 13: Organization Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/organizations/+page.svelte`

- [ ] **Step 1: Create organizations page with lookup, details, custom domain, and join**

```svelte
<!-- web-admin/src/routes/-/admin/organizations/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import {
    getOrganization,
    getOrgAdmins,
    createSetCustomDomainMutation,
    createJoinOrgMutation,
  } from "@rilldata/web-admin/features/admin/organizations/selectors";

  let bannerRef: ActionResultBanner;
  let orgName = "";
  let lookupOrg = "";
  let customDomainOrg = "";
  let customDomain = "";
  let joinOrg = "";
  let joinEmail = "";
  let joinRole = "admin";

  const setCustomDomain = createSetCustomDomainMutation();
  const joinOrgMutation = createJoinOrgMutation();

  $: orgQuery = getOrganization(lookupOrg);
  $: adminsQuery = getOrgAdmins(lookupOrg);

  function handleLookup() {
    lookupOrg = orgName;
  }

  async function handleSetCustomDomain() {
    if (!customDomainOrg || !customDomain) return;
    try {
      await $setCustomDomain.mutateAsync({
        data: { name: customDomainOrg, customDomain },
      });
      bannerRef.show("success", `Custom domain set for ${customDomainOrg}`);
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  async function handleJoinOrg() {
    if (!joinOrg || !joinEmail) return;
    try {
      await $joinOrgMutation.mutateAsync({
        org: joinOrg,
        data: { email: joinEmail, role: joinRole, superuserForceAccess: true },
      });
      bannerRef.show("success", `${joinEmail} added to ${joinOrg} as ${joinRole}`);
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Organizations"
  description="Lookup organizations, view their details, set custom domains, and add users."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <!-- Org Lookup -->
  <section class="card">
    <h2 class="card-title">Organization Lookup</h2>
    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder="Organization name"
        bind:value={orgName}
        on:keydown={(e) => e.key === "Enter" && handleLookup()}
      />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if $orgQuery.data?.organization}
      {@const org = $orgQuery.data.organization}
      <div class="detail-grid">
        <div class="detail-item">
          <span class="detail-label">ID</span>
          <span class="detail-value font-mono">{org.id}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Name</span>
          <span class="detail-value">{org.name}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Description</span>
          <span class="detail-value">{org.description ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Billing Plan</span>
          <span class="detail-value">{org.billingPlanDisplayName ?? "-"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Custom Domain</span>
          <span class="detail-value">{org.customDomain ?? "None"}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">Created</span>
          <span class="detail-value">
            {org.createdOn ? new Date(org.createdOn).toLocaleDateString() : "-"}
          </span>
        </div>
      </div>

      {#if $adminsQuery.data?.members?.length}
        <h3 class="mt-4 text-xs font-semibold text-slate-500 uppercase">Members</h3>
        <div class="mt-2">
          {#each $adminsQuery.data.members as member}
            <div class="member-row">
              <span class="text-sm">{member.userEmail}</span>
              <span class="text-xs text-slate-500">{member.roleName}</span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </section>

  <!-- Set Custom Domain -->
  <section class="card">
    <h2 class="card-title">Set Custom Domain</h2>
    <div class="form-row">
      <input type="text" class="input" placeholder="Organization name" bind:value={customDomainOrg} />
      <input type="text" class="input" placeholder="Custom domain (e.g. analytics.acme.com)" bind:value={customDomain} />
      <button class="btn-primary" on:click={handleSetCustomDomain}>Set Domain</button>
    </div>
  </section>

  <!-- Join Organization -->
  <section class="card">
    <h2 class="card-title">Add User to Organization</h2>
    <div class="form-row">
      <input type="text" class="input" placeholder="Organization name" bind:value={joinOrg} />
      <input type="email" class="input" placeholder="User email" bind:value={joinEmail} />
      <select class="input" bind:value={joinRole}>
        <option value="admin">Admin</option>
        <option value="editor">Editor</option>
        <option value="viewer">Viewer</option>
      </select>
      <button class="btn-primary" on:click={handleJoinOrg}>Add User</button>
    </div>
  </section>
</div>

<style lang="postcss">
  .sections { @apply flex flex-col gap-6; }
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .card-title { @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  .detail-grid { @apply grid grid-cols-2 lg:grid-cols-3 gap-3; }
  .detail-item { @apply flex flex-col; }
  .detail-label { @apply text-[11px] text-slate-500 dark:text-slate-400 uppercase tracking-wider; }
  .detail-value { @apply text-sm text-slate-900 dark:text-slate-100; }
  .member-row { @apply flex justify-between items-center px-3 py-1.5 rounded bg-slate-50 dark:bg-slate-800 mb-1; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/organizations/+page.svelte
git commit -m "feat(admin-console): add organization management page"
```

---

### Task 14: Quota Management — Selectors

**Files:**
- Create: `web-admin/src/features/admin/quotas/selectors.ts`

- [ ] **Step 1: Create quota selectors**

```typescript
// web-admin/src/features/admin/quotas/selectors.ts
import {
  createAdminServiceGetOrganization,
  createAdminServiceSudoUpdateOrganizationQuotas,
  createAdminServiceSudoUpdateUserQuotas,
} from "@rilldata/web-admin/client";

export function getOrgForQuotas(org: string) {
  return createAdminServiceGetOrganization(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createUpdateOrgQuotasMutation() {
  return createAdminServiceSudoUpdateOrganizationQuotas();
}

export function createUpdateUserQuotasMutation() {
  return createAdminServiceSudoUpdateUserQuotas();
}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/quotas/selectors.ts
git commit -m "feat(admin-console): add quota management selectors"
```

---

### Task 15: Quota Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/quotas/+page.svelte`

- [ ] **Step 1: Create quotas page with org/user lookup and editable fields**

```svelte
<!-- web-admin/src/routes/-/admin/quotas/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import {
    getOrgForQuotas,
    createUpdateOrgQuotasMutation,
    createUpdateUserQuotasMutation,
  } from "@rilldata/web-admin/features/admin/quotas/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let quotaType: "org" | "user" = "org";
  let lookupValue = "";
  let activeOrg = "";
  let activeUser = "";
  let lookupDone = false;

  const queryClient = useQueryClient();
  const updateOrgQuotas = createUpdateOrgQuotasMutation();
  const updateUserQuotas = createUpdateUserQuotasMutation();

  // Quota fields for editing (org quotas)
  let projects = "";
  let deployments = "";
  let slotsTotal = "";
  let slotsPerDeployment = "";
  let outstandingInvites = "";
  let storageLimitBytes = "";

  // Quota fields for editing (user quotas)
  let singleuserOrgs = "";

  $: orgQuery = getOrgForQuotas(activeOrg);

  function handleLookup() {
    lookupDone = true;
    if (quotaType === "org") {
      activeOrg = lookupValue;
      activeUser = "";
    } else {
      activeUser = lookupValue;
      activeOrg = "";
    }
  }

  // Populate fields when org data loads
  $: if ($orgQuery.data?.organization?.quotas) {
    const q = $orgQuery.data.organization.quotas;
    projects = q.projects ?? "";
    deployments = q.deployments ?? "";
    slotsTotal = q.slotsTotal ?? "";
    slotsPerDeployment = q.slotsPerDeployment ?? "";
    outstandingInvites = q.outstandingInvites ?? "";
    storageLimitBytes = q.storageLimitBytesPerDeployment ?? "";
  }

  async function handleSaveQuotas() {
    try {
      if (quotaType === "org") {
        await $updateOrgQuotas.mutateAsync({
          data: {
            org: activeOrg,
            projects: projects ? Number(projects) : undefined,
            deployments: deployments ? Number(deployments) : undefined,
            slotsTotal: slotsTotal ? Number(slotsTotal) : undefined,
            slotsPerDeployment: slotsPerDeployment ? Number(slotsPerDeployment) : undefined,
            outstandingInvites: outstandingInvites ? Number(outstandingInvites) : undefined,
            storageLimitBytesPerDeployment: storageLimitBytes ? storageLimitBytes : undefined,
          },
        });
        bannerRef.show("success", `Quotas updated for org: ${activeOrg}`);
      } else {
        await $updateUserQuotas.mutateAsync({
          data: {
            email: activeUser,
            singleuserOrgs: singleuserOrgs ? Number(singleuserOrgs) : undefined,
          },
        });
        bannerRef.show("success", `Quotas updated for user: ${activeUser}`);
      }
      await queryClient.invalidateQueries({
        predicate: (q) =>
          (q.queryKey[0] as string)?.includes("/v1/superuser/quotas") ||
          (q.queryKey[0] as string)?.includes("/v1/organizations"),
      });
    } catch (err) {
      bannerRef.show("error", `Failed to update quotas: ${err}`);
    }
  }
</script>

<AdminPageHeader
  title="Quotas"
  description="View and adjust resource quotas for organizations and users."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <section class="card">
    <div class="flex gap-4 mb-4">
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" value="org" bind:group={quotaType} />
        Organization
      </label>
      <label class="flex items-center gap-2 text-sm">
        <input type="radio" value="user" bind:group={quotaType} />
        User
      </label>
    </div>

    <div class="form-row mb-4">
      <input
        type="text"
        class="input"
        placeholder={quotaType === "org" ? "Organization name" : "User email"}
        bind:value={lookupValue}
        on:keydown={(e) => e.key === "Enter" && handleLookup()}
      />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if quotaType === "org" && activeOrg && $orgQuery.data?.organization}
      <div class="quota-grid">
        <div class="quota-field">
          <label class="quota-label" for="projects">Projects</label>
          <input id="projects" type="number" class="input" bind:value={projects} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="deployments">Deployments</label>
          <input id="deployments" type="number" class="input" bind:value={deployments} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="slotsTotal">Total Slots</label>
          <input id="slotsTotal" type="number" class="input" bind:value={slotsTotal} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="slotsPerDeployment">Slots per Deployment</label>
          <input id="slotsPerDeployment" type="number" class="input" bind:value={slotsPerDeployment} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="outstandingInvites">Outstanding Invites</label>
          <input id="outstandingInvites" type="number" class="input" bind:value={outstandingInvites} />
        </div>
        <div class="quota-field">
          <label class="quota-label" for="storageLimitBytes">Storage Limit (bytes)</label>
          <input id="storageLimitBytes" type="text" class="input" bind:value={storageLimitBytes} />
        </div>
      </div>

      <div class="mt-4">
        <button class="btn-primary" on:click={handleSaveQuotas}>Save Quotas</button>
      </div>
    {:else if quotaType === "user" && activeUser && lookupDone}
      <div class="quota-grid">
        <div class="quota-field">
          <label class="quota-label" for="singleuserOrgs">Single-user Orgs Limit</label>
          <input id="singleuserOrgs" type="number" class="input" bind:value={singleuserOrgs} />
        </div>
      </div>
      <p class="text-xs text-slate-500 mt-2">User quotas are limited to the single-user orgs field. Other quotas are managed at the org level.</p>

      <div class="mt-4">
        <button class="btn-primary" on:click={handleSaveQuotas}>Save Quotas</button>
      </div>
    {/if}
  </section>
</div>

<style lang="postcss">
  .sections { @apply flex flex-col gap-6; }
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  .quota-grid { @apply grid grid-cols-2 lg:grid-cols-3 gap-4; }
  .quota-field { @apply flex flex-col gap-1; }
  .quota-label { @apply text-xs font-medium text-slate-500 dark:text-slate-400; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/quotas/+page.svelte
git commit -m "feat(admin-console): add quota management page with editable fields"
```

---

## Phase 3: Projects + Whitelist

### Task 16: Project Management — Selectors

**Files:**
- Create: `web-admin/src/features/admin/projects/selectors.ts`

- [ ] **Step 1: Create project selectors**

```typescript
// web-admin/src/features/admin/projects/selectors.ts
import {
  createAdminServiceSearchProjectNames,
  createAdminServiceGetProject,
  createAdminServiceUpdateProject,
  createAdminServiceRedeployProject,
  createAdminServiceHibernateProject,
} from "@rilldata/web-admin/client";

export function searchProjects(namePattern: string) {
  return createAdminServiceSearchProjectNames(
    { namePattern, pageSize: 50 },
    { query: { enabled: namePattern.length >= 2 } },
  );
}

export function getProject(org: string, project: string) {
  return createAdminServiceGetProject(org, project);
}

export function createUpdateProjectMutation() {
  return createAdminServiceUpdateProject();
}

export function createRedeployProjectMutation() {
  return createAdminServiceRedeployProject();
}

export function createHibernateProjectMutation() {
  return createAdminServiceHibernateProject();
}
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/features/admin/projects/selectors.ts
git commit -m "feat(admin-console): add project management selectors"
```

---

### Task 17: Project Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/projects/+page.svelte`

- [ ] **Step 1: Create projects page with search, edit, hibernate, and reset**

This page follows the same pattern as the users page: search bar, results table, action buttons per row. Key actions:
- **Edit**: opens inline inputs for `prodSlots` and `prodVersion`
- **Hibernate**: confirmation dialog, calls `HibernateProject`
- **Reset (Redeploy)**: destructive confirmation, calls `RedeployProject`

```svelte
<!-- web-admin/src/routes/-/admin/projects/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import SearchInput from "@rilldata/web-admin/features/admin/shared/SearchInput.svelte";
  import {
    searchProjects,
    createRedeployProjectMutation,
    createHibernateProjectMutation,
  } from "@rilldata/web-admin/features/admin/projects/selectors";

  let bannerRef: ActionResultBanner;
  let searchQuery = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

  const redeployProject = createRedeployProjectMutation();
  const hibernateProject = createHibernateProjectMutation();

  $: projectsQuery = searchProjects(searchQuery);

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
  }

  function handleHibernate(name: string) {
    const [org, project] = name.split("/");
    confirmTitle = "Hibernate Project";
    confirmDescription = `This will hibernate the deployment for ${name}. The project data will be preserved but the deployment will be stopped.`;
    confirmDestructive = false;
    confirmAction = async () => {
      try {
        await $hibernateProject.mutateAsync({ organization: org, project });
        bannerRef.show("success", `Project ${name} hibernated`);
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }

  function handleRedeploy(name: string) {
    const [org, project] = name.split("/");
    confirmTitle = "Redeploy Project";
    confirmDescription = `This will completely redeploy ${name}. This is a disruptive operation.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $redeployProject.mutateAsync({ organization: org, project });
        bannerRef.show("success", `Project ${name} redeployed`);
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Projects"
  description="Search projects by name pattern, view details, hibernate or redeploy."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="mb-4 max-w-md">
  <SearchInput
    placeholder="Search projects (e.g. org/project, min 2 chars)..."
    on:search={handleSearch}
  />
</div>

{#if $projectsQuery.isLoading && searchQuery.length >= 2}
  <p class="text-sm text-slate-500">Searching...</p>
{:else if $projectsQuery.data?.names?.length}
  <table class="w-full">
    <thead>
      <tr>
        <th>Project</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody>
      {#each $projectsQuery.data.names as name}
        <tr>
          <td class="font-mono text-xs">{name}</td>
          <td>
            <div class="flex gap-2">
              <a
                href={`/${name}`}
                target="_blank"
                class="action-btn"
              >
                View
              </a>
              <button class="action-btn" on:click={() => handleHibernate(name)}>
                Hibernate
              </button>
              <button
                class="action-btn destructive"
                on:click={() => handleRedeploy(name)}
              >
                Redeploy
              </button>
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{:else if searchQuery.length >= 2 && $projectsQuery.isSuccess}
  <p class="text-sm text-slate-500">No projects found for "{searchQuery}"</p>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirmTitle}
  description={confirmDescription}
  destructive={confirmDestructive}
  onConfirm={confirmAction}
/>

<style lang="postcss">
  th {
    @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider
      px-4 py-2 border-b border-slate-200 dark:border-slate-700;
  }
  td {
    @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300
      border-b border-slate-100 dark:border-slate-800;
  }
  tr:hover td { @apply bg-slate-50 dark:bg-slate-800/50; }
  .action-btn {
    @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600
      text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700;
  }
  .action-btn.destructive {
    @apply border-red-300 text-red-600 hover:bg-red-50
      dark:border-red-700 dark:text-red-400 dark:hover:bg-red-900/20;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/projects/+page.svelte
git commit -m "feat(admin-console): add project management page with search, hibernate, and redeploy"
```

---

### Task 18: Whitelist Management — Selectors + Page

**Files:**
- Create: `web-admin/src/features/admin/whitelist/selectors.ts`
- Create: `web-admin/src/routes/-/admin/whitelist/+page.svelte`

- [ ] **Step 1: Create whitelist selectors**

```typescript
// web-admin/src/features/admin/whitelist/selectors.ts
import {
  createAdminServiceCreateWhitelistedDomain,
  createAdminServiceRemoveWhitelistedDomain,
  createAdminServiceListWhitelistedDomains,
} from "@rilldata/web-admin/client";

export function getWhitelistedDomains(org: string) {
  return createAdminServiceListWhitelistedDomains(
    org,
    { superuserForceAccess: true },
    { query: { enabled: org.length > 0 } },
  );
}

export function createAddWhitelistMutation() {
  return createAdminServiceCreateWhitelistedDomain();
}

export function createRemoveWhitelistMutation() {
  return createAdminServiceRemoveWhitelistedDomain();
}
```

- [ ] **Step 2: Create whitelist page**

```svelte
<!-- web-admin/src/routes/-/admin/whitelist/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import {
    getWhitelistedDomains,
    createAddWhitelistMutation,
    createRemoveWhitelistMutation,
  } from "@rilldata/web-admin/features/admin/whitelist/selectors";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let org = "";
  let activeOrg = "";
  let newDomain = "";
  let newRole = "viewer";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const addWhitelist = createAddWhitelistMutation();
  const removeWhitelist = createRemoveWhitelistMutation();

  $: domainsQuery = getWhitelistedDomains(activeOrg);

  function handleLookup() {
    activeOrg = org;
  }

  async function handleAdd() {
    if (!activeOrg || !newDomain) return;
    try {
      await $addWhitelist.mutateAsync({
        org: activeOrg,
        data: { domain: newDomain, role: newRole },
      });
      bannerRef.show("success", `Domain ${newDomain} whitelisted for ${activeOrg}`);
      newDomain = "";
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  function handleRemove(domain: string) {
    confirmTitle = "Remove Whitelisted Domain";
    confirmDescription = `Remove "${domain}" from the whitelist for ${activeOrg}?`;
    confirmAction = async () => {
      try {
        await $removeWhitelist.mutateAsync({ org: activeOrg, domain });
        bannerRef.show("success", `Domain ${domain} removed from whitelist`);
        await queryClient.invalidateQueries();
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Domain Whitelist"
  description="Manage whitelisted email domains for organizations."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="sections">
  <section class="card">
    <div class="form-row mb-4">
      <input type="text" class="input" placeholder="Organization name" bind:value={org}
        on:keydown={(e) => e.key === "Enter" && handleLookup()} />
      <button class="btn-primary" on:click={handleLookup}>Lookup</button>
    </div>

    {#if activeOrg}
      <div class="form-row mb-4">
        <input type="text" class="input" placeholder="Domain (e.g. acme.com)" bind:value={newDomain} />
        <select class="input" bind:value={newRole}>
          <option value="admin">Admin</option>
          <option value="editor">Editor</option>
          <option value="viewer">Viewer</option>
        </select>
        <button class="btn-primary" on:click={handleAdd}>Add Domain</button>
      </div>

      {#if $domainsQuery.data?.domains?.length}
        <table class="w-full">
          <thead>
            <tr><th>Domain</th><th>Role</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {#each $domainsQuery.data.domains as d}
              <tr>
                <td class="font-mono text-xs">{d.domain}</td>
                <td class="text-xs">{d.role}</td>
                <td>
                  <button class="action-btn destructive" on:click={() => handleRemove(d.domain ?? "")}>
                    Remove
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else if $domainsQuery.isSuccess}
        <p class="text-sm text-slate-500">No whitelisted domains.</p>
      {/if}
    {/if}
  </section>
</div>

<ConfirmDialog bind:open={confirmOpen} title={confirmTitle} description={confirmDescription} onConfirm={confirmAction} />

<style lang="postcss">
  .sections { @apply flex flex-col gap-6; }
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  th { @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700; }
  td { @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800; }
  .action-btn { @apply text-xs px-2 py-1 rounded border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300; }
  .action-btn.destructive { @apply border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400; }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add web-admin/src/features/admin/whitelist/ web-admin/src/routes/-/admin/whitelist/
git commit -m "feat(admin-console): add domain whitelist management page"
```

---

## Phase 4: Superuser Management + Annotations

### Task 19: Superuser Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/superusers/+page.svelte`

- [ ] **Step 1: Create superusers page with list + add/remove**

```svelte
<!-- web-admin/src/routes/-/admin/superusers/+page.svelte -->
<script lang="ts">
  import {
    createAdminServiceListSuperusers,
    createAdminServiceSetSuperuser,
  } from "@rilldata/web-admin/client";
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";
  import ConfirmDialog from "@rilldata/web-admin/features/admin/shared/ConfirmDialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  let bannerRef: ActionResultBanner;
  let newEmail = "";
  let confirmOpen = false;
  let confirmTitle = "";
  let confirmDescription = "";
  let confirmDestructive = false;
  let confirmAction: () => Promise<void> = async () => {};

  const queryClient = useQueryClient();
  const superusersQuery = createAdminServiceListSuperusers();
  const setSuperuser = createAdminServiceSetSuperuser();

  async function handleAdd() {
    if (!newEmail) return;
    try {
      await $setSuperuser.mutateAsync({ data: { email: newEmail, superuser: true } });
      bannerRef.show("success", `${newEmail} added as superuser`);
      newEmail = "";
      await queryClient.invalidateQueries();
    } catch (err) {
      bannerRef.show("error", `Failed: ${err}`);
    }
  }

  function handleRemove(email: string) {
    confirmTitle = "Remove Superuser";
    confirmDescription = `Remove superuser access for ${email}? They will lose access to this admin console.`;
    confirmDestructive = true;
    confirmAction = async () => {
      try {
        await $setSuperuser.mutateAsync({ data: { email, superuser: false } });
        bannerRef.show("success", `${email} removed as superuser`);
        await queryClient.invalidateQueries();
      } catch (err) {
        bannerRef.show("error", `Failed: ${err}`);
      }
    };
    confirmOpen = true;
  }
</script>

<AdminPageHeader
  title="Superusers"
  description="Manage who has superuser (super admin) access to Rill Cloud."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card mb-6">
  <h2 class="card-title">Add Superuser</h2>
  <div class="form-row">
    <input type="email" class="input" placeholder="Email address" bind:value={newEmail}
      on:keydown={(e) => e.key === "Enter" && handleAdd()} />
    <button class="btn-primary" on:click={handleAdd}>Add Superuser</button>
  </div>
</div>

{#if $superusersQuery.data?.users?.length}
  <table class="w-full">
    <thead>
      <tr><th>Email</th><th>Display Name</th><th>Actions</th></tr>
    </thead>
    <tbody>
      {#each $superusersQuery.data.users as user}
        <tr>
          <td class="font-mono text-xs">{user.email}</td>
          <td>{user.displayName ?? "-"}</td>
          <td>
            <button class="action-btn destructive" on:click={() => handleRemove(user.email ?? "")}>
              Remove
            </button>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
{/if}

<ConfirmDialog bind:open={confirmOpen} title={confirmTitle} description={confirmDescription}
  destructive={confirmDestructive} onConfirm={confirmAction} />

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .card-title { @apply text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 whitespace-nowrap; }
  th { @apply text-left text-xs font-medium text-slate-500 uppercase tracking-wider px-4 py-2 border-b border-slate-200 dark:border-slate-700; }
  td { @apply px-4 py-3 text-sm text-slate-700 dark:text-slate-300 border-b border-slate-100 dark:border-slate-800; }
  .action-btn.destructive { @apply text-xs px-2 py-1 rounded border border-red-300 text-red-600 hover:bg-red-50 dark:border-red-700 dark:text-red-400; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/superusers/+page.svelte
git commit -m "feat(admin-console): add superuser management page"
```

---

### Task 20: Annotations Management — Page

**Files:**
- Create: `web-admin/src/routes/-/admin/annotations/+page.svelte`

- [ ] **Step 1: Create annotations page**

```svelte
<!-- web-admin/src/routes/-/admin/annotations/+page.svelte -->
<script lang="ts">
  import {
    createAdminServiceSudoUpdateAnnotations,
    createAdminServiceSudoGetResource,
  } from "@rilldata/web-admin/client";
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;
  let org = "";
  let project = "";
  let annotationsJson = "{}";
  let loaded = false;

  const updateAnnotations = createAdminServiceSudoUpdateAnnotations();

  async function handleLoad() {
    if (!org || !project) return;
    try {
      // Load existing project annotations via GetProject
      // The project's annotations field contains the current values
      const resp = await fetch(`/v1/organizations/${org}/projects/${project}`);
      if (resp.ok) {
        const data = await resp.json();
        const existing = data.project?.annotations ?? {};
        annotationsJson = JSON.stringify(existing, null, 2);
        loaded = true;
        bannerRef.show("success", `Loaded annotations for ${org}/${project}`);
      } else {
        bannerRef.show("error", `Project not found or access denied`);
      }
    } catch (err) {
      bannerRef.show("error", `Failed to load annotations: ${err}`);
    }
  }

  async function handleSave() {
    if (!org || !project) return;
    try {
      const annotations = JSON.parse(annotationsJson);
      await $updateAnnotations.mutateAsync({
        data: { organization: org, project, annotations },
      });
      bannerRef.show("success", `Annotations updated for ${org}/${project}`);
    } catch (err) {
      if (err instanceof SyntaxError) {
        bannerRef.show("error", "Invalid JSON format");
      } else {
        bannerRef.show("error", `Failed: ${err}`);
      }
    }
  }
</script>

<AdminPageHeader
  title="Annotations"
  description="View and update project annotations (key-value metadata used for billing, categorization, etc.)."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <div class="form-row mb-4">
    <input type="text" class="input" placeholder="Organization name" bind:value={org} />
    <input type="text" class="input" placeholder="Project name" bind:value={project} />
    <button class="btn-primary" on:click={handleLoad}>Load Current</button>
  </div>
  <div class="mb-4">
    <label class="text-xs font-medium text-slate-500 mb-1 block">
      Annotations (JSON) {#if !loaded}<span class="text-yellow-600">— click "Load Current" first to avoid overwriting</span>{/if}
    </label>
    <textarea
      class="input w-full h-32 font-mono text-xs"
      placeholder='{"key": "value"}'
      bind:value={annotationsJson}
    ></textarea>
  </div>
  <button class="btn-primary" on:click={handleSave}>Save Annotations</button>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700; }
  textarea { @apply resize-y; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-admin/src/routes/-/admin/annotations/+page.svelte
git commit -m "feat(admin-console): add annotations management page"
```

---

## Phase 5: Virtual Files + Runtime

### Task 21: Virtual Files + Runtime — Stub Pages

**Files:**
- Create: `web-admin/src/routes/-/admin/virtual-files/+page.svelte`
- Create: `web-admin/src/routes/-/admin/runtime/+page.svelte`

These are more technical/engineer-leaning and can start as functional stubs.

- [ ] **Step 1: Create virtual-files page**

```svelte
<!-- web-admin/src/routes/-/admin/virtual-files/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;
  let org = "";
  let project = "";

  // Virtual files management will be implemented in a follow-up.
  // The API calls needed are:
  // - PullVirtualRepo (list)
  // - GetVirtualFile (read)
  // - DeleteVirtualFile (delete)
</script>

<AdminPageHeader
  title="Virtual Files"
  description="Browse, read, and delete virtual files in project deployments."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <div class="form-row mb-4">
    <input type="text" class="input" placeholder="Organization name" bind:value={org} />
    <input type="text" class="input" placeholder="Project name" bind:value={project} />
    <button class="btn-primary" disabled>List Files (Coming Soon)</button>
  </div>
  <p class="text-sm text-slate-500">Virtual file management will be available in a future update.</p>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
  .form-row { @apply flex gap-3 items-center flex-wrap; }
  .input {
    @apply px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500;
  }
  .btn-primary { @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50; }
</style>
```

- [ ] **Step 2: Create runtime page**

```svelte
<!-- web-admin/src/routes/-/admin/runtime/+page.svelte -->
<script lang="ts">
  import AdminPageHeader from "@rilldata/web-admin/features/admin/layout/AdminPageHeader.svelte";
  import ActionResultBanner from "@rilldata/web-admin/features/admin/shared/ActionResultBanner.svelte";

  let bannerRef: ActionResultBanner;

  // Runtime management will be implemented in a follow-up.
  // The API calls needed are:
  // - SudoIssueRuntimeManagerToken
  // - ListInstances (via runtime client)
  // - DeleteInstance (via runtime client)
</script>

<AdminPageHeader
  title="Runtime"
  description="Manage runtime instances, issue manager tokens, and view deployment infrastructure."
/>

<ActionResultBanner bind:this={bannerRef} />

<div class="card">
  <p class="text-sm text-slate-500">Runtime management will be available in a future update. Use the CLI for now:</p>
  <pre class="mt-2 p-3 bg-slate-100 dark:bg-slate-800 rounded text-xs font-mono">rill sudo runtime list-instances &lt;host&gt;
rill sudo runtime delete-instance &lt;host&gt; &lt;instance_id&gt;
rill sudo runtime manager-token &lt;host&gt;</pre>
</div>

<style lang="postcss">
  .card { @apply p-5 rounded-lg border border-slate-200 dark:border-slate-700; }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add web-admin/src/routes/-/admin/virtual-files/ web-admin/src/routes/-/admin/runtime/
git commit -m "feat(admin-console): add stub pages for virtual files and runtime management"
```

---

---

## Summary

| Phase | Tasks | What Ships |
|-------|-------|------------|
| 1 | Tasks 1-11 | Layout shell, auth guard, sidebar, dashboard home, Users page, Billing page, Nav link |
| 2 | Tasks 12-15 | Organizations page, Quotas page |
| 3 | Tasks 16-18 | Projects page, Whitelist page |
| 4 | Tasks 19-20 | Superusers page, Annotations page |
| 5 | Task 21 | Virtual Files stub, Runtime stub |

**Total: 21 tasks, ~50 new files, 0 generated API changes needed (all endpoints already exist)**

### Command Coverage

| CLI Command Group | Admin Console Status |
|-------------------|---------------------|
| `sudo user` (6 cmds) | Phase 1: Full coverage (search, assume, unassume, open, delete, list) |
| `sudo billing` (5 cmds) | Phase 1: Full coverage (extend trial, set customer, delete issue, repair, setup) |
| `sudo org` (5 cmds) | Phase 2: Full coverage (show, join, list-admins, set-custom-domain, set-internal-plan) |
| `sudo quota` (2 cmds) | Phase 2: Full coverage (get, set for both org and user) |
| `sudo project` (6 cmds) | Phase 3: Core coverage (search, hibernate, reset); edit and dump-resources follow-up |
| `sudo whitelist` (2 cmds) | Phase 3: Full coverage (add, remove) |
| `sudo superuser` (3 cmds) | Phase 4: Full coverage (list, add, remove) |
| `sudo annotations` (2 cmds) | Phase 4: Full coverage (get, set) |
| `sudo virtual-files` (3 cmds) | Phase 5: Stub (list, get, delete to be implemented) |
| `sudo runtime` (3 cmds) | Phase 5: Stub (manager-token, list-instances, delete-instance) |
| `sudo lookup` (1 cmd) | Future: Resource lookup by ID |
| `sudo clone` (1 cmd) | Future: Clone operations |
