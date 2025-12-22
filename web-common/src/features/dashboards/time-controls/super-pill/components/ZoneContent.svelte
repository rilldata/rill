<script lang="ts">
  import {
    getLocalIANA,
    formatIANAs,
    allTimeZones,
  } from "@rilldata/web-common/lib/time/timezone";
  import ZoneDisplay from "./ZoneDisplay.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import type { DateTime } from "luxon";
  import { Check } from "lucide-svelte";

  const browserIANA = getLocalIANA();

  export let referencePoint: DateTime;
  export let availableTimeZones: string[];
  export let activeTimeZone: string;
  export let context: string;
  export let onSelectTimeZone: (timeZone: string) => void;

  const recents = localStorageStore<string[]>(`${context}-recent-zones`, []);

  let searchValue = "";

  $: ianaMap = formatIANAs(allTimeZones, referencePoint);

  $: pinnedTimeZones = formatIANAs(
    [...availableTimeZones, "UTC"],
    referencePoint,
  );

  $: filteredPinnedTimeZones = filterTimeZones(pinnedTimeZones, searchValue);

  $: filteredTimeZones = filterTimeZones(ianaMap, searchValue);

  $: recentIANAs = $recents;

  $: formatted = ianaMap.get(activeTimeZone);

  function filterTimeZones(
    zones: Map<string, { iana: string; offset: string; abbreviation: string }>,
    searchValue: string,
  ) {
    return new Map(
      Array.from(zones).filter(
        ([iana, { abbreviation }]) =>
          iana.toLowerCase().includes(searchValue.toLowerCase()) ||
          abbreviation?.toLowerCase().includes(searchValue.toLowerCase()),
      ),
    );
  }
</script>

<div class="p-1.5 pb-1 flex items-center gap-x-2">
  <Search bind:value={searchValue} autofocus={false} />
</div>

{#if !pinnedTimeZones.has(activeTimeZone) && !recentIANAs.includes(activeTimeZone)}
  <div class="group">
    {#if formatted}
      <button
        class="item"
        on:click={() => {
          onSelectTimeZone(activeTimeZone);
        }}
      >
        <ZoneDisplay
          abbreviation={formatted.abbreviation}
          offset={formatted.offset}
          iana={activeTimeZone}
        />
        <!-- {#if activeTimeZone === iana} -->
        <Check class="size-4" color="var(--color-gray-800)" />
        <!-- {/if} -->
      </button>
    {/if}
  </div>

  <div class="separator" />
{/if}

<div class="group">
  {#each filteredPinnedTimeZones as [iana, { offset, abbreviation }] (iana)}
    <button
      class="item"
      on:click={() => {
        onSelectTimeZone(iana);
      }}
    >
      <ZoneDisplay
        {abbreviation}
        {offset}
        isBrowserTime={iana === browserIANA}
        {iana}
      />
      <span class="flex flex-none h-3.5 w-3.5 items-center justify-center">
        {#if activeTimeZone === iana}
          <Check class="size-4" color="var(--color-gray-800)" />
        {/if}
      </span>
    </button>
  {/each}
</div>

{#if !searchValue && recentIANAs.length}
  <div class="separator" />
  <div class="group">
    <div class="flex justify-between pr-2 items-center">
      <h3>Recent</h3>
      {#if recentIANAs.length}
        <button
          class="text-[11px] text-gray-500 hover:bg-gray-100 p-1 rounded-sm h-fit"
          on:click={() => {
            recents.set([]);
          }}
        >
          Clear recents
        </button>
      {/if}
    </div>

    {#each recentIANAs as iana, i (i)}
      {@const formatted = ianaMap.get(iana)}
      {#if formatted && !availableTimeZones.includes(iana)}
        <button
          class="item"
          on:click={() => {
            onSelectTimeZone(iana);
          }}
        >
          <ZoneDisplay
            abbreviation={formatted.abbreviation}
            offset={formatted.offset}
            {iana}
          />
          <span class="flex flex-none h-3.5 w-3.5 items-center justify-center">
            {#if activeTimeZone === iana}
              <Check class="size-4" color="var(--color-gray-800)" />
            {/if}
          </span>
        </button>
      {/if}
    {/each}
  </div>
{/if}

{#if searchValue}
  <div class="separator" />
  <div class="group max-h-72 overflow-y-auto">
    <h3
      class="sticky top-0 bg-gradient-to-b z-10 from-surface from-75% to-transparent"
    >
      Search Results
    </h3>

    {#each filteredTimeZones as [iana, { abbreviation, offset }], i (i)}
      <button
        class="item"
        on:click={() => {
          onSelectTimeZone(iana);
          recents.set(Array.from(new Set([iana, ...$recents])).slice(0, 5));
        }}
      >
        <ZoneDisplay {iana} {offset} {abbreviation} />
        <span class="flex flex-none h-3.5 w-3.5 items-center justify-center">
          {#if activeTimeZone === iana}
            <Check class="size-4" color="var(--color-gray-800)" />
          {/if}
        </span>
      </button>
    {:else}
      <div>
        <p class="pt-0 pb-2 text-gray-500 text-center">No options found</p>
      </div>
    {/each}
  </div>
{/if}

<style lang="postcss">
  .item {
    @apply w-full relative justify-between flex cursor-pointer select-none items-start rounded-sm py-1.5 px-2 gap-x-2 text-xs outline-none;
  }

  .item:hover {
    @apply bg-accent text-accent-foreground;
  }

  .separator {
    @apply h-px w-full bg-gray-200 my-1;
  }

  h3 {
    @apply px-2 py-1.5 text-xs text-gray-500 font-semibold;
  }
</style>
