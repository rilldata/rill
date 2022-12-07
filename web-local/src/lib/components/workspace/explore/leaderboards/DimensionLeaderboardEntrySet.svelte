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
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-local/lib/application-config";
  import { createEventDispatcher } from "svelte";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";

  import TooltipShortcutContainer from "../../../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../../../tooltip/TooltipTitle.svelte";

  import { createShiftClickAction } from "../../../../util/shift-click-action";
  import { notifications } from "../../../notifications";
  import Shortcut from "../../../tooltip/Shortcut.svelte";
  import StackingWord from "../../../tooltip/StackingWord.svelte";

  export let values;
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
  $: {
    renderValues = values.map((v) => {
      const active = activeValues.findIndex((value) => value === v.label) >= 0;

      // Super important special case: if there is not at least one "active" (selected) value,
      // we need to set *all* items to be included, because by default if a user has not
      // selected any values, we assume they want all values included in all calculations.
      const excluded = atLeastOneActive
        ? (filterExcludeMode && active) || (!filterExcludeMode && !active)
        : false;

      return { ...v, active, excluded };
    });
  }
</script>

{#each renderValues as { label, value, __formatted_value, active, excluded } (label)}
  <div
    use:shiftClickAction
    on:click={() => {
      dispatch("select-item", {
        label,
      });
    }}
    on:shift-click={async () => {
      await navigator.clipboard.writeText(value);
      notifications.send({
        message: `copied column name "${value}" to clipboard`,
      });
    }}
  >
    <DimensionLeaderboardEntry
      measureValue={value}
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
      <svelte:fragment slot="tooltip">
        <TooltipTitle>
          <svelte:fragment slot="name">
            {label}
          </svelte:fragment>
        </TooltipTitle>

        <TooltipShortcutContainer>
          {#if atLeastOneActive}
            <div>
              {excluded ? "include" : "exclude"}
              <span class="italic">{label}</span>
              {excluded ? "in" : "from"} output
            </div>
          {:else}
            <div class="text-ellipsis overflow-hidden whitespace-nowrap">
              filter {filterExcludeMode ? "out" : "on"}
              <span class="italic"
                >{label?.length > TOOLTIP_STRING_LIMIT
                  ? label?.slice(0, TOOLTIP_STRING_LIMIT)?.trim() + "..."
                  : label}</span
              >
            </div>
          {/if}
          <Shortcut>Click</Shortcut>
        </TooltipShortcutContainer>
        <TooltipShortcutContainer>
          <div>
            <StackingWord key="shift">copy</StackingWord>
            {value} to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        </TooltipShortcutContainer>
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
