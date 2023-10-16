<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. read the existing active comparison from somewhere.
3. update comparisons on user interactions
-->
<script lang="ts">
  import { matchSorter } from "match-sorter";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { NO_COMPARISON_LABEL } from "@rilldata/web-common/lib/time/config";
  import SelectorButton from "./SelectorButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import ClockCircle from "@rilldata/web-common/components/icons/ClockCircle.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricViewName;
  export let chipStyle = false;

  const TIME = "Time";

  let dimensions = [];

  $: dashboardStore = useDashboardStore(metricViewName);
  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: showTimeComparison = $dashboardStore?.showTimeComparison;
  $: selectedDimension = $dashboardStore?.selectedComparisonDimension;
  $: dimensions = $metaQuery?.data?.dimensions;

  let searchText = "";

  function getLabelForDimension(dimension: string) {
    return dimensions.find((d) => d.name === dimension)?.label;
  }

  /** compile the comparison options */
  $: options = dimensions.map((d) => ({
    name: d.name,
    label: d.label,
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
      <Chip
        on:click={() => {
          toggleFloatingElement();
        }}
        {label}
      >
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
      <SelectorButton
        {active}
        on:click={() => {
          toggleFloatingElement();
        }}
      >
        <div class="flex items-center gap-x-3">
          <span class="ui-copy-icon"><Compare size="16px" /></span>

          <span style:transform="translateY(-1px)" class="font-normal">
            {showTimeComparison || selectedDimension ? "Comparing by" : ""}
            <span class="font-bold">{label}</span>
          </span>
        </div>
      </SelectorButton>
    {/if}
    <TooltipContent slot="tooltip-content" maxWidth="220px">
      Select a comparison for the dashboard
    </TooltipContent>
  </Tooltip>
  <Menu
    minWidth="280px"
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    label="Comparison selector"
  >
    <Search placeholder="Search Dimension" bind:value={searchText} />
    <MenuItem
      focusOnMount={false}
      selected={!(showTimeComparison || selectedDimension)}
      on:before-select={() => {
        intermediateSelection = NO_COMPARISON_LABEL;
      }}
      on:select={() => {
        disableAllComparisons();
        toggleFloatingElement();
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
        toggleFloatingElement();
      }}
    >
      <span class:font-bold={intermediateSelection === TIME}> {TIME} </span>
      <span slot="right"><ClockCircle size="16px" /></span>
    </MenuItem>
    <Divider marginTop={0.5} marginBottom={0.5} />

    <div style:max-height="200px" class="overflow-y-auto">
      {#each menuOptions as option}
        <MenuItem
          selected={option.name === intermediateSelection}
          on:before-select={() => {
            intermediateSelection = option.name;
          }}
          on:select={() => {
            enableComparison("dimension", option.name);
            toggleFloatingElement();
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
