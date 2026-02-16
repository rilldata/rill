<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { EllipsisIcon, KeyRound, Trash2, Eye } from "lucide-svelte";
  import { createUserTokenListQuery } from "./token-queries";
  import { timeAgo } from "@rilldata/web-common/lib/time";

  export let organization: string;

  const dispatch = createEventDispatcher<{
    create: void;
    delete: { tokenId: string; tokenName: string };
    viewDetails: { tokenId: string };
  }>();

  let searchText = "";
  let debouncedSearch = "";
  let debounceTimer: ReturnType<typeof setTimeout>;
  let pageToken: string | undefined = undefined;
  let pageTokenStack: string[] = [];

  function handleSearchInput(value: string) {
    searchText = value;
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debouncedSearch = searchText;
      pageToken = undefined;
      pageTokenStack = [];
    }, 300);
  }

  $: listQuery = createUserTokenListQuery({
    pageSize: 20,
    pageToken,
  });

  $: tokens = $listQuery.data?.tokens ?? [];
  $: nextPageToken = $listQuery.data?.nextPageToken;
  $: isLoading = $listQuery.isLoading;
  $: isError = $listQuery.isError;
  $: error = $listQuery.error;

  // Client-side search filtering since user token API may not support server-side search
  $: filteredTokens = debouncedSearch
    ? tokens.filter(
        (t) =>
          t.description?.toLowerCase().includes(debouncedSearch.toLowerCase()) ||
          t.id?.toLowerCase().includes(debouncedSearch.toLowerCase()),
      )
    : tokens;

  function isExpired(expiresAt: string | undefined): boolean {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  }

  function formatDate(dateStr: string | undefined): string {
    if (!dateStr) return "Never";
    return timeAgo(new Date(dateStr));
  }

  function formatExpiration(expiresAt: string | undefined): string {
    if (!expiresAt) return "Never";
    const date = new Date(expiresAt);
    if (date < new Date()) {
      return `Expired ${timeAgo(date)}`;
    }
    return timeAgo(date);
  }

  function getTokenPrefix(token: { id?: string }): string {
    // The API may return a tokenPrefix field or we derive from id
    return (token as any).tokenPrefix ?? `rilluser_****`;
  }

  function getTokenName(token: { description?: string; id?: string }): string {
    return token.description || token.id || "Unnamed token";
  }

  function handleNextPage() {
    if (nextPageToken) {
      if (pageToken) {
        pageTokenStack = [...pageTokenStack, pageToken];
      }
      pageToken = nextPageToken;
    }
  }

  function handlePrevPage() {
    if (pageTokenStack.length > 0) {
      const stack = [...pageTokenStack];
      pageToken = stack.pop();
      pageTokenStack = stack;
    } else {
      pageToken = undefined;
    }
  }

  function handleDelete(token: { id?: string; description?: string }) {
    dispatch("delete", {
      tokenId: token.id ?? "",
      tokenName: getTokenName(token),
    });
  }

  function handleViewDetails(token: { id?: string }) {
    dispatch("viewDetails", { tokenId: token.id ?? "" });
  }
</script>

<div class="flex flex-col gap-4">
  <!-- Search -->
  <div class="flex items-center gap-2">
    <div class="w-full max-w-sm">
      <Search
        value={searchText}
        on:input={(e) => handleSearchInput(e.detail ?? e.target?.value ?? "")}
        placeholder="Search tokens..."
      />
    </div>
  </div>

  <!-- Loading state -->
  {#if isLoading}
    <div class="border rounded-md overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b bg-gray-50">
            <th class="text-left py-3 px-4 font-medium text-gray-600">Name</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Token Prefix</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Created</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Expires</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Last Used</th>
            <th class="text-right py-3 px-4 font-medium text-gray-600">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each Array(5) as _}
            <tr class="border-b">
              <td class="py-3 px-4"><div class="h-4 w-32 bg-gray-200 rounded animate-pulse" /></td>
              <td class="py-3 px-4"><div class="h-4 w-24 bg-gray-200 rounded animate-pulse" /></td>
              <td class="py-3 px-4"><div class="h-4 w-20 bg-gray-200 rounded animate-pulse" /></td>
              <td class="py-3 px-4"><div class="h-4 w-20 bg-gray-200 rounded animate-pulse" /></td>
              <td class="py-3 px-4"><div class="h-4 w-20 bg-gray-200 rounded animate-pulse" /></td>
              <td class="py-3 px-4"><div class="h-4 w-6 bg-gray-200 rounded animate-pulse ml-auto" /></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

  <!-- Error state -->
  {:else if isError}
    <div class="border rounded-md p-8 flex flex-col items-center gap-4 bg-red-50">
      <p class="text-red-600 text-sm">
        {error?.message ?? "Failed to load user tokens. Please try again."}
      </p>
      <Button variant="secondary" on:click={() => $listQuery.refetch()}>
        Retry
      </Button>
    </div>

  <!-- Empty state -->
  {:else if filteredTokens.length === 0 && !debouncedSearch}
    <div class="border rounded-md p-12 flex flex-col items-center gap-4">
      <div class="rounded-full bg-gray-100 p-3">
        <KeyRound class="h-6 w-6 text-gray-400" />
      </div>
      <div class="text-center">
        <p class="text-sm font-medium text-gray-900">No personal tokens yet</p>
        <p class="text-sm text-gray-500 mt-1">
          Create one to use the Rill CLI or API.
        </p>
      </div>
      <Button on:click={() => dispatch("create")}>
        Create User Token
      </Button>
    </div>

  <!-- Empty search results -->
  {:else if filteredTokens.length === 0 && debouncedSearch}
    <div class="border rounded-md p-8 flex flex-col items-center gap-3">
      <p class="text-sm text-gray-500">
        No tokens matching "<span class="font-medium">{debouncedSearch}</span>"
      </p>
      <Button
        variant="secondary"
        on:click={() => {
          searchText = "";
          debouncedSearch = "";
        }}
      >
        Clear search
      </Button>
    </div>

  <!-- Token table -->
  {:else}
    <div class="border rounded-md overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b bg-gray-50">
            <th class="text-left py-3 px-4 font-medium text-gray-600">Name</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Token Prefix</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Created</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Expires</th>
            <th class="text-left py-3 px-4 font-medium text-gray-600">Last Used</th>
            <th class="text-right py-3 px-4 font-medium text-gray-600">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredTokens as token (token.id)}
            {@const expired = isExpired(token.expiresOn)}
            <tr
              class="border-b last:border-b-0 hover:bg-gray-50 cursor-pointer transition-colors"
              class:opacity-60={expired}
              on:click={() => handleViewDetails(token)}
            >
              <td class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <span
                    class="font-medium text-gray-900"
                    class:line-through={expired}
                  >
                    {getTokenName(token)}
                  </span>
                  {#if expired}
                    <span
                      class="inline-flex items-center rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700"
                    >
                      Expired
                    </span>
                  {/if}
                </div>
              </td>
              <td class="py-3 px-4">
                <code class="text-xs font-mono bg-gray-100 px-1.5 py-0.5 rounded text-gray-700">
                  {getTokenPrefix(token)}
                </code>
              </td>
              <td class="py-3 px-4 text-gray-600">
                {formatDate(token.createdOn)}
              </td>
              <td class="py-3 px-4">
                <span
                  class:text-red-600={expired}
                  class:font-medium={expired}
                  class:text-gray-600={!expired}
                >
                  {formatExpiration(token.expiresOn)}
                </span>
              </td>
              <td class="py-3 px-4 text-gray-600">
                {formatDate(token.lastUsedOn)}
              </td>
              <td class="py-3 px-4 text-right">
                <!-- svelte-ignore a11y-click-events-have-key-events -->
                <div
                  on:click|stopPropagation
                  role="presentation"
                >
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild let:builder>
                      <Button
                        builders={[builder]}
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8"
                      >
                        <EllipsisIcon class="h-4 w-4" />
                        <span class="sr-only">Open menu</span>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem on:click={() => handleViewDetails(token)}>
                        <Eye class="h-4 w-4 mr-2" />
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        class="text-red-600"
                        on:click={() => handleDelete(token)}
                      >
                        <Trash2 class="h-4 w-4 mr-2" />
                        Revoke Token
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    {#if nextPageToken || pageTokenStack.length > 0}
      <div class="flex items-center justify-between pt-2">
        <p class="text-sm text-gray-500">
          Showing {filteredTokens.length} tokens
        </p>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={pageTokenStack.length === 0 && !pageToken}
            on:click={handlePrevPage}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!nextPageToken}
            on:click={handleNextPage}
          >
            Next
          </Button>
        </div>
      </div>
    {/if}
  {/if}
</div>