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
  import { createEventDispatcher } from "svelte";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";

  export let values;
  export let activeValues: Array<unknown>;
  // false = include, true = exclude
  export let filterExcludeMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;

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
  <div>
    <DimensionLeaderboardEntry
      measureValue={value}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      {excluded}
      on:click={() => {
        dispatch("select-item", {
          label,
        });
      }}
    >
      <svelte:fragment slot="label">
        {label}
      </svelte:fragment>
      <svelte:fragment slot="right">
        {__formatted_value || value || "∅"}
      </svelte:fragment>
      <svelte:fragment slot="tooltip">
        {#if !active}
          {#if atLeastOneActive}
            <div>include <span class="italic">{label}</span> in filter</div>
          {:else}
            <div>
              filter {filterExcludeMode ? "out" : "on"}
              <span class="italic">{label}</span>
            </div>
          {/if}
        {:else}
          <div>remove <span class="italic">{label}</span> from filter</div>
        {/if}
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
