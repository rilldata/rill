<script lang="ts">
  import { slideRight } from "../../transitions";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";

  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import Back from "../icons/Back.svelte";
  import Search from "../icons/Search.svelte";

  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Spinner from "../Spinner.svelte";
  import SearchBar from "../search/Search.svelte";
  import Close from "../icons/Close.svelte";

  export let metricsDefId: string;
  export let isFetching: boolean;

  let searchToggle = false;

  const dispatch = createEventDispatcher();

  let searchText = "";
  function onSearch() {
    dispatch("search", searchText);
  }

  function closeSearchBar() {
    searchText = "";
    searchToggle = !searchToggle;
    onSearch();
  }

  const goBackToLeaderboard = () => {
    metricsExplorerStore.setMetricDimensionId(metricsDefId, null);
  };
</script>

<div
  class="grid grid-auto-cols justify-between grid-flow-col items-center p-1 pb-3"
  style:height="50px"
>
  <button
    on:click={() => goBackToLeaderboard()}
    class="flex flex-row items-center"
    style:grid-column-gap=".4rem"
  >
    {#if isFetching}
      <div transition:slideRight|local={{ leftOffset: 8 }}>
        <Spinner size="16px" status={EntityStatus.Running} />
      </div>
    {:else}
      <Back size="16px" />
      <span> All Dimensions </span>
    {/if}
  </button>

  <div class="flex items-center" style:grid-column-gap=".4rem">
    {#if !searchToggle}
      <div
        class="flex items-center"
        in:fly={{ x: 10, duration: 300 }}
        style:grid-column-gap=".4rem"
        style:cursor="pointer"
        on:click={() => (searchToggle = !searchToggle)}
      >
        <Search size="16px" />
        <span> Search </span>
      </div>
    {:else}
      <div
        transition:slideRight|local={{ leftOffset: 8 }}
        class="flex items-center"
      >
        <SearchBar bind:value={searchText} on:input={onSearch} />
        <span style:cursor="pointer" on:click={() => closeSearchBar()}>
          <Close />
        </span>
      </div>
    {/if}
  </div>
</div>
