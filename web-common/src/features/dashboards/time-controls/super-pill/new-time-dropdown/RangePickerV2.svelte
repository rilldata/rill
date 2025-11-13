<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, Interval } from "luxon";
  import type {
    ISODurationString,
    NamedRange,
    RangeBuckets,
  } from "../../new-time-controls";
  import {
    ALL_TIME_RANGE_ALIAS,
    getRangeLabel,
    RILL_TO_LABEL,
  } from "../../new-time-controls";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import {
    RillAllTimeInterval,
    RillIsoInterval,
    RillPeriodToGrainInterval,
    RillTimeLabel,
    type RillTime,
  } from "../../../url-state/time-ranges/RillTime";
  import {
    getGrainOrder,
    getLowerOrderGrain,
    getSmallestGrainFromISODuration,
    GrainAliasToV1TimeGrain,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import * as Popover from "@rilldata/web-common/components/popover";
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
    constructNewString,
  } from "../../new-time-controls";
  import PrimaryRangeTooltip from "./PrimaryRangeTooltip.svelte";

  export let timeString: string | undefined;
  export let interval: Interval<true>;
  export let timeGrain: V1TimeGrain | undefined;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let context: string;
  export let minDate: DateTime;
  export let maxDate: DateTime;
  export let rangeBuckets: RangeBuckets;
  export let watermark: DateTime | undefined;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let allowCustomTimeRange = true;
  export let availableTimeZones: string[];
  export let lockTimeZone = false;
  export let showFullRange = true;
  export let onSelectTimeZone: (timeZone: string) => void;
  export let onSelectRange: (range: string) => void;

  let open = false;
  let allTimeAllowed = true;
  let searchComponent: TimeRangeSearch;
  let filter = "";
  let parsedTime: RillTime | undefined = undefined;
  let showCalendarPicker = false;
  let truncationGrain: V1TimeGrain | undefined = undefined;
  let timeZonePickerOpen = false;
  let searchValue: string | undefined = timeString;

  $: if (timeString) {
    try {
      parsedTime = parseRillTime(timeString);
    } catch {
      parsedTime = undefined;
    }
  }

  $: hideTruncationSelector =
    parsedTime?.interval instanceof RillIsoInterval ||
    parsedTime?.interval instanceof RillAllTimeInterval;

  $: usingLegacyTime = parsedTime?.isOldFormat;

  $: hasAsOfClause = !!parsedTime?.asOfLabel;

  $: snapToEnd = usingLegacyTime ? true : !!parsedTime?.asOfLabel?.offset;
  $: ref = usingLegacyTime
    ? RillTimeLabel.Latest
    : parsedTime?.asOfLabel?.label;

  $: truncationGrain = usingLegacyTime
    ? timeString?.startsWith("rill") && !timeString.endsWith("C")
      ? V1TimeGrain.TIME_GRAIN_DAY
      : getSmallestGrainFromISODuration(timeString ?? "PT1M")
    : parsedTime?.asOfLabel?.snap
      ? GrainAliasToV1TimeGrain[parsedTime.asOfLabel?.snap]
      : undefined;

  $: dateTimeAnchor = returnAnchor(ref, zone);

  $: selectedLabel = getRangeLabel(timeString);

  $: zoneAbbreviation = getAbbreviationForIANA(maxDate, zone);

  $: smallestTimeGrainOrder = getGrainOrder(
    smallestTimeGrain || V1TimeGrain.TIME_GRAIN_MINUTE,
  );

  function handleRangeSelect(range: string, ignoreSnap?: boolean) {
    try {
      const parsed = parseRillTime(range);

      const isPeriodToDate =
        parsed.interval instanceof RillPeriodToGrainInterval;

      const rangeGrainOrder =
        getGrainOrder(parsed.rangeGrain) - (isPeriodToDate ? 1 : 0);

      const asOfGrainOrder = getGrainOrder(truncationGrain);

      const shouldAppendAsOfString =
        !parsed.asOfLabel && !(parsed.interval instanceof RillIsoInterval);

      if (asOfGrainOrder > rangeGrainOrder && parsed.rangeGrain) {
        truncationGrain = isPeriodToDate
          ? getLowerOrderGrain(parsed.rangeGrain)
          : parsed.rangeGrain;
      }

      if (shouldAppendAsOfString) {
        const isTruncationGrainAllowed =
          getGrainOrder(truncationGrain) >= smallestTimeGrainOrder;
        const newAsOfString = constructAsOfString(
          ref ?? RillTimeLabel.Latest,
          ignoreSnap
            ? undefined
            : truncationGrain
              ? isTruncationGrainAllowed
                ? truncationGrain
                : parsed.rangeGrain
              : (smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE),
          hasAsOfClause || snapToEnd ? snapToEnd : true,
        );

        overrideRillTimeRef(parsed, newAsOfString);
      }

      onSelectRange(parsed.toString());
      closeMenu();
    } catch {
      // This function is called in a controlled manner and should not throw
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
    ref: RillTimeLabel | undefined,
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

  // Zone is taken as a param to make it reactive
  function returnAnchor(
    asOf: string | undefined,
    zone: string,
  ): DateTime | undefined {
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
  onOpenChange={(o) => {
    if (o) {
      searchValue = timeString;
    }
  }}
>
  <Popover.Trigger asChild let:builder id="super-pill-trigger-{context}">
    <Tooltip.Root openDelay={800}>
      <Tooltip.Trigger
        asChild
        let:builder={tooltipBuilder}
        id="super-pill-trigger-{context}"
      >
        <button
          use:builderActions={{ builders: [builder, tooltipBuilder] }}
          {...getAttrs([builder, tooltipBuilder])}
          class="flex gap-x-1.5"
          aria-label="Select time range"
          type="button"
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

          {#if showFullRange}
            <RangeDisplay {interval} {timeGrain} />

            <div
              class="font-bold bg-gray-100 rounded-[2px] p-1 py-0 text-gray-600 text-[11px]"
            >
              {zoneAbbreviation}
            </div>
          {/if}

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
      bind:searchValue
      onSelectRange={(range) => {
        open = false;
        handleRangeSelect(range);
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
            <TimeRangeOptionGroup
              {filter}
              {timeString}
              options={[parseRillTime(defaultTimeRange)]}
              onClick={handleRangeSelect}
            />
          {/if}

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={rangeBuckets.custom}
            onClick={handleRangeSelect}
          />

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={rangeBuckets.latest}
            onClick={handleRangeSelect}
          />

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={rangeBuckets.periodToDate}
            onClick={handleRangeSelect}
          />

          <TimeRangeOptionGroup
            {filter}
            {timeString}
            options={rangeBuckets.previous}
            onClick={(r) => {
              handleRangeSelect(r, true);
            }}
          />

          {#if allTimeAllowed}
            <div class="w-full h-fit px-1">
              <button
                type="button"
                role="menuitem"
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
              type="button"
              role="menuitem"
              class:font-bold={false}
              on:click={() => {
                showCalendarPicker = !showCalendarPicker;
              }}
              class="truncate w-full text-left gap-x-1 pr-1 hover:bg-accent flex items-center flex-shrink pl-2 h-7 rounded-sm"
            >
              <Calendar size="14px" />
              <div class="mr-auto">Custom</div>

              <CaretDownIcon className="-rotate-90" size="14px" />
            </button>
          </div>
        {/if}

        {#if !lockTimeZone}
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
                    type="button"
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
                  referencePoint={dateTimeAnchor ??
                    interval.end ??
                    DateTime.now()}
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
            {interval}
            {zone}
            minTimeGrain={V1TimeGrainToDateTimeUnit[
              smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE
            ]}
            {maxDate}
            {minDate}
            onApply={() => {
              if (searchValue) handleRangeSelect(searchValue);
            }}
            updateRange={(string) => {
              searchValue = string;
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
