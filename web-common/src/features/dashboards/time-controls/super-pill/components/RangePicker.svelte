<script lang="ts" context="module">
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";

  const exampleSearches = [
    "-45M",
    "-30D",
    "-1Y",
    "1/1/2025",
    "-52W",
    "0M, latest : |d|",
  ];
</script>

<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
  import { DateTime, Interval } from "luxon";
  import type { ISODurationString, NamedRange } from "../../new-time-controls";
  import {
    ALL_TIME_RANGE_ALIAS,
    getRangeLabel,
    RILL_TO_LABEL,
  } from "../../new-time-controls";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";
  import RangeDisplay from "./RangeDisplay.svelte";
  import SyntaxElement from "./SyntaxElement.svelte";
  import { type V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
  import { bucketTimeRanges } from "../../time-range-store";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import TimeRangeMenuItem from "./TimeRangeMenuItem.svelte";
  import Switch from "@rilldata/web-common/components/button/Switch.svelte";
  import { Clock } from "lucide-svelte";
  import {
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import Timestamp from "./Timestamp.svelte";

  export let timeRanges: V1ExploreTimeRange[];
  export let selected: string | undefined;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: string;
  export let context: string;
  export let minDate: DateTime<true>;
  export let maxDate: DateTime<true>;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: string, syntax?: boolean) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  const latestNSearches = localStorageStore(`${context}-recent-searches`, [
    "-45M",
    "-32D",
    "-1Y",
    "-2Q, latest/Q",
  ]);

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let showSelector = false;
  let searchValue = "";
  let searchElement: HTMLInputElement;
  let allTimeAllowed = true;

  $: rangeBuckets = bucketTimeRanges(timeRanges, "rill-TD");

  function updateSearch(value: string) {
    searchValue = value;
    searchElement.focus();
  }

  function handleRangeSelect(range: string, syntax = false) {
    onSelectRange(range, syntax);

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }

  // All of the below is hacky for the sake of development

  function getColloquialGrain(range: string | undefined) {
    // find the first character that matches H M D W Q Y in upper or lower case
    const grain = range?.match(/[HMDWQY]/i)?.[0];
    if (range === "CUSTOM") return undefined;
    return grain;
  }

  $: colloquialGrain = getColloquialGrain(selected);

  $: grainPhrase = getGrainPhrase(colloquialGrain);

  $: meta = selected?.startsWith("P")
    ? LATEST_WINDOW_TIME_RANGES[selected]
    : selected?.startsWith("rill")
      ? (PERIOD_TO_DATE_RANGES[selected] ??
        PREVIOUS_COMPLETE_DATE_RANGES[selected])
      : undefined;

  function getGrainPhrase(grain: string | undefined) {
    switch (grain?.toUpperCase()) {
      case "H":
        return "this hour";
      case "M":
        return "this month";
      case "D":
        return "today";
      case "W":
        return "this week";
      case "Q":
        return "this quarter";
      case "Y":
        return "this year";
      default:
        return "current period";
    }
  }

  $: includesCurrentPeriod = interval.end.diff(maxDate).milliseconds > 0;
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
    showSelector = selected === "CUSTOM";
  }}
  closeOnItemClick={false}
  typeahead={false}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex gap-x-1"
      aria-label="Select time range"
    >
      {#if selected}
        <b class="mr-1 line-clamp-1 flex-none">{getRangeLabel(selected)}</b>
      {/if}
      {#if interval.isValid}
        <RangeDisplay {interval} {grain} />
      {/if}
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="p-0 overflow-hidden max-w-[480px]">
    <div class="border-b h-fit pt-2.5 flex p-3 py-0 flex-col w-full">
      <form
        on:submit={() => {
          latestNSearches.update((searches) => {
            return Array.from(new Set([searchValue, ...searches].slice(0, 20)));
          });
          onSelectRange(searchValue, true);
          open = false;
          searchValue = "";
        }}
      >
        <span class="mr-1 flex-none">
          <Clock size={15} />
        </span>
        <input
          placeholder="Search"
          type="text"
          class="h-7 border w-full"
          bind:this={searchElement}
          bind:value={searchValue}
        />
      </form>

      <div class="flex gap-x-2 size-full overflow-x-auto py-2.5">
        {#each $latestNSearches as search, i (i)}
          <SyntaxElement range={search} onClick={updateSearch} />
        {/each}
      </div>
    </div>
    <div class="flex">
      <div class="flex flex-col w-56 p-1">
        {#if showDefaultItem && defaultTimeRange}
          <DropdownMenu.Item
            on:click={() => {
              handleRangeSelect(defaultTimeRange);
            }}
          >
            <div class:font-bold={selected === defaultTimeRange}>
              Last {humaniseISODuration(defaultTimeRange)}
            </div>
          </DropdownMenu.Item>

          <DropdownMenu.Separator />
        {/if}

        {#each rangeBuckets.ranges as ranges, i (i)}
          {#each ranges as range, i (i)}
            <TimeRangeMenuItem
              selected={selected === range.range}
              {range}
              onClick={handleRangeSelect}
            />
          {/each}
          {#if ranges.length}
            <DropdownMenu.Separator />
          {/if}
        {/each}

        {#if allTimeAllowed}
          <DropdownMenu.Item
            on:click={() => {
              handleRangeSelect(ALL_TIME_RANGE_ALIAS);
            }}
          >
            <span class:font-bold={selected === ALL_TIME_RANGE_ALIAS}>
              {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
            </span>
          </DropdownMenu.Item>

          <DropdownMenu.Separator />
        {/if}

        <DropdownMenu.Item
          on:click={() => (showSelector = !showSelector)}
          data-range="custom"
        >
          <span class:font-bold={selected === "CUSTOM"}> Calendar </span>
        </DropdownMenu.Item>

        <DropdownMenu.Separator />

        <div class="flex justify-between items-center py-1 px-2">
          <span>Include {grainPhrase}</span>
          <Switch
            id="Show comparison"
            checked={includesCurrentPeriod}
            on:click={() => {
              if (includesCurrentPeriod) {
                const updatedString = `${meta.rillSyntax}, latest/${colloquialGrain}`;
                console.log({ updatedString });
                onSelectRange(updatedString, true);
              } else {
                onSelectRange(meta.rillSyntax, true);
              }
            }}
          />
        </div>
      </div>

      <div class="bg-slate-50 border-l flex flex-col w-64 p-2 py-1">
        {#if showSelector}
          <CalendarPlusDateInput
            {firstVisibleMonth}
            {interval}
            {zone}
            {maxDate}
            {minDate}
            applyRange={applyCustomRange}
            closeMenu={() => (open = false)}
          />
        {:else}
          <div class="size-full p-2 flex-col gap-y-5 flex">
            <div class="flex flex-col gap-y-3">
              <span class="text-gray-500 text-xs">Example searches</span>
              <div class="flex gap-x-2 flex-wrap gap-y-1">
                {#each exampleSearches as search, i (i)}
                  <SyntaxElement range={search} onClick={updateSearch} />
                {/each}
              </div>
            </div>

            <div class="flex flex-col gap-y-3">
              <span class="text-gray-500 text-xs">Custom ranges</span>
              <div class="flex gap-x-2 flex-wrap gap-y-1">
                {#each rangeBuckets.customTimeRanges as range, i (i)}
                  <SyntaxElement range={range.range} onClick={updateSearch} />
                {:else}
                  <span class="text-gray-500"> No ranges available </span>
                {/each}
              </div>
            </div>

            <div class="flex flex-col gap-y-3">
              <span class="text-gray-500 text-xs">Available time range</span>

              <div class="flex flex-col gap-y-1">
                {#if minDate}
                  <div class="flex justify-between items-center">
                    <SyntaxElement range="earliest" onClick={updateSearch} />
                    <Timestamp date={minDate} {zone} />
                  </div>
                {/if}

                {#if maxDate}
                  <div class="flex justify-between items-center">
                    <SyntaxElement range="latest" onClick={updateSearch} />
                    <Timestamp date={maxDate} {zone} />
                  </div>
                {/if}

                <div class="flex justify-between items-center">
                  <SyntaxElement range="now" onClick={updateSearch} />
                  <Timestamp {zone} />
                </div>
              </div>
            </div>
            <a href="https://www.rilldata.com" class="mt-auto">
              Syntax documentation
            </a>
          </div>
        {/if}
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
  }

  form {
    @apply overflow-hidden;
    @apply flex justify-center gap-x-1 items-center pl-2 pr-0.5;
    @apply bg-background justify-center;
    @apply border border-gray-300 rounded-[2px];
    @apply cursor-pointer;
    @apply h-7 w-full truncate;
  }

  form:focus-within {
    @apply border-primary-500;
  }

  input,
  .multiline-input {
    @apply p-0 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }
</style>
