<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { notifications } from "@rilldata/web-common/lib/notifications";
  import { Button } from "@rilldata/web-common/components/button";
  import { Checkbox } from "@rilldata/web-common/components/checkbox";
  import {
    createServiceTokenListQuery,
    deleteServiceTokenMutation,
  } from "./token-queries";
  import DeleteTokenDialog from "./DeleteTokenDialog.svelte";
  import TokenDetailsDrawer from "./TokenDetailsDrawer.svelte";

  export let organization: string;

  const dispatch = createEventDispatcher();

  let searchText = "";
  let debouncedSearch = "";
  let debounceTimer: ReturnType<typeof setTimeout>;
  let pageToken: string | undefined = undefined;

  // Drawer state
  let drawerOpen = false;
  let selectedTokenId = "";

  // Delete dialog state
  let deleteDialogOpen = false;
  let tokenToDelete: { id: string; name: string } | null = null;

  // Bulk selection state
  let selectedIds: Set<string> = new Set();
  let bulkDeleteDialogOpen = false;
  let bulkDeleting = false;

  function handleSearchInput(e: Event) {
    const value = (e.target as HTMLInputElement).value;
    searchText = value;
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debouncedSearch = value;
      pageToken = undefined;
      selectedIds = new Set();
    }, 300);
  }

  $: tokensQuery = createServiceTokenListQuery(organization, {
    search: debouncedSearch || undefined,
    pageToken,
  });

  $: tokens = $tokensQuery.data?.tokens ?? [];
  $: nextPageToken = $tokensQuery.data?.nextPageToken;
  $: isLoading = $tokensQuery.isLoading;
  $: isError = $tokensQuery.isError;

  // Checkbox helpers
  $: allOnPageSelected =
    tokens.length > 0 && tokens.every((t) => selectedIds.has(t.id));
  $: someOnPageSelected =
    tokens.some((t) => selectedIds.has(t.id)) && !allOnPageSelected;
  $: selectionCount = selectedIds.size;

  function toggleSelectAll() {
    if (allOnPageSelected) {
      // Deselect all on current page
      const updated = new Set(selectedIds);
      for (const t of tokens) {
        updated.delete(t.id);
      }
      selectedIds = updated;
    } else {
      // Select all on current page
      const updated = new Set(selectedIds);
      for (const t of tokens) {
        updated.add(t.id);
      }
      selectedIds = updated;
    }
  }

  function toggleSelectRow(id: string) {
    const updated = new Set(selectedIds);
    if (updated.has(id)) {
      updated.delete(id);
    } else {
      updated.add(id);
    }
    selectedIds = updated;
  }

  function handleRowClick(token: { id: string }) {
    selectedTokenId = token.id;
    drawerOpen = true;
  }

  function openViewDetails(token: { id: string }) {
    selectedTokenId = token.id;
    drawerOpen = true;
  }

  function openDeleteDialog(token: { id: string; name: string }) {
    tokenToDelete = token;
    deleteDialogOpen = true;
  }

  function closeDeleteDialog() {
    deleteDialogOpen = false;
    tokenToDelete = null;
  }

  function closeDrawer() {
    drawerOpen = false;
    selectedTokenId = "";
  }

  function loadNextPage() {
    if (nextPageToken) {
      pageToken = nextPageToken;
      selectedIds = new Set();
    }
  }

  // Delete mutation
  const deleteMutation = deleteServiceTokenMutation(organization);

  async function handleDeleteConfirm() {
    if (!tokenToDelete) return;
    await $deleteMutation.mutateAsync({ tokenId: tokenToDelete.id });
    notifications.send({
      message: `Token "${tokenToDelete.name}" has been revoked.`,
    });
    closeDeleteDialog();
  }

  // Bulk delete
  function openBulkDeleteDialog() {
    if (selectionCount === 0) return;
    bulkDeleteDialogOpen = true;
  }

  function closeBulkDeleteDialog() {
    bulkDeleteDialogOpen = false;
  }

  async function handleBulkDeleteConfirm() {
    bulkDeleting = true;
    const ids = Array.from(selectedIds);
    const results = await Promise.allSettled(
      ids.map((tokenId) => $deleteMutation.mutateAsync({ tokenId })),
    );

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    const failed = results.filter((r) => r.status === "rejected").length;

    if (failed === 0) {
      notifications.send({
        message: `Successfully revoked ${succeeded} token${succeeded !== 1 ? "s" : ""}.`,
      });
    } else {
      notifications.send({
        message: `Revoked ${succeeded} token${succeeded !== 1 ? "s" : ""}, but ${failed} failed. Please try again.`,
        type: "error",
      });
    }

    selectedIds = new Set();
    bulkDeleting = false;
    bulkDeleteDialogOpen = false;
  }

  function formatRelativeTime(dateStr: string | undefined | null): string {
    if (!dateStr) return "Never";
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHr = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHr / 24);

    if (diffDay > 30) return date.toLocaleDateString();
    if (diffDay > 0) return `${diffDay} day${diffDay !== 1 ? "s" : ""} ago`;
    if (diffHr > 0) return `${diffHr} hour${diffHr !== 1 ? "s" : ""} ago`;
    if (diffMin > 0) return `${diffMin} minute${diffMin !== 1 ? "s" : ""} ago`;
    return "Just now";
  }

  function getPermissionBadgeClass(perm: string): string {
    switch (perm?.toLowerCase()) {
      case "admin":
        return "bg-red-100 text-red-800";
      case "editor":
        return "bg-blue-100 text-blue-800";
      case "read":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  }
</script>

<div class="flex flex-col gap-3">
  <!-- Search bar -->
  <div class="flex items-center gap-2">
    <input
      type="text"
      placeholder="Search service tokens..."
      value={searchText}
      on:input={handleSearchInput}
      class="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
    />
  </div>

  <!-- Bulk action toolbar -->
  {#if selectionCount > 0}
    <div
      class="flex items-center justify-between rounded-md border border-blue-200 bg-blue-50 px-4 py-2"
    >
      <span class="text-sm font-medium text-blue-800">
        {selectionCount} token{selectionCount !== 1 ? "s" : ""} selected
      </span>
      <Button
        type="destructive"
        on:click={openBulkDeleteDialog}
        disabled={bulkDeleting}
      >
        {#if bulkDeleting}
          Deleting...
        {:else}
          Delete Selected
        {/if}
      </Button>
    </div>
  {/if}

  <!-- Table -->
  {#if isLoading}
    <div class="flex flex-col gap-2">
      {#each Array(5) as _}
        <div class="h-12 animate-pulse rounded bg-gray-100"></div>
      {/each}
    </div>
  {:else if isError}
    <div class="rounded-md border border-red-200 bg-red-50 p-6 text-center">
      <p class="mb-3 text-sm text-red-700">
        Failed to load service tokens. Please try again.
      </p>
      <Button on:click={() => $tokensQuery.refetch()}>Retry</Button>
    </div>
  {:else if tokens.length === 0}
    <div class="py-12 text-center">
      <div class="mb-2 text-4xl">🔑</div>
      <h3 class="mb-1 text-lg font-medium text-gray-900">
        No service tokens yet
      </h3>
      <p class="mb-4 text-sm text-gray-500">
        Create a service token for CI/CD pipelines or integrations.
      </p>
      <Button on:click={() => dispatch("create")}>Create Service Token</Button>
    </div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full min-w-[700px] text-left text-sm">
        <thead>
          <tr class="border-b border-gray-200 text-xs uppercase text-gray-500">
            <th class="w-10 px-3 py-3">
              <Checkbox
                checked={allOnPageSelected}
                indeterminate={someOnPageSelected}
                on:change={toggleSelectAll}
              />
            </th>
            <th class="px-3 py-3 font-medium">Name</th>
            <th class="px-3 py-3 font-medium">Token Prefix</th>
            <th class="px-3 py-3 font-medium">Scope</th>
            <th class="px-3 py-3 font-medium">Permissions</th>
            <th class="px-3 py-3 font-medium">Created</th>
            <th class="px-3 py-3 font-medium">Last Used</th>
            <th class="px-3 py-3 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each tokens as token (token.id)}
            <tr
              class="cursor-pointer border-b border-gray-100 transition-colors hover:bg-gray-50"
              class:bg-blue-50={selectedIds.has(token.id)}
              on:click={() => handleRowClick(token)}
            >
              <td class="px-3 py-3" on:click|stopPropagation>
                <Checkbox
                  checked={selectedIds.has(token.id)}
                  on:change={() => toggleSelectRow(token.id)}
                />
              </td>
              <td class="px-3 py-3 font-medium text-gray-900">
                {token.name || "Unnamed"}
              </td>
              <td class="px-3 py-3">
                <code
                  class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs"
                >
                  {token.tokenPrefix || "rillserv_****"}
                </code>
              </td>
              <td class="px-3 py-3 text-gray-600">
                {token.projectName || "Organization"}
              </td>
              <td class="px-3 py-3">
                <span
                  class="inline-block rounded-full px-2 py-0.5 text-xs font-medium {getPermissionBadgeClass(
                    token.permissions,
                  )}"
                >
                  {token.permissions || "read"}
                </span>
              </td>
              <td class="px-3 py-3 text-gray-500">
                {formatRelativeTime(token.createdAt)}
              </td>
              <td class="px-3 py-3 text-gray-500">
                {formatRelativeTime(token.lastUsedAt)}
              </td>
              <td class="px-3 py-3" on:click|stopPropagation>
                <div class="relative">
                  <button
                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                    on:click|stopPropagation={() => {
                      /* toggle menu handled inline */
                    }}
                  >
                    <svg
                      class="h-4 w-4"
                      fill="currentColor"
                      viewBox="0 0 16 16"
                    >
                      <circle cx="8" cy="3" r="1.5" />
                      <circle cx="8" cy="8" r="1.5" />
                      <circle cx="8" cy="13" r="1.5" />
                    </svg>
                  </button>
                  <!-- Inline action buttons as a simple row for now -->
                  <div class="flex gap-1">
                    <button
                      class="rounded px-2 py-1 text-xs text-gray-600 hover:bg-gray-100"
                      on:click|stopPropagation={() => openViewDetails(token)}
                    >
                      Details
                    </button>
                    <button
                      class="rounded px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                      on:click|stopPropagation={() => openDeleteDialog(token)}
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    {#if nextPageToken}
      <div class="flex justify-center pt-2">
        <Button type="secondary" on:click={loadNextPage}>Load More</Button>
      </div>
    {/if}
  {/if}
</div>

<!-- Single delete confirmation dialog -->
{#if deleteDialogOpen && tokenToDelete}
  <DeleteTokenDialog
    tokenName={tokenToDelete.name}
    tokenType="service"
    onConfirm={handleDeleteConfirm}
    onCancel={closeDeleteDialog}
  />
{/if}

<!-- Bulk delete confirmation dialog -->
{#if bulkDeleteDialogOpen}
  <DeleteTokenDialog
    tokenName="{selectionCount} token{selectionCount !== 1 ? 's' : ''}"
    tokenType="service"
    onConfirm={handleBulkDeleteConfirm}
    onCancel={closeBulkDeleteDialog}
    customMessage="Are you sure you want to revoke {selectionCount} token{selectionCount !== 1 ? 's' : ''}? This action cannot be undone. Any integrations using these tokens will immediately lose access."
  />
{/if}

<!-- Token details drawer -->
{#if drawerOpen && selectedTokenId}
  <TokenDetailsDrawer
    tokenId={selectedTokenId}
    tokenType="service"
    orgId={organization}
    open={drawerOpen}
    on:close={closeDrawer}
  />
{/if}