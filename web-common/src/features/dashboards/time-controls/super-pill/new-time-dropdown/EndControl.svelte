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
    getAllowedGrains,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { TimeGrainOptions } from "@rilldata/web-common/lib/time/defaults";

  export let timeString: string | undefined;
  export let interval: Interval<true>;
  export let timeGrainOptions: V1TimeGrain[];
  export let smallestTimeGrain: V1TimeGrain | undefined;

  $: console.log({ timeGrainOptions });

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let showPanel = false;
  let calendarOpen = false;
  let filter = "";
  let canShowEndingControl = true;

  //   $: timeGrainOptions = getAllowedGrains(smallestTimeGrain);

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

  function closeMenu() {
    open = false;
  }

  // LAST 7 DAYS
  // excluding this day
  // excluding this hour
  // to now
  // including this hour
  // including this day

  // LAST 3 Months
  // excluding this month
  // excluding this week
  // excluding this day
  // excluding this hour
  // to now
  // including this hour
  // including this day
  // including this week
  // including this month
</script>

<Popover.Root
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
      <b>ending </b>

      <span
        class="flex-none transition-transform"
        class:-rotate-180={calendarOpen}
      >
        <CaretDownIcon />
      </span>
    </button>
  </Popover.Trigger>

  <Popover.Content align="start" class="w-fit overflow-hidden flex flex-col">
    {#each [...timeGrainOptions].reverse() as grain}
      <span class="mr-1 line-clamp-1 flex-none">
        excluding this {V1TimeGrainToDateTimeUnit[grain]}
      </span>
    {/each}

    <div>now</div>

    {#each timeGrainOptions as grain}
      <span class="mr-1 line-clamp-1 flex-none">
        including this {V1TimeGrainToDateTimeUnit[grain]}
      </span>
    {/each}
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
