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
  import {
    RillIsoInterval,
    type RillTime,
  } from "../../../url-state/time-ranges/RillTime";
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
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import { builderActions, Tooltip, getAttrs } from "bits-ui";
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
  export let availableTimeZones: string[];
  export let lockTimeZone = false;
  export let onSelectTimeZone: (timeZone: string) => void;
  export let onSelectRange: (range: string) => void;
  export let onTimeGrainSelect: (grain: V1TimeGrain) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let filter = "";
  let parsedTime: RillTime | undefined = undefined;
  let showCustomSelector = false;
  let truncationGrain: V1TimeGrain | undefined = undefined;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      // no op
    }
  }

  $: hideTruncationSelector = parsedTime?.interval instanceof RillIsoInterval;

  $: usingLegacyTime = isUsingLegacyTime(timeString);

  $: padded = usingLegacyTime ? true : !!parsedTime?.asOfLabel?.offset;
  $: ref = usingLegacyTime ? "latest" : (parsedTime?.asOfLabel?.label ?? "now");

  $: truncationGrain = usingLegacyTime
    ? timeString?.startsWith("rill")
      ? V1TimeGrain.TIME_GRAIN_DAY
      : getSmallestGrainFromISODuration(timeString ?? "PT1M")
    : parsedTime?.asOfLabel?.snap
      ? GrainAliasToV1TimeGrain[parsedTime.asOfLabel?.snap]
      : undefined;

  $: dateTimeAnchor = returnAnchor(ref);

  $: selectedLabel = timeString && getRangeLabel(timeString);

  $: hasCustomSelected = !parsedTime && timeString !== "inf";

  $: timeGrainOptions = getAllowedGrains(smallestTimeGrain);

  $: allOptions = timeGrainOptions.map(getTimeRangeOptionsByGrain);

  $: if (truncationGrain) onTimeGrainSelect(truncationGrain);

  $: zoneAbbreviation = getAbbreviationForIANA(maxDate, zone);

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
    if (range === ALL_TIME_RANGE_ALIAS) {
      onSelectRange(ALL_TIME_RANGE_ALIAS);
    } else {
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
        padded,
      );

      overrideRillTimeRef(parsed, newAsOfString);
      onSelectRange(parsed.toString());
    }

    closeMenu();
  }

  function closeMenu() {
    open = false;
  }

  function onSelectGrain(grain: V1TimeGrain | undefined) {
    if (!timeString) return;

    const newString = constructNewString({
      currentString: timeString,
      truncationGrain: grain,
      inclusive: padded,
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

    const dateUnits: Record<string, string> = {
      Y: "Y",
      M: "M",
      W: "W",
      D: "D",
    };

    const timeUnits: Record<string, string> = {
      H: "H",
      M: "m",
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
    <Tooltip.Root openDelay={800}>
      <Tooltip.Trigger asChild let:builder={tooltipBuilder}>
        <button
          {...getAttrs([builder, tooltipBuilder])}
          use:builderActions={{ builders: [builder, tooltipBuilder] }}
          class="flex gap-x-1.5"
          aria-label="Select time range"
          data-state={open ? "open" : "closed"}
        >
          {#if timeString}
            <b class=" line-clamp-1 flex-none">
              {#if selectedLabel?.startsWith("-") || !isNaN(Number(selectedLabel?.[0]))}
                Custom
              {:else}
                {selectedLabel}
              {/if}
            </b>
          {/if}

          {#if interval.isValid}
            <RangeDisplay {interval} />
          {/if}

          <div
            class="font-bold bg-gray-100 rounded-[2px] p-1 py-0 text-gray-600 text-[11px]"
          >
            {zoneAbbreviation}
          </div>

          <span class="flex-none transition-transform" class:-rotate-180={open}>
            <CaretDownIcon />
          </span>
        </button>
      </Tooltip.Trigger>

      <Tooltip.Content side="bottom" sideOffset={8}>
        <TooltipContent class="flex-col flex items-center gap-y-0 p-3">
          <span class="font-semibold italic mb-1">{timeString}</span>
          <span
            >{interval.start.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
          </span>
          <span>to</span>
          <span
            >{interval.end.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
          </span>
        </TooltipContent>
      </Tooltip.Content>
    </Tooltip.Root>
  </Popover.Trigger>

  <Popover.Content
    align="start"
    class="p-0 w-fit overflow-hidden flex flex-col"
  >
    <TimeRangeSearch
      inError={!parsedTime && !!timeString && !usingLegacyTime}
      width={showCustomSelector ? 456 : 224}
      bind:this={searchComponent}
      {context}
      {timeString}
      onSelectRange={(range) => {
        open = false;
        onSelectRange(range);
      }}
    />

    <div
      class="flex w-56 max-h-fit"
      class:!w-[456px]={showCustomSelector}
      style:height="460px"
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

            <div class="h-px w-full bg-gray-200" />
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

          {#if allTimeAllowed}
            <div class="w-full h-fit px-1">
              <button
                class="group h-7 px-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
                on:click={() => {
                  handleRangeSelect("inf");
                }}
              >
                <span class:font-bold={timeString === ALL_TIME_RANGE_ALIAS}>
                  {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
                </span>
              </button>
            </div>
          {/if}
        </div>

        {#if allowCustomTimeRange}
          <TimeRangeOptionGroup
            {filter}
            hideDivider
            timeString={hasCustomSelected ? "custom" : ""}
            options={[{ label: "Custom", string: "custom" }]}
            onClick={() => {
              showCustomSelector = !showCustomSelector;
            }}
          />
        {/if}
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
    <!-- {#if showCustomSelector} -->
    <!-- <div class="bg-slate-50 border-l p-3 size-full">
      <Elements.Zone
        {context}
        watermark={interval.end ?? DateTime.fromJSDate(new Date())}
        activeTimeZone={zone}
        {lockTimeZone}
        {availableTimeZones}
        {onSelectTimeZone}
      ></Elements.Zone>
    </div> -->
    <!-- {/if} -->
  </Popover.Content>
</Popover.Root>

{#if dateTimeAnchor && !hideTruncationSelector}
  <TruncationSelector
    {dateTimeAnchor}
    grain={truncationGrain}
    {watermark}
    latest={maxDate}
    rangeGrain={parsedTime?.rangeGrain ?? truncationGrain}
    {smallestTimeGrain}
    inclusive={padded}
    {ref}
    onSelectEnding={onSelectGrain}
    onToggleAlignment={(inclusive) => {
      onSelectAsOfOption(ref, inclusive);
    }}
    onSelectAsOfOption={(o) => {
      onSelectAsOfOption(o, padded);
    }}
  />
{/if}

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
