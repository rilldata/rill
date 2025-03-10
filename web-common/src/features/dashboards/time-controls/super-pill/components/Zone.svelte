<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
    formatIANAs,
    allTimeZones,
  } from "@rilldata/web-common/lib/time/timezone";
  import ZoneDisplay from "./ZoneDisplay.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import type { DateTime } from "luxon";

  const browserIANA = getLocalIANA();

  // watermark indicates the latest reference point in the dashboard
  export let watermark: DateTime;
  export let availableTimeZones: string[];
  export let activeTimeZone: string;
  export let context: string;
  export let lockTimeZone = false;
  export let onSelectTimeZone: (timeZone: string) => void;

  const recents = localStorageStore<string[]>(`${context}-recent-zones`, []);

  let open = false;
  let searchValue = "";

  $: ianaMap = formatIANAs(allTimeZones, watermark);

  $: pinnedTimeZones = formatIANAs([...availableTimeZones, "UTC"], watermark);

  $: filteredPinnedTimeZones = filterTimeZones(pinnedTimeZones, searchValue);

  $: filteredTimeZones = filterTimeZones(ianaMap, searchValue);

  $: recentIANAs = $recents;

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

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      class="flex items-center gap-x-1"
      aria-label="Timezone selector"
      title={!availableTimeZones.length ? "No timezones configured" : ""}
      disabled={lockTimeZone}
    >
      {getAbbreviationForIANA(watermark, activeTimeZone)}
      {#if !lockTimeZone}
        <span class="flex-none transition-transform" class:-rotate-180={open}>
          <CaretDownIcon />
        </span>
      {/if}
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-80 ">
    <div class="p-1.5 pb-1 flex items-center gap-x-2">
      <Search bind:value={searchValue} autofocus={false} />
    </div>

    {#if !pinnedTimeZones.has(activeTimeZone) && !recentIANAs.includes(activeTimeZone)}
      <DropdownMenu.Group>
        {@const formatted = ianaMap.get(activeTimeZone)}
        {#if formatted}
          <DropdownMenu.Item>
            <ZoneDisplay
              abbreviation={formatted.abbreviation}
              offset={formatted.offset}
              iana={activeTimeZone}
            />
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Group>
      <DropdownMenu.Separator />
    {/if}

    <DropdownMenu.Group>
      {#each filteredPinnedTimeZones as [iana, { offset, abbreviation }] (iana)}
        <DropdownMenu.CheckboxItem
          checkRight
          checked={activeTimeZone === iana}
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
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Group>

    {#if !searchValue && recentIANAs.length}
      <DropdownMenu.Separator />

      <DropdownMenu.Group>
        <div class="flex justify-between pr-2">
          <DropdownMenu.Label>Recent</DropdownMenu.Label>
          {#if recentIANAs.length}
            <button
              on:click={() => {
                recents.set([]);
              }}
              class="text-[10px] text-gray-500">Clear recents</button
            >
          {/if}
        </div>

        {#each recentIANAs as iana, i (i)}
          {@const formatted = ianaMap.get(iana)}
          {#if formatted && !availableTimeZones.includes(iana) && iana !== browserIANA}
            <DropdownMenu.CheckboxItem
              checkRight
              checked={activeTimeZone === iana}
              on:click={() => {
                onSelectTimeZone(iana);
              }}
            >
              <ZoneDisplay
                abbreviation={formatted.abbreviation}
                offset={formatted.offset}
                {iana}
              />
            </DropdownMenu.CheckboxItem>
          {/if}
        {/each}
      </DropdownMenu.Group>
    {/if}

    {#if searchValue}
      <DropdownMenu.Separator />

      <DropdownMenu.Group class="max-h-72 overflow-y-auto">
        <DropdownMenu.Label
          class="sticky top-0 bg-gradient-to-b z-10 from-white from-75% to-transparent"
        >
          Search Results
        </DropdownMenu.Label>

        {#each filteredTimeZones as [iana, { abbreviation, offset }], i (i)}
          <DropdownMenu.CheckboxItem
            checkRight
            on:click={() => {
              onSelectTimeZone(iana);
              recents.set(Array.from(new Set([iana, ...$recents])).slice(0, 5));
            }}
          >
            <ZoneDisplay {iana} {offset} {abbreviation} />
          </DropdownMenu.CheckboxItem>
        {:else}
          <DropdownMenu.Group>
            <p class="pt-0 pb-2 text-gray-500 text-center">No options found</p>
          </DropdownMenu.Group>
        {/each}
      </DropdownMenu.Group>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
