<script lang="ts">
  import { Switch } from "@rilldata/web-common/components/button";
  import Back from "@rilldata/web-common/components/icons/Back.svelte";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import SearchIcon from "@rilldata/web-common/components/icons/Search.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import ReplacePivotDialog from "@rilldata/web-common/features/dashboards/pivot/ReplacePivotDialog.svelte";
  import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { createEventDispatcher, onDestroy } from "svelte";
  import { fly } from "svelte/transition";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { SortType } from "../proto-state/derived-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import SelectAllButton from "./SelectAllButton.svelte";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import exportToplist from "./export-toplist";
  import {
    createQueryServiceExport,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { getDimensionTableExportArgs } from "./dimension-table-export-utils";

  export let dimensionName: string;
  export let isFetching: boolean;
  export let areAllTableRowsSelected = false;
  export let isRowsEmpty = true;

  const dispatch = createEventDispatcher();
  const exportDash = createQueryServiceExport();

  const stateManagers = getStateManagers();
  const {
    selectors: {
      sorting: { sortedByDimensionValue },
      dimensions: { getDimensionDisplayName },
      dimensionTable: { dimensionTableSearchString },
      dimensionFilters: { isFilterExcludeMode },
      measures: { visibleMeasures },
    },
    actions: {
      sorting: { toggleSort },
      dimensionTable: {
        setDimensionTableSearchString,
        clearDimensionTableSearchString,
      },
      dimensions: { setPrimaryDimension },
      dimensionsFilter: { toggleDimensionFilterMode },
    },
    dashboardStore,
    exploreName,
  } = stateManagers;

  const { adminServer, exports } = featureFlags;

  $: excludeMode = $isFilterExcludeMode(dimensionName);

  $: filterKey = excludeMode ? "exclude" : "include";
  $: otherFilterKey = excludeMode ? "include" : "exclude";

  $: metricsViewProto = $dashboardStore.proto;

  let searchBarOpen = false;

  // FIXME: this extra `searchText` variable should be eliminated,
  // but there is no way to make the <Search> component a fully
  // "controlled" component for now, so we have to go through the
  // `value` binding it exposes.
  let searchText: string | undefined = undefined;
  $: searchText = $dimensionTableSearchString;
  function onSearch() {
    setDimensionTableSearchString(searchText);
  }

  function closeSearchBar() {
    clearDimensionTableSearchString();
    searchBarOpen = false;
  }

  function onSubmit() {
    if (!areAllTableRowsSelected) {
      dispatch("toggle-all-search-items");
      closeSearchBar();
    }
  }

  const goBackToLeaderboard = () => {
    if ($sortedByDimensionValue) {
      toggleSort(SortType.VALUE);
    }
    setPrimaryDimension(undefined);
  };
  function toggleFilterMode() {
    toggleDimensionFilterMode(dimensionName);
  }

  let showReplacePivotModal = false;
  function startPivotForDimensionTable() {
    const pivot = $dashboardStore?.pivot;

    if (
      pivot.rows.dimension.length ||
      pivot.columns.measure.length ||
      pivot.columns.dimension.length
    ) {
      showReplacePivotModal = true;
    } else {
      createPivot();
    }
  }

  function createPivot() {
    showReplacePivotModal = false;

    const rowDimensions = dimensionName
      ? [
          {
            id: dimensionName,
            title: $getDimensionDisplayName(dimensionName),
            type: PivotChipType.Dimension,
          },
        ]
      : [];

    const measures = $visibleMeasures
      .filter((m) => m.name !== undefined)
      .map((m) => {
        return {
          id: m.name as string,
          title: m.label || (m.name as string),
          type: PivotChipType.Measure,
        };
      });

    metricsExplorerStore.createPivot(
      $exploreName,
      { dimension: rowDimensions },
      {
        dimension: [],
        measure: measures,
      },
    );
  }

  onDestroy(() => {
    clearDimensionTableSearchString();
  });

  const scheduledReportsQueryArgs = getDimensionTableExportArgs(stateManagers);

  const handleExportTopList = async (format: V1ExportFormat) => {
    await exportToplist({
      ctx: stateManagers,
      query: exportDash,
      format,
    });
  };
</script>

<div class="flex justify-between items-center p-1 pr-5 h-7">
  <button class="flex items-center" on:click={() => goBackToLeaderboard()}>
    {#if isFetching}
      <DelayedSpinner isLoading={isFetching} size="16px" />
    {:else}
      <span class="ui-copy-icon">
        <Back size="16px" />
      </span>
      <span>All Dimensions</span>
    {/if}
  </button>

  <!-- We fix the height to avoid a layout shift when the Search component is expanded. -->
  <div class="flex items-center gap-x-1 cursor-pointer h-9">
    {#if !isRowsEmpty}
      <SelectAllButton {areAllTableRowsSelected} on:toggle-all-search-items />
    {/if}
    {#if searchBarOpen || (searchText && searchText !== "")}
      <div
        transition:slideRight={{ leftOffset: 8 }}
        class="flex items-center gap-x-2 p-1.5"
      >
        <Search
          bind:value={searchText}
          on:input={onSearch}
          on:submit={onSubmit}
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
        <span>Search</span>
      </button>
    {/if}

    <Tooltip distance={16} location="left">
      <div class="flex items-center gap-x-1 px-1.5 ui-copy-icon">
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

    {#if $exports}
      <ExportMenu
        label="Export dimension table data"
        onExport={handleExportTopList}
        includeScheduledReport={$adminServer}
        queryArgs={$scheduledReportsQueryArgs}
        {metricsViewProto}
        exploreName={$exploreName}
      />
    {/if}
    <button
      class="h-6 px-1.5 py-px rounded-sm hover:bg-gray-200 text-gray-700"
      on:click={() => {
        startPivotForDimensionTable();
      }}
    >
      Start Pivot
    </button>
  </div>
</div>

<ReplacePivotDialog
  open={showReplacePivotModal}
  onCancel={() => {
    showReplacePivotModal = false;
  }}
  onReplace={createPivot}
/>
