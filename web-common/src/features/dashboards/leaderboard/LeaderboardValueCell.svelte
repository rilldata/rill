<script lang="ts">
  import { cubicOut as easing } from "svelte/easing";
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { LeaderboardItemData } from "./leaderboard-utils";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import { tweened } from "svelte/motion";
  import LongBarZigZag from "./LongBarZigZag.svelte";

  const valueTween = tweened(0, {
    duration: 200,
    easing,
  });

  const {
    selectors: {
      numberFormat: { activeMeasureFormatter },
      activeMeasure: { isSummableMeasure },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
    },
  } = getStateManagers();

  export let dimensionName: string;
  export let itemData: LeaderboardItemData;
  export let tableWidth: number;
  export let label: string;
  export let comparisonValue: number | null;

  let hovered = false;

  $: ({ dimensionValue: label, selectedIndex, pctOfTotal } = itemData);

  $: selected = selectedIndex >= 0;

  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);
  /** for summable measures, this is the value we use to calculate the bar % to fill */

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? $activeMeasureFormatter(comparisonValue)
      : undefined;
  $: showPreviousTimeValue = hovered && previousValueString !== undefined;
  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: renderedBarValue = $isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: color = excluded
    ? "ui-measure-bar-excluded"
    : selected
      ? "ui-measure-bar-included-selected"
      : "ui-measure-bar-included";

  $: valueTween.set(renderedBarValue);
</script>

<!-- NOTE: empty class leaderboard-label is used to locate this elt in e2e tests -->
<div
  class="relative size-full pl-2 flex flex-none items-center leaderboard-label"
  class:ui-copy={!atLeastOneActive}
  class:ui-copy-disabled={excluded}
  class:ui-copy-strong={!excluded && selected}
>
  <FormattedDataType value={label} truncate />

  <div
    class="{color} h-full absolute left-0 -z-10"
    style:width="{tableWidth * Math.min(1, $valueTween)}px"
  >
    {#if renderedBarValue > 1.001}
      <LongBarZigZag />
    {/if}
  </div>

  <div
    class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
  >
    <div class="flex items-baseline gap-x-1" in:fly={{ duration: 200, y: 4 }}>
      {#if showPreviousTimeValue}
        <span
          class="inline-block opacity-50"
          transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
        >
          {previousValueString}
          →
        </span>
      {/if}
    </div>
  </div>
</div>
