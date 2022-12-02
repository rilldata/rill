<script lang="ts">
  import { EntityStatus } from "@rilldata/web-local/lib/temp/entity";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import { slideRight } from "../../transitions";

  import { Switch } from "@rilldata/web-local/lib/components/button";

  import Shortcut from "../tooltip/Shortcut.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../tooltip/TooltipTitle.svelte";

  import Back from "../icons/Back.svelte";
  import Search from "../icons/Search.svelte";

  import { metricsExplorerStore } from "../../application-state-stores/explorer-stores";
  import Close from "../icons/Close.svelte";
  import SearchBar from "../search/Search.svelte";
  import Spinner from "../Spinner.svelte";

  export let metricViewName: string;
  export let dimensionName: string;
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
    metricsExplorerStore.setMetricDimensionName(metricViewName, null);
  };
  function toggleFilterMode() {
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }
</script>

<div
  class="grid grid-auto-cols justify-between grid-flow-col items-center p-1 pb-3"
  style:height="50px"
>
  <button
    class="flex flex-row items-center"
    on:click={() => goBackToLeaderboard()}
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
    style:cursor="pointer"
    style:grid-column-gap=".4rem"
  >
    <Tooltip distance={16} location="left">
      <div class="mr-3 ui-copy-icon" style:grid-column-gap=".4rem">
        <Switch checked={excludeMode} on:click={() => toggleFilterMode()}>
          Exclude
        </Switch>
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
