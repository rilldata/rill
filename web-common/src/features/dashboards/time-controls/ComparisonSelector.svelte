<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. dispatch to TimeControl.svelte the selected comparison.
3. read the existing active comparison from somewhere.
-->
<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { matchSorter } from "match-sorter";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { MetricsViewDimension } from "@rilldata/web-common/runtime-client";
  import { NO_COMPARISON_LABEL } from "@rilldata/web-common/lib/time/config";
  import SelectorButton from "./SelectorButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import ClockCircle from "@rilldata/web-common/components/icons/ClockCircle.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  const dispatch = createEventDispatcher();

  export let showTimeComparison = true;
  export let selectedDimension;
  export let dimensions: MetricsViewDimension[];

  const TIME = "Time";

  let searchText = "";

  function getLabelForDimension(dimension: string) {
    return dimensions.find((d) => d.name === dimension)?.label;
  }

  /** compile the comparison options */
  let options = dimensions.map((d) => ({
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
</script>

<WithTogglableFloatingElement
  distance={8}
  alignment="start"
  let:toggleFloatingElement
  let:active
>
  <Tooltip distance={8} suppress={active}>
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
        dispatch("disable-all-comparison");
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
        dispatch("enable-comparison", { type: "time" });
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
            dispatch("enable-comparison", {
              type: "dimension",
              name: option.name,
            });
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
