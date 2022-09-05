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
  import Check from "$lib/components/icons/Check.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
  import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
  import { isMac } from "$lib/util/os-detection";
  import { createEventDispatcher } from "svelte";
  import DimensionLeaderboardEntry from "./DimensionLeaderboardEntry.svelte";

  export let values;
  export let includeValues: Array<unknown>;
  export let excludeValues: Array<unknown>;
  export let isSummableMeasure: boolean;
  export let referenceValue;
  export let atLeastOneActive;
  export let loading = false;

  const dispatch = createEventDispatcher();
</script>

{#each values as { label, value, formattedValue } (label)}
  {@const included = includeValues.findIndex((value) => value === label) >= 0}
  {@const excluded = excludeValues.findIndex((value) => value === label) >= 0}
  {@const active = included || excluded}
  <div>
    <DimensionLeaderboardEntry
      measureValue={value}
      {loading}
      {isSummableMeasure}
      {referenceValue}
      {atLeastOneActive}
      {active}
      on:click={(evt) => {
        dispatch("select-item", { label, include: !evt.metaKey });
      }}
    >
      <svelte:fragment slot="label">
        <div class="flex flex-row">
          {#if included}<Check size="16px" />{:else if excluded}<div
              style="margin-top: 2px;"
            >
              <Close size="14px" />
            </div>{/if}
          {label}
        </div>
      </svelte:fragment>
      <svelte:fragment slot="right">
        {formattedValue || value || "∅"}
      </svelte:fragment>
      <svelte:fragment slot="tooltip">
        <TooltipShortcutContainer>
          {#if !active}
            <div>include filter on <span class="italic">{label}</span></div>
            <Shortcut>Click</Shortcut>
            <div>exclude filter on <span class="italic">{label}</span></div>
            <Shortcut>
              {#if isMac()}<span
                  style="
          font-family: var(--system);
          font-size: 11.5px;
        ">⌘</span
                >{:else}ctrl{/if} + Click</Shortcut
            >
          {:else}
            <div>remove filter for <span class="italic">{label}</span></div>
            <Shortcut>Click</Shortcut>
          {/if}
        </TooltipShortcutContainer>
      </svelte:fragment>
    </DimensionLeaderboardEntry>
  </div>
{/each}
