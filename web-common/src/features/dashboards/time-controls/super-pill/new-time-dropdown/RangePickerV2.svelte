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
    RillPeriodToGrainInterval,
    type RillTime,
  } from "../../../url-state/time-ranges/RillTime";
  import { getTimeRangeOptionsByGrain } from "@rilldata/web-common/lib/time/defaults";
  import {
    getAllowedGrains,
    getGrainOrder,
    getLowerOrderGrain,
    getSmallestGrainFromISODuration,
    GrainAliasToV1TimeGrain,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { TimeGrainOptions } from "@rilldata/web-common/lib/time/defaults";
  import TimeRangeOptionGroup from "./TimeRangeOptionGroup.svelte";
  import RangeDisplay from "../components/RangeDisplay.svelte";
  import TruncationSelector from "./TruncationSelector.svelte";
  import { overrideRillTimeRef } from "../../../url-state/time-ranges/parser";
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import { builderActions, Tooltip, getAttrs } from "bits-ui";
  import ZoneContent from "../components/ZoneContent.svelte";
  import SyntaxElement from "../components/SyntaxElement.svelte";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import {
    constructAsOfString,
    isUsingLegacyTime,
    constructNewString,
  } from "../../new-time-controls";
  import PrimaryRangeTooltip from "./PrimaryRangeTooltip.svelte";

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
  let showCalendarPicker = false;
  let truncationGrain: V1TimeGrain | undefined = undefined;
  let timeZonePickerOpen = false;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      // This is not necessarily an error as the parser does not work with Legacy syntax
      parsedTime = undefined;
    }
  }

  $: hideTruncationSelector = parsedTime?.interval instanceof RillIsoInterval;

  $: usingLegacyTime = isUsingLegacyTime(timeString);

  $: snapToEnd = usingLegacyTime ? true : !!parsedTime?.asOfLabel?.offset;
  $: ref = usingLegacyTime ? "latest" : (parsedTime?.asOfLabel?.label ?? "now");

  $: truncationGrain = usingLegacyTime
    ? timeString?.startsWith("rill")
      ? V1TimeGrain.TIME_GRAIN_DAY
      : getSmallestGrainFromISODuration(timeString ?? "PT1M")
    : parsedTime?.asOfLabel?.snap
      ? GrainAliasToV1TimeGrain[parsedTime.asOfLabel?.snap]
      : undefined;

  $: dateTimeAnchor = returnAnchor(ref);

  $: selectedLabel = getRangeLabel(timeString);

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

  function handleRangeSelect(range: string, ignoreSnap?: boolean) {
    if (range === ALL_TIME_RANGE_ALIAS) {
      onSelectRange(ALL_TIME_RANGE_ALIAS);
      closeMenu();
    } else {
      try {
        const parsed = parseRillTime(range);

        const isPeriodToDate =
          parsed.interval instanceof RillPeriodToGrainInterval;

        const rangeGrainOrder =
          getGrainOrder(parsed.rangeGrain) - (isPeriodToDate ? 1 : 0);
        const asOfGrainOrder = getGrainOrder(truncationGrain);

        if (asOfGrainOrder > rangeGrainOrder && parsed.rangeGrain) {
          truncationGrain = isPeriodToDate
            ? getLowerOrderGrain(parsed.rangeGrain)
            : parsed.rangeGrain;
        }

        const newAsOfString = constructAsOfString(
          ref,
          ignoreSnap
            ? undefined
            : (truncationGrain ??
                smallestTimeGrain ??
                V1TimeGrain.TIME_GRAIN_MINUTE),
          snapToEnd,
        );

        overrideRillTimeRef(parsed, newAsOfString);
        onSelectRange(parsed.toString());
        closeMenu();
      } catch {
        // This function is called in a controlled manner and should not throw
      }
    }
  }

  function onSelectGrain(grain: V1TimeGrain | undefined) {
    if (!timeString) return;

    const newString = constructNewString({
      currentString: timeString,
      truncationGrain: grain === truncationGrain ? undefined : grain,
      snapToEnd: grain === truncationGrain ? false : snapToEnd,
      ref: ref,
    });

    onSelectRange(newString);
  }

  function onSelectAsOfOption(
    ref: "latest" | "watermark" | "now" | string,
    inclusive: boolean,
  ) {
    if (!timeString) return;
    const newString = constructNewString({
      currentString: timeString,
      truncationGrain: truncationGrain,
      snapToEnd: ref === "watermark" ? false : inclusive,
      ref,
    });

    onSelectRange(newString);
  }

  function returnAnchor(asOf: string): DateTime | undefined {
    if (asOf === "latest") {
      return maxDate.setZone(zone);
    } else if (asOf === "watermark" && watermark) {
      return watermark.setZone(zone);
    } else if (asOf === "now") {
      return DateTime.now().setZone(zone);
    }
  }

  function closeMenu() {
    open = false;
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
  <Popover.Trigger asChild let:builder id="super-pill-trigger">
    <Tooltip.Root openDelay={800}>
      <Tooltip.Trigger
        asChild
        let:builder={tooltipBuilder}
        id="super-pill-trigger"
      >
        <button
          use:builderActions={{ builders: [builder, tooltipBuilder] }}
          {...getAttrs([builder, tooltipBuilder])}
          class="flex gap-x-1.5"
          aria-label="Select time range"
        >
          {#if timeString}
            <b class="line-clamp-1 flex-none">
              {#if selectedLabel?.startsWith("-") || !isNaN(Number(selectedLabel?.[0]))}
                Custom
              {:else}
                {selectedLabel}
              {/if}
            </b>
          {/if}

          <RangeDisplay {interval} />

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

      <Tooltip.Content side="bottom" sideOffset={8} class="z-50">
        <PrimaryRangeTooltip {timeString} {interval} />
      </Tooltip.Content>
    </Tooltip.Root>
  </Popover.Trigger>

  <Popover.Content
    align="start"
    class="p-0 w-fit overflow-hidden flex flex-col"
  >
    <TimeRangeSearch
      inError={!parsedTime && !!timeString && !usingLegacyTime}
      width={showCalendarPicker ? 456 : 224}
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
      class:!w-[456px]={showCalendarPicker}
      style:height="470px"
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
          <div class="w-full h-fit px-1">
            <div class="h-px w-full bg-gray-200 my-1" />
            <button
              class:font-bold={false}
              on:click={() => {
                showCalendarPicker = !showCalendarPicker;
              }}
              class="truncate w-full text-left gap-x-1 pr-1 hover:bg-accent flex items-center flex-shrink pl-2 h-7 rounded-sm"
            >
              <Calendar size="14px" />
              <div class="mr-auto">Calendar</div>

              <CaretDownIcon className="-rotate-90" size="14px" />
            </button>
          </div>
        {/if}

        {#if !lockTimeZone && dateTimeAnchor}
          <div class="w-full h-fit px-1">
            <div class="h-px w-full bg-gray-200 my-1" />

            <Popover.Root portal="#rill-portal" bind:open={timeZonePickerOpen}>
              <Popover.Trigger asChild let:builder>
                <div
                  {...builder}
                  use:builder.action
                  on:click={() => {
                    showCalendarPicker = false;
                  }}
                  role="presentation"
                  class="group h-7 overflow-hidden hover:bg-gray-100 flex-none rounded-sm w-full select-none flex items-center"
                >
                  <button
                    class:font-bold={false}
                    class="truncate w-full text-left gap-x-1 pr-1 flex items-center flex-shrink pl-2 h-full"
                  >
                    <Globe size="14px" />
                    <div class="mr-auto">Time zone</div>
                    <div class="sr-only group-hover:not-sr-only">
                      <SyntaxElement range={zoneAbbreviation} />
                    </div>
                    <CaretDownIcon className="-rotate-90" size="14px" />
                  </button>
                </div>
              </Popover.Trigger>

              <Popover.Content
                align="center"
                side="right"
                sideOffset={12}
                class="p-1 z-50"
              >
                <ZoneContent
                  {context}
                  {availableTimeZones}
                  activeTimeZone={zone}
                  watermark={dateTimeAnchor}
                  onSelectTimeZone={(z) => {
                    onSelectTimeZone(z);
                    closeMenu();
                    timeZonePickerOpen = false;
                  }}
                />
              </Popover.Content>
            </Popover.Root>
          </div>
        {/if}
      </div>

      {#if showCalendarPicker}
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

{#if dateTimeAnchor && !hideTruncationSelector}
  <TruncationSelector
    {dateTimeAnchor}
    grain={truncationGrain}
    rangeGrain={parsedTime?.rangeGrain ?? truncationGrain}
    isPeriodToDate={parsedTime?.interval instanceof RillPeriodToGrainInterval}
    {watermark}
    latest={maxDate}
    {smallestTimeGrain}
    {snapToEnd}
    {ref}
    {zone}
    onSelectEnding={onSelectGrain}
    onToggleAlignment={(inclusive) => {
      onSelectAsOfOption(ref, inclusive);
    }}
    onSelectAsOfOption={(o) => {
      onSelectAsOfOption(o, snapToEnd);
    }}
  />
{/if}
