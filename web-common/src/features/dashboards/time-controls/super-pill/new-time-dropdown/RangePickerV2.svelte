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
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { type V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
  import { bucketTimeRanges } from "../../time-range-store";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import TimeRangeMenuItem from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeMenuItem.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";

  import {
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";
  import { InfoIcon } from "lucide-svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";

  export let timeRanges: V1ExploreTimeRange[];
  export let selected: string | undefined;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: string;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: string, syntax?: boolean) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let showPanel = false;

  $: rangeBuckets = bucketTimeRanges(timeRanges, "rill-TD");

  $: colloquialGrain = getColloquialGrain(selected);

  // $: includesCurrentPeriod = interval.end.diff(maxDate).milliseconds > 0;

  let parsedTime: RillTime | undefined = undefined;

  $: isComplete = parsedTime?.isComplete ?? false;

  $: if (selected) {
    try {
      parsedTime = parseRillTime(selected);
    } catch {
      // no op
    }
  }

  $: selectedMeta = selected?.startsWith("P")
    ? LATEST_WINDOW_TIME_RANGES[selected]
    : selected?.startsWith("rill")
      ? (PERIOD_TO_DATE_RANGES[selected] ??
        PREVIOUS_COMPLETE_DATE_RANGES[selected])
      : undefined;

  function handleRangeSelect(range: string, syntax = false) {
    onSelectRange(range, syntax);

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }

  function getColloquialGrain(range: string | undefined) {
    // find the first character that matches H M D W Q Y in upper or lower case
    const grain = range?.match(/[HMDWQY]/i)?.[0];
    if (range === "CUSTOM") return undefined;
    return grain;
  }
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
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

  <DropdownMenu.Content
    align="start"
    class="p-0 w-fit overflow-hidden flex flex-col"
  >
    <TimeRangeSearch
      width={showPanel ? 500 : 240}
      bind:this={searchComponent}
      {context}
      onSelectRange={(range, syntax) => {
        open = false;
        onSelectRange(range, syntax);
      }}
    />

    <div class="flex w-fit max-h-fit" style:height="600px">
      <div
        class="flex flex-col w-60 overflow-y-auto overflow-x-hidden flex-none pt-1"
      >
        <div class="overflow-x-hidden">
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

          {#each rangeBuckets.customTimeRanges as range, i (i)}
            <TimeRangeMenuItem
              {range}
              selected={selected === range.range}
              onClick={handleRangeSelect}
            />
          {/each}

          {#if rangeBuckets.customTimeRanges.length}
            <DropdownMenu.Separator />
          {/if}

          {#each rangeBuckets.ranges as ranges, i (i)}
            {#each ranges as range, i (i)}
              <TimeRangeMenuItem
                {range}
                selected={selected === range.meta?.rillSyntax}
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
          {/if}
        </div>

        <DropdownMenu.Separator />

        <DropdownMenu.Item
          on:click={() => {
            showPanel = !showPanel;
          }}
        >
          <span class:font-bold={selected === "Custom"}> Custom...</span>
        </DropdownMenu.Item>

        {#if parsedTime}
          <DropdownMenu.Separator />
          <div class="flex justify-between items-center py-2 px-3">
            <span class="flex gap-x-1 items-center">
              <span>Include latest partial period</span>
              <Tooltip distance={8}>
                <InfoIcon size="12px" class="text-gray-500" />
                <TooltipContent slot="tooltip-content">
                  <div class="flex flex-col gap-y-1 items-center">
                    <span>
                      Show all available data, even if the period is not
                      complete.
                    </span>
                  </div>
                </TooltipContent>
              </Tooltip>
            </span>

            <Switch
              id="Show comparison"
              checked={!isComplete}
              small
              on:click={() => {
                if (!isComplete) {
                  if (selectedMeta) return;
                  const updatedString = `${selected}, latest/${colloquialGrain}`;
                  console.log({ updatedString });
                  onSelectRange(updatedString, true);
                } else if (selected) {
                  console.log({ selected });
                  onSelectRange(selected?.split(",")[0], true);
                }
              }}
            />
          </div>
        {/if}
      </div>

      {#if showPanel}
        <div
          class="bg-slate-50 border-l w-[260px] h-full flex flex-col justify-between"
        >
          <CalendarPlusDateInput
            {firstVisibleMonth}
            {interval}
            {zone}
            {maxDate}
            {minDate}
            applyRange={applyCustomRange}
            closeMenu={() => (open = false)}
          />

          <!-- </div> -->
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

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
