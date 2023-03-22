<!-- @component
Creates a set of DimensionLeaderboardEntry components. This component makes it easy
to stitch together  chunks of a list. For instance, we can have:
leaderboard values above the fold
divider
leaderboard values not visible but selected
divider
see more button
-->
<script lang="ts">
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { PerRangeFormatter } from "@rilldata/web-common/lib/number-formatting/strategies/per-range";
  import { createEventDispatcher } from "svelte";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";

  export let values;
  export let comparisonValues;
  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;

  const { shiftClickAction } = createShiftClickAction();

  const dispatch = createEventDispatcher();
  let renderValues = [];

  $: renderValues = values.map((v, i) => {
    const active = activeValues.findIndex((value) => value === v.label) >= 0;
    // pray to god
    const comparisonValue = comparisonValues?.[i]?.value;

    // Super important special case: if there is not at least one "active" (selected) value,
    // we need to set *all* items to be included, because by default if a user has not
    // selected any values, we assume they want all values included in all calculations.
    const excluded = atLeastOneActive
      ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
      : false;

    return { ...v, active, excluded, comparisonValue };
  });

  const formatPercentage = (value) => {
    if (Math.abs(value * 100) < 0.1) {
      return `0%`;
    }
    const factory = new PerRangeFormatter([], {
      strategy: "perRange",
      rangeSpecs: [
        {
          minMag: -2,
          supMag: 3,
          maxDigitsRight: 1,
          baseMagnitude: 0,
          padWithInsignificantZeros: false,
        },
      ],
      defaultMaxDigitsRight: 0,
      numberKind: NumberKind.PERCENT,
    });
    console.log("formatting", value, factory.partsFormat(value));
    return factory.partsFormat(value);
  };
</script>

{#each renderValues as { label, value, __formatted_value, active, excluded, comparisonValue } (label)}
  {@const percDiff =
    comparisonValue && value && (value - comparisonValue) / comparisonValue}
  {@const diffIsPositive = percDiff >= 0}
  {@const diffParts = formatPercentage(percDiff)}
  <div
    use:shiftClickAction
    on:click={() => {
      dispatch("select-item", {
        label,
      });
    }}
    on:keydown
    on:shift-click={async () => {
      await navigator.clipboard.writeText(label);
      let truncatedLabel = label?.toString();
      if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
        truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
      }
      notifications.send({
        message: `copied dimension value "${truncatedLabel}" to clipboard`,
      });
    }}
  >
    <DimensionLeaderboardEntry
      measureValue={value}
      {comparisonValue}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      {excluded}
    >
      <svelte:fragment slot="label">
        {label}
      </svelte:fragment>
      <svelte:fragment slot="right">
        {__formatted_value || value || "∅"}
      </svelte:fragment>
      <div
        slot="context"
        class:text-red-500={!diffIsPositive}
        class:text-gray-700={diffIsPositive}
        style:font-weight={diffIsPositive ? "normal" : "medium"}
      >
        {#if percDiff !== undefined}
          {diffParts?.neg || ""}{diffParts?.int || ""}<span class="opacity-50"
            >{diffParts?.percent || ""}</span
          >
        {/if}
      </div>
      <svelte:fragment slot="tooltip">
        <TooltipTitle>
          <svelte:fragment slot="name">
            {label}
          </svelte:fragment>
        </TooltipTitle>

        <TooltipShortcutContainer>
          {#if atLeastOneActive}
            <div>
              {excluded ? "Include" : "Exclude"}
              this dimension value
            </div>
          {:else}
            <div class="text-ellipsis overflow-hidden whitespace-nowrap">
              Filter {filterExcludeMode ? "out" : "on"}
              this dimension value
            </div>
          {/if}
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
        <TooltipShortcutContainer>
          <div>
            <StackingWord key="shift">Copy</StackingWord>
            this dimension value to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        </TooltipShortcutContainer>
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
