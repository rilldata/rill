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
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import { type RillTime } from "../../../url-state/time-ranges/RillTime";
  import { getTimeRangeOptionsByGrain } from "@rilldata/web-common/lib/time/defaults";
  import {
    getAllowedGrains,
    getGrainOrder,
    getSmallestGrainFromISODuration,
    GrainAliasToV1TimeGrain,
    V1TimeGrainToAlias,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { TimeGrainOptions } from "@rilldata/web-common/lib/time/defaults";
  import TimeRangeOptionGroup from "./TimeRangeOptionGroup.svelte";
  import RangeDisplay from "../components/RangeDisplay.svelte";
  import TruncationSelector from "./TruncationSelector.svelte";
  import { overrideRillTimeRef } from "../../../url-state/time-ranges/parser";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let timeString: string | undefined;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let watermark: DateTime | undefined;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let allowCustomTimeRange = true;
  export let onSelectRange: (range: string) => void;
  export let onTimeGrainSelect: (grain: V1TimeGrain) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let filter = "";
  let parsedTime: RillTime | undefined = undefined;
  let showCustomSelector = false;
  let asOfOpen = false;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      // no op
    }
  }

  $: usingLegacyTime = isUsingLegacyTime(timeString);

  $: ({ ref, truncationGrain, forwardAligned } = parse(timeString ?? ""));

  $: dateTimeAnchor = returnAnchor(ref);

  $: humanReadableAsOf = humanizeAsOf(ref);

  $: selectedLabel = timeString && getRangeLabel(timeString);

  $: hasCustomSelected = !parsedTime;

  $: timeGrainOptions = getAllowedGrains(smallestTimeGrain);

  $: allOptions = timeGrainOptions.map(getTimeRangeOptionsByGrain);

  $: onTimeGrainSelect(truncationGrain);

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

  function isUsingLegacyTime(timeString: string | undefined): boolean {
    return (
      timeString?.startsWith("rill") ||
      timeString?.startsWith("P") ||
      timeString?.startsWith("p") ||
      false
    );
  }

  function handleRangeSelect(range: string, ignoreSnap?: boolean) {
    const parsed = parseRillTime(range);

    const rangeGrainOrder = getGrainOrder(parsed.rangeGrain);
    const asOfGrainOrder = getGrainOrder(truncationGrain);

    if (asOfGrainOrder > rangeGrainOrder) {
      truncationGrain = parsed.rangeGrain;
    }

    const newAsOfString = constructAsOfString(
      ref,
      ignoreSnap
        ? undefined
        : (truncationGrain ??
            smallestTimeGrain ??
            V1TimeGrain.TIME_GRAIN_MINUTE),
      forwardAligned,
    );

    overrideRillTimeRef(parsed, newAsOfString);
    onSelectRange(parsed.toString());

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }

  function humanizeAsOf(asOf: string): string {
    switch (asOf) {
      case "latest":
        return "latest data";
      case "watermark":
        return "watermark";
      case "now":
        return "now";

      default:
        return "custom";
    }
  }

  function onSelectGrain(grain: V1TimeGrain | undefined) {
    if (!timeString) return;

    const newString = constructNewString({
      currentString: timeString,
      truncationGrain: grain,
      inclusive: forwardAligned,
      ref: ref,
    });

    onSelectRange(newString);
  }

  function constructAsOfString(
    asOf: string,
    grain: V1TimeGrain | undefined | null,
    forwardAlign: boolean,
  ): string {
    if (!grain) {
      return asOf;
    }

    const alias = V1TimeGrainToAlias[grain];

    let base: string;

    if (asOf === "latest" || asOf === undefined) {
      base = `latest/${alias}`;
    } else if (asOf === "watermark") {
      base = `watermark/${alias}`;
    } else if (asOf === "now") {
      base = `now/${alias}`;
    } else {
      base = `${asOf}/${alias}`;
    }

    if (forwardAlign) {
      return `${base}+1${alias}`;
    } else {
      return base;
    }
  }

  //
  function convertLegacyTime(timeString: string) {
    if (timeString.startsWith("rill-")) {
      if (timeString === "rill-TD") return "DTD";
      return timeString.replace("rill-", "");
    } else if (timeString.startsWith("P") || timeString.startsWith("p")) {
      return convertIsoToRillTime(timeString);
    }
    return timeString;
  }

  function convertIsoToRillTime(iso: string): string {
    const upper = iso.toUpperCase();

    if (!upper.startsWith("P")) {
      throw new Error("Invalid ISO duration: must start with P");
    }

    const result: string[] = [];

    const [datePartRaw, timePartRaw] = upper.slice(1).split("T");
    const datePart = datePartRaw || "";
    const timePart = timePartRaw || "";

    // Date units: Y, M (Month), W, D
    const dateUnits: Record<string, string> = {
      Y: "Y",
      M: "M", // Month
      W: "W",
      D: "D",
    };

    // Time units: H, M (Minute), S
    const timeUnits: Record<string, string> = {
      H: "H",
      M: "m", // Minute (lowercase in Rill)
      S: "S",
    };

    for (const [unit, rill] of Object.entries(dateUnits)) {
      const match = datePart.match(new RegExp(`(\\d+(\\.\\d+)?)${unit}`));
      if (match) result.push(`${match[1]}${rill}`);
    }

    for (const [unit, rill] of Object.entries(timeUnits)) {
      const match = timePart.match(new RegExp(`(\\d+(\\.\\d+)?)${unit}`));
      if (match) result.push(`${match[1]}${rill}`);
    }

    return result.join("");
  }

  function constructNewString({
    currentString,
    truncationGrain,
    inclusive,
    ref,
  }: {
    currentString: string;
    truncationGrain: V1TimeGrain | undefined | null;
    inclusive: boolean;
    ref: "watermark" | "latest" | "now" | string;
  }): string {
    const legacy = isUsingLegacyTime(currentString);

    const rillTime = parseRillTime(
      legacy ? convertLegacyTime(currentString) : currentString,
    );

    const newAsOfString = constructAsOfString(ref, truncationGrain, inclusive);

    overrideRillTimeRef(rillTime, newAsOfString);

    return rillTime.toString();
  }

  function onSelectAsOfOption(
    ref: "latest" | "watermark" | "now" | string,
    inclusive: boolean,
  ) {
    console.log({ ref, inclusive });
    if (!timeString) return;
    const newString = constructNewString({
      currentString: timeString,
      truncationGrain: truncationGrain,
      inclusive: ref === "watermark" ? false : inclusive,
      ref,
    });

    onSelectRange(newString);
  }

  function returnAnchor(asOf: string) {
    if (asOf === "latest") {
      return maxDate.setZone(zone);
    } else if (asOf === "watermark" && watermark) {
      return watermark.setZone(zone);
    } else if (asOf === "now") {
      return DateTime.now().setZone(zone);
    }
  }

  function parse(timeString: string) {
    if (usingLegacyTime) {
      return {
        ref: "latest",
        truncationGrain: timeString.startsWith("rill")
          ? "TIME_GRAIN_DAY"
          : getSmallestGrainFromISODuration(timeString),
        forwardAligned: true,
      };
    }

    const patterns = {
      asOfClause: /as of (.+)$/i,
      timeWithGrain:
        /^(latest|watermark|now)\/([HhMmSsQqDdYyWw]+)(?:\+1[HhMmSsQqDdYyWw]+)?$/,
      justType: /^(latest|watermark|now)$/,
    };

    let ref: string = "now";
    let truncationGrain: V1TimeGrain | undefined = undefined;
    let forwardAligned = false;

    const asOfMatch = timeString.match(patterns.asOfClause);
    const clauseToCheck = asOfMatch ? asOfMatch[1] : timeString;

    const timeWithGrainMatch = clauseToCheck.match(patterns.timeWithGrain);

    if (timeWithGrainMatch) {
      ref = timeWithGrainMatch[1];
      truncationGrain = GrainAliasToV1TimeGrain[timeWithGrainMatch[2]];
    } else {
      const justTypeMatch = clauseToCheck.match(patterns.justType);
      if (justTypeMatch) {
        ref = justTypeMatch[1];
      } else if (clauseToCheck !== timeString) {
        ref = clauseToCheck;
      }
    }

    if (clauseToCheck.includes("+1")) {
      forwardAligned = true;
    }

    return { ref, truncationGrain, forwardAligned };
  }
</script>

<svelte:window
  on:keydown={(e) => {
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
      class="flex gap-x-1"
      aria-label="Select time range"
      data-state={open ? "open" : "closed"}
    >
      {#if timeString}
        <b class="mr-1 line-clamp-1 flex-none">{selectedLabel}</b>
      {/if}

      {#if interval.isValid}
        <RangeDisplay {interval} />
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
      onSelectRange={(range) => {
        open = false;
        onSelectRange(range);
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
            onClick={(r) => {
              handleRangeSelect(r, true);
            }}
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
            applyRange={(interval) => {
              const string = `${interval.start.toFormat("yyyy-MM-dd")} to ${interval.end.toFormat("yyyy-MM-dd")}`;
              onSelectRange(string);
            }}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

{#if truncationGrain && dateTimeAnchor}
  <TruncationSelector
    {dateTimeAnchor}
    grain={truncationGrain}
    rangeGrain={parsedTime?.rangeGrain ?? truncationGrain}
    {smallestTimeGrain}
    inclusive={forwardAligned}
    {ref}
    onSelectEnding={onSelectGrain}
    onToggleAlignment={(inclusive) => {
      onSelectAsOfOption(ref, inclusive);
    }}
  />
{/if}

<DropdownMenu.Root bind:open={asOfOpen}>
  <DropdownMenu.Trigger class="flex gap-x-1">
    as of <b>{humanReadableAsOf}</b>

    <span class="flex-none transition-transform" class:-rotate-180={asOfOpen}>
      <CaretDownIcon />
    </span>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    <Tooltip alignment="end" location="right" distance={8}>
      <DropdownMenu.CheckboxItem
        checked={ref === "latest"}
        checkRight
        class="flex justify-between"
        on:click={() => {
          onSelectAsOfOption("latest", forwardAligned);
        }}
      >
        latest data
      </DropdownMenu.CheckboxItem>
      <TooltipContent slot="tooltip-content" maxWidth="600px">
        {maxDate
          .setZone(zone)
          .toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
      </TooltipContent>
    </Tooltip>

    {#if watermark}
      <Tooltip alignment="end" location="right" distance={8}>
        <DropdownMenu.CheckboxItem
          checkRight
          checked={ref === "watermark"}
          class="flex justify-between"
          on:click={() => {
            onSelectAsOfOption("watermark", forwardAligned);
          }}
        >
          watermark
        </DropdownMenu.CheckboxItem>

        <TooltipContent slot="tooltip-content" maxWidth="600px">
          {watermark
            .setZone(zone)
            .toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
        </TooltipContent>
      </Tooltip>
    {/if}

    <Tooltip alignment="end" location="right" distance={8}>
      <DropdownMenu.CheckboxItem
        checkRight
        checked={ref === "now"}
        class="flex justify-between"
        on:click={() => {
          onSelectAsOfOption("now", forwardAligned);
        }}
      >
        now
      </DropdownMenu.CheckboxItem>
      <TooltipContent slot="tooltip-content" maxWidth="600px">
        {DateTime.now()
          .setZone(zone)
          .toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
      </TooltipContent>
    </Tooltip>
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
