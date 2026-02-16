<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { createUserTokenListQuery } from "./token-queries";
  import { deleteUserTokenMutation } from "./token-queries";
  import DeleteTokenDialog from "./DeleteTokenDialog.svelte";
  import TokenDetailsDrawer from "./TokenDetailsDrawer.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { formatTimeAgo } from "@rilldata/web-common/lib/time";

  export let orgId: string;

  const dispatch = createEventDispatcher();

  let search = "";
  let debouncedSearch = "";
  let pageToken: string | undefined = undefined;
  let deleteDialogOpen = false;
  let tokenToDelete: { id: string; name: string } | null = null;
  let drawerOpen = false;
  let selectedTokenId: string | null = null;

  const debouncedSetSearch = debounce((value: string) => {
    debouncedSearch = value;
    pageToken = undefined;
  }, 300);

  $: debouncedSetSearch(search);

  $: tokenListQuery = createUserTokenListQuery({
    search: debouncedSearch || undefined,
    pageToken,
  });

  $: tokens = $tokenListQuery?.data?.tokens ?? [];
  $: nextPageToken = $tokenListQuery?.data?.nextPageToken;
  $: isLoading = $tokenListQuery?.isLoading ?? true;
  $: isError = $tokenListQuery?.isError ?? false;
  $: error = $tokenListQuery?.error;

  const deleteMutation = deleteUserTokenMutation();

  function handleSearch(event: Event) {
    search = (event.target as HTMLInputElement).value;
  }

  function handleLoadMore() {
    if (nextPageToken) {
      pageToken = nextPageToken;
    }
  }

  function openDeleteDialog(token: { id: string; name: string }) {
    tokenToDelete = token;
    deleteDialogOpen = true;
  }

  function closeDeleteDialog() {
    deleteDialogOpen = false;
    tokenToDelete = null;
  }

  function openDrawer(tokenId: string) {
    selectedTokenId = tokenId;
    drawerOpen = true;
  }

  function closeDrawer() {
    drawerOpen = false;
    selectedTokenId = null;
  }

  async function handleDelete() {
    if (!tokenToDelete) return;
    await $deleteMutation.mutateAsync({ tokenId: tokenToDelete.id });
    eventBus.emit("notification", {
      message: `Token "${tokenToDelete.name}" has been revoked.`,
      type: "success",
    });
    closeDeleteDialog();
  }

  function isExpired(expiresAt: string | undefined | null): boolean {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  }

  function handleRowClick(tokenId: string) {
    openDrawer(tokenId);
  }

  function handleRowKeydown(event: KeyboardEvent, tokenId: string) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openDrawer(tokenId);
    }
  }
</script>

<div class="user-token-list">
  <div class="list-toolbar">
    <input
      type="text"
      class="search-input"
      placeholder="Search user tokens..."
      value={search}
      on:input={handleSearch}
    />
  </div>

  {#if isLoading}
    <div class="skeleton-container">
      {#each Array(5) as _}
        <div class="skeleton-row">
          <div class="skeleton-cell wide" />
          <div class="skeleton-cell medium" />
          <div class="skeleton-cell medium" />
          <div class="skeleton-cell medium" />
          <div class="skeleton-cell medium" />
          <div class="skeleton-cell narrow" />
        </div>
      {/each}
    </div>
  {:else if isError}
    <div class="error-state">
      <p class="error-message">
        Failed to load user tokens{error?.message ? `: ${error.message}` : ""}.
      </p>
      <button
        class="retry-button"
        on:click={() => $tokenListQuery.refetch()}
      >
        Retry
      </button>
    </div>
  {:else if tokens.length === 0 && !debouncedSearch}
    <div class="empty-state">
      <div class="empty-icon">🔑</div>
      <h3>No personal tokens yet</h3>
      <p>Create one to use the Rill CLI or API.</p>
      <button
        class="create-button"
        on:click={() => dispatch("create")}
      >
        Create User Token
      </button>
    </div>
  {:else if tokens.length === 0 && debouncedSearch}
    <div class="empty-state">
      <p>No tokens match "{debouncedSearch}".</p>
    </div>
  {:else}
    <div class="table-container">
      <table class="token-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Token Prefix</th>
            <th>Created</th>
            <th>Expires</th>
            <th>Last Used</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each tokens as token (token.id)}
            {@const expired = isExpired(token.expiresAt)}
            <tr
              class="token-row"
              class:expired
              role="button"
              tabindex="0"
              on:click={() => handleRowClick(token.id)}
              on:keydown={(e) => handleRowKeydown(e, token.id)}
            >
              <td class="name-cell">
                <span class="token-name" class:expired-text={expired}>
                  {token.name || "Unnamed Token"}
                </span>
                {#if expired}
                  <span class="badge badge-expired">Expired</span>
                {/if}
              </td>
              <td class="prefix-cell">
                <code class="token-prefix" class:expired-text={expired}>
                  {token.tokenPrefix || "rilluser_****"}
                </code>
              </td>
              <td class="date-cell">
                {token.createdAt ? formatTimeAgo(new Date(token.createdAt)) : "—"}
              </td>
              <td class="date-cell">
                {#if token.expiresAt}
                  <span class:expired-text={expired}>
                    {formatTimeAgo(new Date(token.expiresAt))}
                  </span>
                {:else}
                  <span class="never-text">Never</span>
                {/if}
              </td>
              <td class="date-cell">
                {token.lastUsedAt ? formatTimeAgo(new Date(token.lastUsedAt)) : "Never"}
              </td>
              <td class="actions-cell">
                <div class="actions-menu">
                  <button
                    class="action-button"
                    title="Actions"
                    on:click|stopPropagation={() => {}}
                  >
                    ⋮
                  </button>
                  <div class="dropdown-menu">
                    <button
                      class="dropdown-item"
                      on:click|stopPropagation={() => openDrawer(token.id)}
                    >
                      View Details
                    </button>
                    <button
                      class="dropdown-item destructive"
                      on:click|stopPropagation={() =>
                        openDeleteDialog({
                          id: token.id,
                          name: token.name || "Unnamed Token",
                        })}
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

    {#if nextPageToken}
      <div class="pagination">
        <button class="load-more-button" on:click={handleLoadMore}>
          Load More
        </button>
      </div>
    {/if}
  {/if}
</div>

{#if deleteDialogOpen && tokenToDelete}
  <DeleteTokenDialog
    tokenName={tokenToDelete.name}
    tokenType="user"
    onConfirm={handleDelete}
    onCancel={closeDeleteDialog}
  />
{/if}

<TokenDetailsDrawer
  tokenId={selectedTokenId ?? ""}
  tokenType="user"
  {orgId}
  open={drawerOpen}
  on:close={closeDrawer}
  on:revoke={(e) => {
    closeDrawer();
    openDeleteDialog({ id: e.detail.tokenId, name: e.detail.tokenName });
  }}
/>

<style lang="postcss">
  .user-token-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .list-toolbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .search-input {
    flex: 1;
    max-width: 320px;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 6px;
    font-size: 0.875rem;
    outline: none;
    transition: border-color 0.15s ease;
  }

  .search-input:focus {
    border-color: var(--color-primary, #6366f1);
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
  }

  .skeleton-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .skeleton-row {
    display: flex;
    gap: 1rem;
    padding: 0.75rem;
    border-radius: 4px;
    background: var(--color-surface, #f8fafc);
    animation: pulse 1.5s ease-in-out infinite;
  }

  .skeleton-cell {
    height: 1rem;
    border-radius: 4px;
    background: var(--color-border, #e2e8f0);
  }

  .skeleton-cell.wide {
    flex: 2;
  }

  .skeleton-cell.medium {
    flex: 1;
  }

  .skeleton-cell.narrow {
    flex: 0.5;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
    padding: 2rem;
    text-align: center;
  }

  .error-message {
    color: var(--color-error, #ef4444);
    font-size: 0.875rem;
  }

  .retry-button {
    padding: 0.5rem 1rem;
    background: var(--color-primary, #6366f1);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    cursor: pointer;
  }

  .retry-button:hover {
    opacity: 0.9;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 3rem 1rem;
    text-align: center;
    color: var(--color-text-secondary, #64748b);
  }

  .empty-icon {
    font-size: 2.5rem;
    margin-bottom: 0.5rem;
  }

  .empty-state h3 {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-text, #1e293b);
    margin: 0;
  }

  .empty-state p {
    font-size: 0.875rem;
    margin: 0;
  }

  .create-button {
    margin-top: 0.75rem;
    padding: 0.5rem 1rem;
    background: var(--color-primary, #6366f1);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
  }

  .create-button:hover {
    opacity: 0.9;
  }

  .table-container {
    overflow-x: auto;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 8px;
  }

  .token-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .token-table th {
    text-align: left;
    padding: 0.75rem 1rem;
    font-weight: 600;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--color-text-secondary, #64748b);
    background: var(--color-surface, #f8fafc);
    border-bottom: 1px solid var(--color-border, #e2e8f0);
    white-space: nowrap;
  }

  .token-table td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--color-border-light, #f1f5f9);
    white-space: nowrap;
  }

  .token-row {
    cursor: pointer;
    transition: background-color 0.1s ease;
  }

  .token-row:hover {
    background-color: var(--color-surface, #f8fafc);
  }

  .token-row:focus-visible {
    outline: 2px solid var(--color-primary, #6366f1);
    outline-offset: -2px;
  }

  .token-row.expired {
    opacity: 0.7;
  }

  .name-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .token-name {
    font-weight: 500;
    color: var(--color-text, #1e293b);
  }

  .expired-text {
    text-decoration: line-through;
    color: var(--color-text-secondary, #94a3b8);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.6875rem;
    font-weight: 500;
    white-space: nowrap;
  }

  .badge-expired {
    background: var(--color-error-light, #fef2f2);
    color: var(--color-error, #ef4444);
  }

  .prefix-cell code {
    font-family: "SF Mono", SFMono-Regular, Consolas, "Liberation Mono",
      Menlo, monospace;
    font-size: 0.8125rem;
    padding: 0.125rem 0.375rem;
    background: var(--color-surface, #f1f5f9);
    border-radius: 4px;
  }

  .date-cell {
    color: var(--color-text-secondary, #64748b);
    font-size: 0.8125rem;
  }

  .never-text {
    color: var(--color-text-tertiary, #94a3b8);
    font-style: italic;
  }

  .actions-cell {
    width: 3rem;
    text-align: center;
  }

  .actions-menu {
    position: relative;
    display: inline-block;
  }

  .action-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--color-text-secondary, #64748b);
    font-size: 1.125rem;
    cursor: pointer;
    line-height: 1;
  }

  .action-button:hover {
    background: var(--color-surface, #f1f5f9);
    color: var(--color-text, #1e293b);
  }

  .dropdown-menu {
    display: none;
    position: absolute;
    right: 0;
    top: 100%;
    z-index: 20;
    min-width: 140px;
    background: white;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 6px;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    padding: 0.25rem;
  }

  .actions-menu:focus-within .dropdown-menu {
    display: flex;
    flex-direction: column;
  }

  .dropdown-item {
    display: block;
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: none;
    border-radius: 4px;
    background: transparent;
    text-align: left;
    font-size: 0.8125rem;
    color: var(--color-text, #1e293b);
    cursor: pointer;
    white-space: nowrap;
  }

  .dropdown-item:hover {
    background: var(--color-surface, #f8fafc);
  }

  .dropdown-item.destructive {
    color: var(--color-error, #ef4444);
  }

  .dropdown-item.destructive:hover {
    background: var(--color-error-light, #fef2f2);
  }

  .pagination {
    display: flex;
    justify-content: center;
    padding: 0.75rem 0;
  }

  .load-more-button {
    padding: 0.5rem 1.5rem;
    background: white;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 6px;
    font-size: 0.875rem;
    color: var(--color-text, #1e293b);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .load-more-button:hover {
    background: var(--color-surface, #f8fafc);
    border-color: var(--color-border-dark, #cbd5e1);
  }
</style>