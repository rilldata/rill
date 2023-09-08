<!-- @component 
This component needs to do the following:
1. display the set of available comparisons in the menu.
2. dispatch to TimeControl.svelte the selected comparison.
3. read the existing active comparison from somewhere.
-->
<script lang="ts">
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    NO_COMPARISON_LABEL,
    TIME_COMPARISON,
  } from "@rilldata/web-common/lib/time/config";
  import { createEventDispatcher } from "svelte";

  import SelectorButton from "./SelectorButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import type { MetricsViewDimension } from "@rilldata/web-common/runtime-client";

  const dispatch = createEventDispatcher();

  export let showComparison = true;
  export let selectedDimension;
  export let dimensions: MetricsViewDimension[];

  const TIME = "Time";
  $: comparisonOption = selectedDimension?.name;

  /** compile the comparison options */
  let options = dimensions.map((d) => ({
    name: d.name,
    label: d.label,
  }));

  $: label = showComparison
    ? TIME_COMPARISON[comparisonOption]?.label
    : NO_COMPARISON_LABEL;

  $: intermediateSelection = showComparison
    ? comparisonOption
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

        <span class="font-normal">
          {showComparison ? "Comparing by" : ""}
          <span class="font-bold">{label}</span>
        </span>
      </div>
    </SelectorButton>
    <TooltipContent slot="tooltip-content" maxWidth="220px">
      Select a time range to compare to the selected time range
    </TooltipContent>
  </Tooltip>
  <Menu
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    label="Time comparison selector"
  >
    <MenuItem
      selected={!showComparison}
      on:before-select={() => {
        intermediateSelection = NO_COMPARISON_LABEL;
      }}
      on:select={() => {
        dispatch("disable-comparison");
        toggleFloatingElement();
      }}
    >
      <span class:font-bold={intermediateSelection === NO_COMPARISON_LABEL}>
        {NO_COMPARISON_LABEL}
      </span>
    </MenuItem>
    <Divider />
    <MenuItem
      selected={!showComparison}
      on:before-select={() => {
        intermediateSelection = TIME;
      }}
      on:select={() => {
        toggleFloatingElement();
      }}
    >
      <span class:font-bold={intermediateSelection === TIME}> {TIME} </span>
    </MenuItem>
    <Divider />

    {#each options as option}
      <MenuItem
        selected={option.name === intermediateSelection}
        on:before-select={() => {
          intermediateSelection = option.name;
        }}
        on:select={() => {
          // onCompareRangeSelect(option.name);
          toggleFloatingElement();
        }}
      >
        <span class:font-bold={intermediateSelection === option.name}>
          {option.label}
        </span>
      </MenuItem>
    {/each}
  </Menu>
</WithTogglableFloatingElement>
