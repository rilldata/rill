<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import { slideRight } from "../../transitions";

  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

  import Shortcut from "../tooltip/Shortcut.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";

  import Back from "../icons/Back.svelte";
  import Search from "../icons/Search.svelte";

  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Cancel from "../icons/Cancel.svelte";
  import Check from "../icons/Check.svelte";
  import Close from "../icons/Close.svelte";
  import SearchBar from "../search/Search.svelte";
  import Spinner from "../Spinner.svelte";

  export let metricsDefId: string;
  export let dimensionId: string;
  export let isFetching: boolean;
  export let excludeMode = false;

  $: filterKey = excludeMode ? "exclude" : "include";
  $: otherFilterKey = excludeMode ? "include" : "exclude";

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
  function toggleFilterMode() {
    metricsExplorerStore.toggleFilterMode(metricsDefId, dimensionId);
  }
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
      <span class="ui-copy-icon">
        <Back size="16px" />
      </span>
      <span> All Dimensions </span>
    {/if}
  </button>

  <div
    class="flex items-center"
    style:grid-column-gap=".4rem"
    style:cursor="pointer"
  >
    <Tooltip location="left" distance={16}>
      <div
        class="flex items-center mr-3 ui-copy-icon"
        style:grid-column-gap=".2rem"
        on:click={toggleFilterMode}
      >
        {#if excludeMode}<Check size="20px" /> Include{:else}<Cancel
            size="20px"
          /> Exclude{/if}
      </div>
      <TooltipContent slot="tooltip-content">
        <TooltipTitle>
          <svelte:fragment slot="name">
            Output {filterKey}s selected values
          </svelte:fragment>
        </TooltipTitle>
        <TooltipShortcutContainer>
          <div>toggle to {otherFilterKey} values</div>
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
      </TooltipContent>
    </Tooltip>

    {#if !searchToggle}
      <div
        class="flex items-center ui-copy-icon"
        in:fly={{ x: 10, duration: 300 }}
        style:grid-column-gap=".2rem"
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
        <span
          class="ui-copy-icon"
          style:cursor="pointer"
          on:click={() => closeSearchBar()}
        >
          <Close />
        </span>
      </div>
    {/if}
  </div>
</div>
