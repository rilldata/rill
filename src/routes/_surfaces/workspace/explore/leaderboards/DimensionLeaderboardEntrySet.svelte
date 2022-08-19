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
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;

  const dispatch = createEventDispatcher();
</script>

{#each values as { label, value, formattedValue } (label)}
  {@const active = activeValues.find((value) => value === label)}
  <div>
    <DimensionLeaderboardEntry
      measureValue={value}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      on:click={() => {
        dispatch("select-item", { label, isActive: active });
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
          filter on <span class="italic">{label}</span>
        {:else}
          remove filter for <span class="italic">{label}</span>
        {/if}
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
