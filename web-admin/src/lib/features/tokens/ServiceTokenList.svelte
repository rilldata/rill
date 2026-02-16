<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Search } from "@rilldata/web-common/components/icons";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    createServiceTokenListQuery,
    deleteServiceTokenMutation,
  } from "./token-queries";
  import { formatTimeAgo } from "@rilldata/web-common/lib/time";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let organization: string;
  export let onCreateToken: () => void = () => {};

  const dispatch = createEventDispatcher<{
    viewDetails: { tokenId: string };
    deleteToken: { tokenId: string; tokenName: string };
  }>();

  let searchText = "";
  let debouncedSearch = "";
  let debounceTimer: ReturnType<typeof setTimeout>;
  let pageToken = "";
  let pageTokenHistory: string[] = [];

  function handleSearchInput(event: Event) {
    const target = event.target as HTMLInputElement;
    searchText = target.value;
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debouncedSearch = searchText;
      pageToken = "";
      pageTokenHistory = [];
    }, 300);
  }

  $: tokenListQuery = createServiceTokenListQuery(organization, {
    search: debouncedSearch || undefined,
    pageSize: 20,
    pageToken: pageToken || undefined,
  });

  $: tokens = $tokenListQuery.data?.tokens ?? [];
  $: nextPageToken = $tokenListQuery.data?.nextPageToken ?? "";
  $: isLoading = $tokenListQuery.isLoading;
  $: isError = $tokenListQuery.isError;
  $: error = $tokenListQuery.error;

  function handleNextPage() {
    if (nextPageToken) {
      pageTokenHistory = [...pageTokenHistory, pageToken];
      pageToken = nextPageToken;
    }
  }

  function handlePreviousPage() {
    if (pageTokenHistory.length > 0) {
      const history = [...pageTokenHistory];
      pageToken = history.pop() ?? "";
      pageTokenHistory = history;
    }
  }

  function handleViewDetails(tokenId: string) {
    dispatch("viewDetails", { tokenId });
  }

  function handleDeleteToken(tokenId: string, tokenName: string) {
    dispatch("deleteToken", { tokenId, tokenName });
  }

  function handleRetry() {
    $tokenListQuery.refetch();
  }

  function formatPermission(permission: string): string {
    if (!permission) return "Unknown";
    const lower = permission.toLowerCase();
    if (lower.includes("admin")) return "Admin";
    if (lower.includes("editor") || lower.includes("write")) return "Editor";
    if (lower.includes("read") || lower.includes("viewer")) return "Read";
    return permission;
  }

  function permissionBadgeClass(permission: string): string {
    const formatted = formatPermission(permission).toLowerCase();
    switch (formatted) {
      case "admin":
        return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
      case "editor":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200";
      case "read":
        return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200";
    }
  }

  function formatScope(token: any): string {
    if (token.projectName) return token.projectName;
    return "Organization";
  }

  function formatTokenPrefix(token: any): string {
    return token.tokenPrefix || token.id?.slice(0, 12) + "..." || "—";
  }

  function formatLastUsed(token: any): string {
    if (token.lastUsedOn) {
      return formatTimeAgo(new Date(token.lastUsedOn));
    }
    return "Never";
  }

  function formatCreated(token: any): string {
    if (token.createdOn) {
      return formatTimeAgo(new Date(token.createdOn));
    }
    return "—";
  }

  const SKELETON_ROWS = 5;
  const SKELETON_COLS = 7;
</script>

<div class="flex flex-col gap-4">
  <!-- Search bar -->
  <div class="relative">
    <div
      class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3"
    >
      <svg
        class="h-4 w-4 text-gray-400"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
        />
      </svg>
    </div>
    <input
      type="text"
      placeholder="Search service tokens..."
      value={searchText}
      on:input={handleSearchInput}
      class="w-full rounded-md border border-gray-300 bg-white py-2 pl-10 pr-3 text-sm placeholder-gray-400 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
    />
  </div>

  <!-- Table -->
  <div class="overflow-x-auto rounded-md border border-gray-200">
    <table class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Name
          </th>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Token Prefix
          </th>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Scope
          </th>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Permissions
          </th>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Created
          </th>
          <th
            class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Last Used
          </th>
          <th
            class="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500"
          >
            Actions
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200 bg-white">
        {#if isLoading}
          {#each Array(SKELETON_ROWS) as _, rowIndex}
            <tr>
              {#each Array(SKELETON_COLS) as _, colIndex}
                <td class="px-4 py-3">
                  <div
                    class="h-4 animate-pulse rounded bg-gray-200"
                    style="width: {colIndex === 0
                      ? '140px'
                      : colIndex === 1
                        ? '120px'
                        : colIndex === 6
                          ? '24px'
                          : '80px'}"
                  ></div>
                </td>
              {/each}
            </tr>
          {/each}
        {:else if isError}
          <tr>
            <td colspan={SKELETON_COLS} class="px-4 py-12 text-center">
              <div class="flex flex-col items-center gap-3">
                <svg
                  class="h-8 w-8 text-red-400"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"
                  />
                </svg>
                <p class="text-sm text-gray-600">
                  Failed to load service tokens.
                  {#if error?.message}
                    {error.message}
                  {/if}
                </p>
                <Button on:click={handleRetry} type="secondary">
                  Retry
                </Button>
              </div>
            </td>
          </tr>
        {:else if tokens.length === 0}
          <tr>
            <td colspan={SKELETON_COLS} class="px-4 py-12 text-center">
              <div class="flex flex-col items-center gap-3">
                <svg
                  class="h-12 w-12 text-gray-300"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="1.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z"
                  />
                </svg>
                <div>
                  <p class="text-sm font-medium text-gray-900">
                    No service tokens yet
                  </p>
                  <p class="mt-1 text-sm text-gray-500">
                    Service tokens allow machine-to-machine access to your
                    organization's resources.
                  </p>
                </div>
                <Button on:click={onCreateToken} type="primary">
                  Create Service Token
                </Button>
              </div>
            </td>
          </tr>
        {:else}
          {#each tokens as token (token.id)}
            <tr
              class="cursor-pointer transition-colors hover:bg-gray-50"
              on:click={() => handleViewDetails(token.id)}
            >
              <td class="px-4 py-3">
                <div class="flex flex-col">
                  <span class="text-sm font-medium text-gray-900">
                    {token.name || "Unnamed Token"}
                  </span>
                  {#if token.description}
                    <span class="mt-0.5 text-xs text-gray-500 truncate max-w-[200px]">
                      {token.description}
                    </span>
                  {/if}
                </div>
              </td>
              <td class="px-4 py-3">
                <code
                  class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-700"
                >
                  {formatTokenPrefix(token)}
                </code>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600">
                {formatScope(token)}
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {permissionBadgeClass(
                    token.permissions || '',
                  )}"
                >
                  {formatPermission(token.permissions || "")}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-500">
                <Tooltip location="bottom">
                  <span>{formatCreated(token)}</span>
                  <TooltipContent slot="tooltip-content">
                    {token.createdOn
                      ? new Date(token.createdOn).toLocaleString()
                      : "—"}
                  </TooltipContent>
                </Tooltip>
              </td>
              <td class="px-4 py-3 text-sm text-gray-500">
                <Tooltip location="bottom">
                  <span
                    class:text-gray-400={!token.lastUsedOn}
                  >
                    {formatLastUsed(token)}
                  </span>
                  <TooltipContent slot="tooltip-content">
                    {token.lastUsedOn
                      ? new Date(token.lastUsedOn).toLocaleString()
                      : "Never used"}
                  </TooltipContent>
                </Tooltip>
              </td>
              <td class="px-4 py-3 text-right">
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <div
                  on:click|stopPropagation
                  role="button"
                  tabindex="0"
                  on:keydown|stopPropagation
                >
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild let:builder>
                      <button
                        use:builder.action
                        {...builder}
                        class="inline-flex items-center rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500"
                        aria-label="Token actions for {token.name || 'token'}"
                      >
                        <svg
                          class="h-5 w-5"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="currentColor"
                          viewBox="0 0 20 20"
                        >
                          <path
                            d="M10 6a2 2 0 110-4 2 2 0 010 4zm0 6a2 2 0 110-4 2 2 0 010 4zm0 6a2 2 0 110-4 2 2 0 010 4z"
                          />
                        </svg>
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem
                        on:click={() => handleViewDetails(token.id)}
                      >
                        <svg
                          class="mr-2 h-4 w-4"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                          />
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                          />
                        </svg>
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        on:click={() =>
                          handleDeleteToken(
                            token.id,
                            token.name || "Unnamed Token",
                          )}
                      >
                        <svg
                          class="mr-2 h-4 w-4 text-red-500"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          stroke-width="2"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                        <span class="text-red-600">Delete</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  {#if !isLoading && !isError && tokens.length > 0}
    <div class="flex items-center justify-between px-1">
      <span class="text-xs text-gray-500">
        {#if tokens.length > 0}
          Showing {tokens.length} token{tokens.length !== 1 ? "s" : ""}
        {/if}
      </span>
      <div class="flex items-center gap-2">
        {#if pageTokenHistory.length > 0}
          <Button type="secondary" on:click={handlePreviousPage}>
            ← Previous
          </Button>
        {/if}
        {#if nextPageToken}
          <Button type="secondary" on:click={handleNextPage}>
            Next →
          </Button>
        {/if}
      </div>
    </div>
  {/if}
</div>