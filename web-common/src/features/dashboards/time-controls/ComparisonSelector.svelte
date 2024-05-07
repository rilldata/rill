<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. read the existing active comparison
3. update comparisons on user interactions
-->
<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ClockCircle from "@rilldata/web-common/components/icons/ClockCircle.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { NO_COMPARISON_LABEL } from "@rilldata/web-common/lib/time/config";
  import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { matchSorter } from "match-sorter";
  import SelectorButton from "./SelectorButton.svelte";

  export let metricViewName;
  export let chipStyle = false;

  const TIME = "Time";

  let dimensions: MetricsViewSpecDimensionV2[] | undefined = [];

  $: dashboardStore = useDashboardStore(metricViewName);
  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);

  $: showTimeComparison = $dashboardStore?.showTimeComparison;
  $: selectedDimension = $dashboardStore?.selectedComparisonDimension;
  $: dimensions = $metricsView?.data?.dimensions;

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
    : showTimeComparison
      ? TIME
      : NO_COMPARISON_LABEL;

  $: intermediateSelection = selectedDimension
    ? selectedDimension
    : showTimeComparison
      ? TIME
      : NO_COMPARISON_LABEL;

  function enableComparison(type: string, name = "") {
    if (type === "time") {
      metricsExplorerStore.displayTimeComparison(metricViewName, true);
    } else {
      metricsExplorerStore.setComparisonDimension(metricViewName, name);
    }
  }

  function disableAllComparisons() {
    metricsExplorerStore.disableAllComparisons(metricViewName);
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
      <Chip on:click={toggleFloatingElement} {label} {active} outline={true}>
        <div slot="body" class="flex gap-x-2">
          <div
            class="font-bold text-ellipsis overflow-hidden whitespace-nowrap ml-2"
          >
            {label}
          </div>

          <div class="flex items-center">
            <IconSpaceFixer pullRight>
              <div class="transition-transform" class:-rotate-180={active}>
                <CaretDownIcon size="14px" />
              </div>
            </IconSpaceFixer>
          </div>
        </div>
      </Chip>
    {:else}
      <Button type="text" on:click={toggleFloatingElement}>
        <div
          class="flex items-center gap-x-0.5 px-1.5 text-gray-700 hover:text-inherit"
        >
          <span class="font-normal">
            {showTimeComparison || selectedDimension ? "Comparing by" : ""}
            <span class="font-bold">{label}</span>
          </span>
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon />
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
    <Divider marginTop={0.5} marginBottom={0.5} />
    <MenuItem
      selected={showTimeComparison}
      on:before-select={() => {
        intermediateSelection = TIME;
      }}
      on:select={() => {
        enableComparison("time");
        handleClose();
      }}
    >
      <span class:font-bold={intermediateSelection === TIME}> {TIME} </span>
      <span slot="right"><ClockCircle size="16px" /></span>
    </MenuItem>
    <Divider marginTop={0.5} marginBottom={0.5} />

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
