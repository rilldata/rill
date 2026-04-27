<script lang="ts">
  import SearchInput from "./SearchInput.svelte";
  import { searchOrgNames } from "@rilldata/web-admin/features/superuser/organizations/selectors";

  export let value: string = "";
  export let placeholder: string = "Organization name...";

  let searchQuery = "";
  let showResults = false;
  let highlightedIndex = -1;

  $: orgNamesQuery = searchOrgNames(searchQuery);
  $: matchedOrgs = extractUniqueOrgs($orgNamesQuery.data?.names ?? []);
  // Reset highlight when results change
  $: matchedOrgs, (highlightedIndex = -1);

  function extractUniqueOrgs(names: string[]): string[] {
    const orgs = new Set<string>();
    for (const name of names) {
      const slash = name.indexOf("/");
      if (slash > 0) orgs.add(name.substring(0, slash));
    }
    return [...orgs].sort();
  }

  function handleSearch(e: CustomEvent<string>) {
    searchQuery = e.detail;
    showResults = true;
    // Clear selection when user changes the search text
    if (value && searchQuery !== value) {
      value = "";
    }
  }

  function selectOrg(org: string) {
    value = org;
    searchQuery = org;
    showResults = false;
    highlightedIndex = -1;
  }

  function handleKeydown(e: CustomEvent<KeyboardEvent>) {
    const key = e.detail.key;
    if (!showResults || !matchedOrgs.length) return;

    if (key === "ArrowDown") {
      e.detail.preventDefault();
      highlightedIndex = (highlightedIndex + 1) % matchedOrgs.length;
    } else if (key === "ArrowUp") {
      e.detail.preventDefault();
      highlightedIndex =
        highlightedIndex <= 0 ? matchedOrgs.length - 1 : highlightedIndex - 1;
    } else if (key === "Enter" && highlightedIndex >= 0) {
      e.detail.preventDefault();
      selectOrg(matchedOrgs[highlightedIndex]);
    } else if (key === "Escape") {
      showResults = false;
      highlightedIndex = -1;
    }
  }
</script>

<div class="relative">
  <SearchInput
    bind:value={searchQuery}
    {placeholder}
    on:search={handleSearch}
    on:keydown={handleKeydown}
  />
  {#if showResults && searchQuery.length >= 3 && !value}
    {#if $orgNamesQuery.isFetching}
      <div
        class="absolute z-10 left-0 right-0 mt-1 rounded-md border bg-surface-base shadow-md p-2"
      >
        <p class="text-sm text-fg-secondary">Searching...</p>
      </div>
    {:else if matchedOrgs.length > 0}
      <div
        class="absolute z-10 left-0 right-0 mt-1 flex flex-col gap-0.5 max-h-48 overflow-y-auto rounded-md border bg-surface-base shadow-md"
        role="listbox"
      >
        {#each matchedOrgs as org, i}
          <button
            class="px-3 py-2 text-sm text-fg-primary text-left hover:bg-surface-hover cursor-pointer"
            class:bg-surface-hover={i === highlightedIndex}
            role="option"
            aria-selected={i === highlightedIndex}
            on:click={() => selectOrg(org)}
          >
            {org}
          </button>
        {/each}
      </div>
    {:else if $orgNamesQuery.isSuccess}
      <div
        class="absolute z-10 left-0 right-0 mt-1 rounded-md border bg-surface-base shadow-md p-2"
      >
        <p class="text-sm text-fg-secondary">
          No organizations found. Note: orgs with zero projects won't appear in
          search.
        </p>
      </div>
    {/if}
  {/if}
</div>
