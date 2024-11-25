<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { NO_COMPARISON_LABEL } from "@rilldata/web-common/lib/time/config";
  import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
  import { matchSorter } from "match-sorter";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  export let exploreName: string;

  const {
    dashboardStore,
    selectors: {
      dimensions: { allDimensions },
    },
  } = getStateManagers();

  let dimensions: MetricsViewSpecDimensionV2[] | undefined = [];
  let searchText = "";
  let open = false;

  $: ({ showTimeComparison, selectedComparisonDimension } = $dashboardStore);

  $: dimensions = $allDimensions;

  /** compile the comparison options */
  $: options = (dimensions || []).map((d) => ({
    name: d.name,
    label: d.displayName || d.name,
  }));

  $: menuOptions = matchSorter(options, searchText, { keys: ["label"] });

  $: label = selectedComparisonDimension
    ? getLabelForDimension(selectedComparisonDimension)
    : NO_COMPARISON_LABEL;

  function getLabelForDimension(dimension: string) {
    if (!dimensions) return dimension;
    const dimensionObj = dimensions.find((d) => d.name === dimension);
    return dimensionObj?.displayName || dimension;
  }

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

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger>
    <Tooltip distance={8} suppress={open}>
      <Chip
        label="Select a comparison dimension"
        active={open}
        type="dimension"
        caret
      >
        <span class="font-bold truncate" slot="body">{label}</span>
      </Chip>

      <TooltipContent slot="tooltip-content" maxWidth="220px">
        Select a comparison for the dashboard
      </TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    <div class="p-2">
      <Search
        placeholder="Search Dimension"
        bind:value={searchText}
        showBorderOnFocus={false}
      />
    </div>
    <DropdownMenu.Item on:click={disableAllComparisons}>
      <span
        class:font-bold={!selectedComparisonDimension && !showTimeComparison}
      >
        {NO_COMPARISON_LABEL}
      </span>
    </DropdownMenu.Item>

    <div style:max-height="200px" class="overflow-y-auto">
      {#each menuOptions as option (option.name)}
        <DropdownMenu.Item
          on:click={() => {
            enableComparison("dimension", option.name);
          }}
        >
          <span class:font-bold={selectedComparisonDimension === option.name}>
            {option.label}
          </span>
        </DropdownMenu.Item>
      {/each}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
