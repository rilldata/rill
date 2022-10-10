<script lang="ts">
  import { slideRight } from "../../transitions";
  import { createEventDispatcher } from "svelte";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import Back from "../icons/Back.svelte";
  import Search from "../icons/Search.svelte";

  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Spinner from "../Spinner.svelte";
  import SearchBar from "../search/Search.svelte";
  import CrossIcon from "../icons/CrossIcon.svelte";

  export let metricsDefId: string;
  export let isFetching: boolean;

  let searchToggle = false;

  const dispatch = createEventDispatcher();

  let searchText = "";
  function onSearch() {
    dispatch("search", searchText);
  }

  const goBackToLeaderboard = () => {
    metricsExplorerStore.setMetricDimensionId(metricsDefId, null);
  };
</script>

<div
  class="grid grid-auto-cols justify-between grid-flow-col items-end p-1 pb-3"
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
        style:grid-column-gap=".4rem"
        on:click={() => (searchToggle = !searchToggle)}
      >
        <Search size="16px" />
        <span> Search </span>
      </div>
    {:else}
      <SearchBar bind:value={searchText} on:input={onSearch} />
      <CrossIcon on:click={() => (searchToggle = !searchToggle)} />
    {/if}
  </div>
</div>
