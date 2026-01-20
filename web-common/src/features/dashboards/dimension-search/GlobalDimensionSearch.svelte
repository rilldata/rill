<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import GlobalDimensionSearchResults from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearchResults.svelte";
  import { slideRight } from "@rilldata/web-common/lib/transitions";

  let searchBarOpen = false;
  let searchText = "";

  function reset() {
    searchBarOpen = false;
  }

  let submittedSearchText = "";
  let searchResultsOpen = false;
  function onSubmit() {
    submittedSearchText = searchText;
    searchResultsOpen = true;
  }
</script>

<div class="relative flex flex-row">
  {#if searchBarOpen}
    <div
      transition:slideRight={{}}
      class="flex items-center gap-x-2 pr-2 w-60 bg-gray-50 border border-primary-300"
    >
      <Search
        bind:value={searchText}
        {onSubmit}
        placeholder="Search dimensions"
        autofocus
        border={false}
        background={false}
      />
      <button class="text-fg-secondary" on:click={reset}>
        <Cancel size="16px" />
      </button>
    </div>
  {:else}
    <Button
      class="flex items-center gap-x-2 p-1.5 text-fg-primary"
      onClick={() => (searchBarOpen = !searchBarOpen)}
      type="secondary"
      compact
    >
      <SearchIcon size="16px" />
    </Button>
  {/if}

  <GlobalDimensionSearchResults
    searchText={submittedSearchText}
    onSelect={reset}
    bind:open={searchResultsOpen}
  />
</div>
