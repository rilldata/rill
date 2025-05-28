<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, Interval } from "luxon";
  import type { ISODurationString, NamedRange } from "../../new-time-controls";
  import {
    ALL_TIME_RANGE_ALIAS,
    getRangeLabel,
    RILL_TO_LABEL,
  } from "../../new-time-controls";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import {
    V1TimeGrain,
    type V1ExploreTimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import {
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";
  import { getTimeRangeOptionsByGrain } from "@rilldata/web-common/lib/time/defaults";
  import {
    getAllowedEndingGrains,
    getAllowedGrains,
    getGrainAliasFromString,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { TimeGrainOptions } from "@rilldata/web-common/lib/time/defaults";
  import TimeRangeOptionGroup from "./TimeRangeOptionGroup.svelte";
  import RangeDisplay from "../components/RangeDisplay.svelte";
  import InControl from "./InControl.svelte";
  import TimeRangeMenuItem from "../components/TimeRangeMenuItem.svelte";

  export let timeString: string | undefined;
  export let timeRanges: V1ExploreTimeRange[];
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: V1TimeGrain;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let allowCustomTimeRange = true;
  export let onSelectRange: (range: string, syntax?: boolean) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let showPanel = false;
  let filter = "";
  let parsedTime: RillTime | undefined = undefined;
  let showCustomSelector = false;

  let isShortHandSyntax = true;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      // no op
    }
  }

  $: selectedLabel = timeString && getRangeLabel(timeString);

  $: meta = parsedTime?.getMeta();

  $: canShowEndingControl = parsedTime && selectedLabel !== timeString;

  $: hasCustomSelected = !parsedTime;

  $: timeGrainOptions = getAllowedGrains(smallestTimeGrain);

  $: allowedEndingGrains = getAllowedEndingGrains(
    timeString,
    smallestTimeGrain,
  );

  $: allOptions = timeGrainOptions.map((grain) => {
    return getTimeRangeOptionsByGrain(grain, smallestTimeGrain);
  });

  $: groups = allOptions.reduce(
    (acc, options) => {
      acc.lastN.push(...options.lastN);
      acc.this.push(...options.this);
      acc.previous.push(...options.previous);

      return acc;
    },
    {
      lastN: [],
      this: [],
      previous: [],
      grainBy: [],
    } as TimeGrainOptions,
  );

  $: console.log({ groups });

  $: selectedMeta = timeString?.startsWith("P")
    ? LATEST_WINDOW_TIME_RANGES[timeString]
    : timeString?.startsWith("rill")
      ? (PERIOD_TO_DATE_RANGES[timeString] ??
        PREVIOUS_COMPLETE_DATE_RANGES[timeString])
      : undefined;

  function handleRangeSelect(range: string, syntax = false) {
    onSelectRange(range, syntax);

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }
</script>

<svelte:window
  on:keydown={(e) => {
    console.log(e);
    if (e.metaKey && e.key === "k") {
      open = !open;
    }
  }}
/>

<Popover.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
  }}
>
  <Popover.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex"
      aria-label="Select time range"
      data-state={open ? "open" : "closed"}
    >
      {#if timeString}
        <b class="mr-1 line-clamp-1 flex-none">{selectedLabel}</b>
      {/if}

      {#if interval.isValid}
        <RangeDisplay {interval} {grain} />
      {/if}

      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </Popover.Trigger>

  <Popover.Content
    align="start"
    class="p-0 w-fit overflow-hidden flex flex-col"
  >
    <TimeRangeSearch
      width={showCustomSelector ? 456 : 224}
      bind:this={searchComponent}
      {context}
      onSelectRange={(range, syntax) => {
        open = false;
        onSelectRange(range, syntax);
      }}
    />

    <div
      class="flex w-56 max-h-fit"
      class:!w-[456px]={showCustomSelector}
      style:height="500px"
    >
      <div
        class="flex flex-col w-56 overflow-y-auto overflow-x-hidden flex-none py-1"
      >
        <div class="overflow-x-hidden">
          {#if showDefaultItem && defaultTimeRange}
            <DropdownMenu.Item
              on:click={() => {
                handleRangeSelect(defaultTimeRange);
              }}
            >
              <div class:font-bold={timeString === defaultTimeRange}>
                Last {humaniseISODuration(defaultTimeRange)}
              </div>
            </DropdownMenu.Item>

            <div class="h-px w-full bg-gray-300" />
          {/if}

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={groups.lastN}
            onClick={handleRangeSelect}
          />

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={groups.this}
            onClick={handleRangeSelect}
          />

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={groups.previous}
            onClick={handleRangeSelect}
          />

          {#if allowCustomTimeRange}
            <TimeRangeOptionGroup
              {filter}
              timeString={hasCustomSelected ? "custom" : ""}
              options={[{ label: "Custom", string: "custom" }]}
              onClick={() => {
                showCustomSelector = !showCustomSelector;
              }}
            />
          {/if}

          {#if allTimeAllowed}
            <div class="w-full h-fit px-1">
              <button
                class="group h-7 px-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
                on:click={() => {
                  handleRangeSelect(ALL_TIME_RANGE_ALIAS);
                }}
              >
                <span class:font-bold={timeString === ALL_TIME_RANGE_ALIAS}>
                  {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
                </span>
              </button>
            </div>
          {/if}
        </div>
      </div>

      {#if showCustomSelector}
        <div class="bg-slate-50 border-l p-3 size-full">
          <CalendarPlusDateInput
            {firstVisibleMonth}
            {interval}
            {zone}
            {maxDate}
            {minDate}
            applyRange={applyCustomRange}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

{#if isShortHandSyntax && parsedTime}
  <InControl
    {parsedTime}
    {smallestTimeGrain}
    onSelectEnding={(grain, complete) => {
      console.log(grain, complete);
      // onSelectRange(grain);
    }}
  />
{/if}

<!-- <Popover.Root
  bind:open={calendarOpen}
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
  }}
>
  <Popover.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex"
      aria-label="Select time range"
      data-state={calendarOpen ? "open" : "closed"}
    >
      {#if interval.isValid}
        <RangeDisplay {interval} {grain} />
      {/if}

      <span
        class="flex-none transition-transform"
        class:-rotate-180={calendarOpen}
      >
        <CaretDownIcon />
      </span>
    </button>
  </Popover.Trigger>

  <Popover.Content align="start" class="w-fit overflow-hidden flex flex-col">
    <CalendarPlusDateInput
      {firstVisibleMonth}
      {interval}
      {zone}
      {maxDate}
      {minDate}
      applyRange={applyCustomRange}
      closeMenu={() => (calendarOpen = false)}
    />
  </Popover.Content>
</Popover.Root> -->

<style>
  /* The wrapper shrinks to the width of its content */
  .wrapper {
    display: inline-grid;
    grid-template-columns: 1fr; /* single column that both items share */
  }

  /* Vertical scroll container has an explicit width */
  .vertical-scroll {
    overflow-y: auto;
  }

  /* Horizontal container becomes a grid item and stretches to fill the column */
  .horizontal-scroll {
    overflow-x: auto;
    white-space: nowrap;

    /* No explicit width is set here */
  }
</style>
