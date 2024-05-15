<script lang="ts">
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import GlobalDimensionSearchResults from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearchResults.svelte";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { SearchIcon } from "lucide-svelte";
  import { fly } from "svelte/transition";

  export let metricsViewName: string;

  let searchBarOpen = false;
  let searchText = "";

  function closeSearchBar() {
    searchBarOpen = false;
    searchText = "";
  }

  let submittedSearchText = "";
  function onSubmit() {
    submittedSearchText = searchText;
  }
</script>

{#if searchBarOpen}
  <div
    transition:slideRight={{ leftOffset: 8 }}
    class="flex items-center gap-x-2"
  >
    <Search
      bind:value={searchText}
      on:submit={onSubmit}
      placeholder="Search dimensions"
    />
    <button class="ui-copy-icon" on:click={() => closeSearchBar()}>
      <Close />
    </button>
  </div>
{:else}
  <button
    class="flex items-center gap-x-2 p-1.5 text-gray-700"
    in:fly|global={{ x: 10, duration: 300 }}
    on:click={() => (searchBarOpen = !searchBarOpen)}
  >
    <SearchIcon size="16px" />
  </button>
{/if}

<GlobalDimensionSearchResults
  {metricsViewName}
  searchText={submittedSearchText}
/>
