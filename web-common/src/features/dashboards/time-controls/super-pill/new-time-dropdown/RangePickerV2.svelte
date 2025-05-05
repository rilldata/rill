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
    TIME_GRAIN,
  } from "@rilldata/web-common/lib/time/config";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";

  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";
  import { getTimeRangeOptionsByGrain } from "@rilldata/web-common/lib/time/defaults";

  import {
    getAllowedGrains,
    V1TimeGrainToAlias,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { TimeGrainOptions } from "@rilldata/web-common/lib/time/defaults";

  export let timeRanges: V1ExploreTimeRange[];
  export let timeString: string | undefined;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let grain: V1TimeGrain;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: string, syntax?: boolean) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let showPanel = false;
  let filter = "";

  // let selectedTab = smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE;

  $: timeGrainOptions = getAllowedGrains(smallestTimeGrain);

  $: allOptions = timeGrainOptions.map((grain) => {
    return getTimeRangeOptionsByGrain(grain, smallestTimeGrain);
  });

  $: groups = allOptions.reduce(
    (acc, options) => {
      acc.lastN.push(...options.lastN);
      acc.this.push(...options.this);
      acc.previous.push(...options.previous);
      acc.grainBy.push(...options.grainBy);

      return acc;
    },
    {
      lastN: [],
      this: [],
      previous: [],
      grainBy: [],
    } as TimeGrainOptions,
  );

  //  $: filtered = allOptions.map((options) => {
  //     return options.filter((item) => {
  //       return item.label.toLowerCase().includes(searchComponent?.searchText);
  //     });
  //   });

  // $: rangeBuckets = {
  //   ranges: Object.values(groups).map((ranges) => {
  //     return ranges.map((range) => {
  //       // console.log({ range });
  //       return { range, parsed: parseRillTime(range) };
  //     });
  //   }),
  //   customTimeRanges: [],
  //   showDefaultItem: false,
  // };

  // $: rangeBuckets = bucketTimeRanges(
  //   timeRanges,
  //   defaultTimeRange,
  //   smallestTimeGrain,
  // );

  // const what = rangeBuckets.ranges[0];

  // what[0].

  let parsedTime: RillTime | undefined = undefined;

  $: isComplete = parsedTime?.isComplete ?? false;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      // no op
    }
  }

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

  // $: smallestTimeGrainOrder = getGrainOrder(smallestTimeGrain);

  import * as Tabs from "@rilldata/web-common/components/tabs";
  import TimeRangeOptionGroup from "./TimeRangeOptionGroup.svelte";

  import RangeDisplay from "../components/RangeDisplay.svelte";
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
    >
      {#if timeString}
        <b class="mr-1 line-clamp-1 flex-none">{getRangeLabel(timeString)}</b>
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
      width={showPanel ? 540 : 280}
      bind:this={searchComponent}
      {context}
      onSelectRange={(range, syntax) => {
        open = false;
        onSelectRange(range, syntax);
      }}
    />
    <!-- <Popover.Label>Filter options by</Popover.Label> -->
    <Tabs.Root value={filter}>
      <Tabs.List class="w-full justify-evenly">
        <Tabs.Trigger
          value="favorites"
          on:click={() => {
            filter = "favorites";
            // selectedTab = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
          }}
          class="rounded-lg p-1 hover:bg-primary-50 data-[state=active]:!text-primary-700 font-semibold px-1.5 flex-none  data-[state=active]:bg-primary-50"
        >
          favorites
        </Tabs.Trigger>
        <Tabs.Trigger
          value=""
          on:click={() => {
            filter = "";
          }}
          class="rounded-lg p-1 hover:bg-primary-50 data-[state=active]:!text-primary-700 font-semibold px-1.5 flex-none  data-[state=active]:bg-primary-50"
        >
          all
        </Tabs.Trigger>
        <!-- <div class="flex gap-x-2 p-2 border-b"> -->
        {#each timeGrainOptions as option (option)}
          <Tabs.Trigger
            class="rounded-lg p-1 hover:bg-primary-50 data-[state=active]:!text-primary-700 font-semibold px-1.5 flex-none  data-[state=active]:bg-primary-50"
            value={TIME_GRAIN[option].label.toLowerCase()}
            on:click={() => {
              filter = TIME_GRAIN[option].label.toLowerCase();
            }}
            title={TIME_GRAIN[option].label}
          >
            {V1TimeGrainToAlias[option]}
          </Tabs.Trigger>
        {/each}
      </Tabs.List>
    </Tabs.Root>
    <!-- </div> -->
    <div class="flex w-[280px] max-h-fit pb-1" style:height="500px">
      <div
        class="flex flex-col w-full overflow-y-auto overflow-x-hidden flex-none pt-1"
      >
        <div class="overflow-x-hidden px-1">
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

          {#if filter === "favorites"}
            <TimeRangeOptionGroup
              {filter}
              {timeString}
              type="last"
              options={timeRanges.map((range) => {
                return {
                  string: range.range ?? "",
                  parsed: {
                    getLabel() {
                      return range.range;
                    },
                  },
                };
              })}
              onClick={handleRangeSelect}
            />
          {/if}

          {#if filter !== "favorites"}
            <TimeRangeOptionGroup
              {filter}
              {timeString}
              type="last"
              options={groups.lastN}
              onClick={handleRangeSelect}
            />

            <TimeRangeOptionGroup
              {filter}
              {timeString}
              type="this"
              options={groups.this}
              onClick={handleRangeSelect}
            />

            <TimeRangeOptionGroup
              {filter}
              {timeString}
              type="ago"
              options={groups.previous}
              onClick={handleRangeSelect}
            />

            <TimeRangeOptionGroup
              {filter}
              {timeString}
              type="by"
              options={groups.grainBy}
              onClick={handleRangeSelect}
            />
          {/if}

          <!-- {#each groups.byGrain as range, i (i)}
            <TimeRangeMenuItem
              {range}
              type="by"
              selected={timeString === range}
              parsed={parseRillTime(range)}
              onClick={handleRangeSelect}
            />
          {/each}

          {#if groups.byGrain.length}
            <div class="h-px w-full bg-gray-300" />
          {/if} -->

          <!-- {#each rangeBuckets.ranges as ranges, i (i)} -->

          <!-- {#if i === 0}
              <form
                class="flex gap-x-1 items-center px-2 py-1"
                on:submit={(e) => {
                  // get the input value
                  const inputValue = e.target[0].value;
                  console.log(inputValue);
                  const grainAlias = V1TimeGrainToAlias[selectedTab];
                  onSelectRange(`${inputValue}${grainAlias}~`, true);
                  open = false;
                }}
              >
                Last
                <input
                  class="w-12 rounded-sm outline-none pl-1 border"
                  type="number"
                  name="integer"
                />
                {V1TimeGrainToDateTimeUnit[selectedTab]}s
              </form>
            {/if} -->

          <!-- {#if ranges.length}
            <div class="h-px w-full bg-gray-300" />
          {/if} -->
          <!-- {/each} -->

          {#if allTimeAllowed}
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
          {/if}

          <div
            class="h-2 w-full bg-surface my-1 sticky bottom-7 flex justify-center items-center"
          >
            <div class="h-px w-full bg-gray-200" />
          </div>

          <button
            class="group h-7 sticky bottom-0 bg-surface flex-none px-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
            on:click={() => {
              showPanel = !showPanel;
            }}
          >
            <span class:font-bold={timeString === "Custom"}> Custom...</span>
          </button>
        </div>

        {#if parsedTime}
          <!-- <div  class="h-px w-full bg-gray-300"/>
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
                if (isComplete) {
                  if (selectedMeta) return;
                  const updatedString = `${selected}~`;

                  onSelectRange(updatedString, true);
                } else if (selected) {
                  const updatedString = selected.replace("~", "");
                  onSelectRange(updatedString, true);
                }
              }}
            />
          </div> -->
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
  </Popover.Content>
</Popover.Root>

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
