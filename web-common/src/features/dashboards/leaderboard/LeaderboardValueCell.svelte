<script lang="ts">
  import { cubicOut as easing } from "svelte/easing";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { getStateManagers } from "../state-managers/state-managers";
  import { tweened } from "svelte/motion";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import { slide } from "svelte/transition";

  const valueTween = tweened(0, {
    duration: 200,
    easing,
  });

  const {
    selectors: {
      activeMeasure: { isSummableMeasure },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
    },
  } = getStateManagers();

  export let dimensionName: string;
  export let tableWidth: number;
  export let label: string;
  export let previousValueString: string | null = null;
  export let pctOfTotal: number | null;
  export let selected: boolean;
  export let hovered: boolean;

  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  /** for summable measures, this is the value we use to calculate the bar % to fill */
  $: renderedBarValue = $isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: color = excluded
    ? "rgb(243 244 246)"
    : selected
      ? "var(--color-primary-200)"
      : "var(--color-primary-100)";

  $: valueTween.set(renderedBarValue).catch(console.error);
</script>

<!-- NOTE: empty class leaderboard-label is used to locate this elt in e2e tests -->
<div
  class="relative size-full pl-2 flex flex-none justify-between items-center leaderboard-label"
  class:ui-copy={!atLeastOneActive}
  class:ui-copy-disabled={excluded}
  class:ui-copy-strong={!excluded && selected}
>
  <FormattedDataType value={label} truncate />

  {#if previousValueString && hovered}
    <span
      class="opacity-50 whitespace-nowrap font-normal"
      transition:slide={{ axis: "x", duration: 200 }}
    >
      {previousValueString} →
    </span>
  {/if}

  <div
    style:width="{tableWidth}px"
    class="h-full absolute left-0 -z-10"
    style:background="linear-gradient(to right, {color}
    {renderedBarValue * 100}%, hsl(var(--background)) {renderedBarValue * 100}%
    100%)"
  >
    {#if renderedBarValue > 1.001}
      <LongBarZigZag />
    {/if}
  </div>
</div>
