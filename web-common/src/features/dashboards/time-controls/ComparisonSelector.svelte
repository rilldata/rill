<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. read the existing active comparison
3. update comparisons on user interactions
-->
<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { NO_COMPARISON_LABEL } from "@rilldata/web-common/lib/time/config";
  import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
  import { matchSorter } from "match-sorter";

  export let exploreName: string;
  export let chipStyle = false;

  const {
    dashboardStore,
    selectors: {
      dimensions: { allDimensions },
    },
  } = getStateManagers();

  let dimensions: MetricsViewSpecDimensionV2[] | undefined = [];

  $: showTimeComparison = $dashboardStore?.showTimeComparison;
  $: selectedDimension = $dashboardStore?.selectedComparisonDimension;
  $: dimensions = $allDimensions;

  let searchText = "";

  function getLabelForDimension(dimension: string) {
    if (!dimensions) return dimension;
    const dimensionObj = dimensions.find((d) => d.name === dimension);
    return dimensionObj?.label || dimension;
  }

  /** compile the comparison options */
  $: options = (dimensions || []).map((d) => ({
    name: d.name,
    label: d.label || d.name,
  }));

  $: menuOptions = matchSorter(options, searchText, { keys: ["label"] });

  $: label = selectedDimension
    ? getLabelForDimension(selectedDimension)
    : NO_COMPARISON_LABEL;

  $: intermediateSelection = selectedDimension
    ? selectedDimension
    : NO_COMPARISON_LABEL;

  function enableComparison(type: string, name = "") {
    if (type === "time") {
      metricsExplorerStore.displayTimeComparison(exploreName, true);
    } else {
      // Temporary until these are not mutually exclusive
      metricsExplorerStore.displayTimeComparison(exploreName, false);
      metricsExplorerStore.setComparisonDimension(exploreName, name);
    }
  }

  function disableAllComparisons() {
    metricsExplorerStore.disableAllComparisons(exploreName);
  }
</script>

<WithTogglableFloatingElement
  distance={8}
  alignment="start"
  let:toggleFloatingElement
  let:active
>
  <Tooltip distance={8} suppress={active}>
    {#if chipStyle}
      <Chip
        on:click={toggleFloatingElement}
        {label}
        {active}
        type="dimension"
        caret
      >
        <span class="font-bold truncate" slot="body"> {label}</span>
      </Chip>
    {:else}
      <Button type="text" on:click={toggleFloatingElement}>
        <div
          class="flex items-center gap-x-0.5 px-1.5 text-gray-700 hover:text-inherit"
        >
          <span class="font-normal">
            {showTimeComparison || selectedDimension ? "Broken down by" : ""}
            <span class="font-bold">{label}</span>
          </span>
        </div>
      </Button>
    {/if}
    <TooltipContent slot="tooltip-content" maxWidth="220px">
      Select a comparison for the dashboard
    </TooltipContent>
  </Tooltip>
  <Menu
    minWidth="280px"
    slot="floating-element"
    let:handleClose
    on:escape={handleClose}
    on:click-outside={handleClose}
    label="Comparison selector"
  >
    <div class="px-2 pb-2">
      <Search
        placeholder="Search Dimension"
        bind:value={searchText}
        showBorderOnFocus={false}
      />
    </div>
    <MenuItem
      focusOnMount={false}
      selected={!(showTimeComparison || selectedDimension)}
      on:before-select={() => {
        intermediateSelection = NO_COMPARISON_LABEL;
      }}
      on:select={() => {
        disableAllComparisons();
        handleClose();
      }}
    >
      <span class:font-bold={intermediateSelection === NO_COMPARISON_LABEL}>
        {NO_COMPARISON_LABEL}
      </span>
    </MenuItem>

    <div style:max-height="200px" class="overflow-y-auto">
      {#each menuOptions as option (option.name)}
        <MenuItem
          selected={option.name === intermediateSelection}
          on:before-select={() => {
            if (option.name) {
              intermediateSelection = option.name;
            }
          }}
          on:select={() => {
            enableComparison("dimension", option.name);
            handleClose();
          }}
        >
          <span class:font-bold={intermediateSelection === option.name}>
            {option.label}
          </span>
        </MenuItem>
      {/each}
    </div>
  </Menu>
</WithTogglableFloatingElement>
