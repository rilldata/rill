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
  export let filterMode: boolean;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;

  const dispatch = createEventDispatcher();
</script>

{#each values as { label, value, formattedValue } (label)}
  {@const active = activeValues.findIndex((value) => value === label) >= 0}
  <div>
    <DimensionLeaderboardEntry
      measureValue={value}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
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
        {formattedValue || value || "∅"}
      </svelte:fragment>
      <svelte:fragment slot="tooltip">
        {#if !active}
          {#if atLeastOneActive}
            <div>
              filter {filterMode ? "out" : "on"}
              <span class="italic">{label}</span>
            </div>
          {:else}
            <div>include <span class="italic">{label}</span> in filter</div>
          {/if}
        {:else}
          <div>remove <span class="italic">{label}</span> from filter</div>
        {/if}
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
