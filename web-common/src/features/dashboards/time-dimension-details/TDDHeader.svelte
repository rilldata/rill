<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { Chip } from "@rilldata/web-common/components/chip";

  export let metricViewName: string;
  export let dimensionName: string;

  const queryClient = useQueryClient();

  $: dashboardStore = useDashboardStore(metricViewName);

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

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

  function toggleFilterMode() {
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }
</script>

<div
  class="grid grid-auto-cols justify-between grid-flow-col items-center p-1 pb-3"
>
  <div class="flex gap-x-4 font-bold">
    <Chip label="Measure A">
      <div slot="body" class="flex">Measure A</div>
    </Chip>
    <span> | </span>
    <Chip>
      <div slot="body" class="flex">Time</div>
    </Chip>
    <span> : </span>

    <Chip label="Measure A">
      <div slot="body" class="flex">Measure A</div>
    </Chip>
  </div>

  <div
    class="flex items-center mr-4"
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
          <div>Toggle to {otherFilterKey} values</div>
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
      </TooltipContent>
    </Tooltip>

    {#if !searchToggle}
      <button
        class="flex items-center ui-copy-icon"
        in:fly={{ x: 10, duration: 300 }}
        style:grid-column-gap=".2rem"
        on:click={() => (searchToggle = !searchToggle)}
      >
        <SearchIcon size="16px" />
        <span> Search </span>
      </button>
    {:else}
      <div
        transition:slideRight|local={{ leftOffset: 8 }}
        class="flex items-center"
      >
        <Search bind:value={searchText} on:input={onSearch} />
        <button
          class="ui-copy-icon"
          style:cursor="pointer"
          on:click={() => closeSearchBar()}
        >
          <Close />
        </button>
      </div>
    {/if}
  </div>
</div>
