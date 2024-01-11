<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import DefaultTimeRangeMenuItem from "@rilldata/web-common/features/dashboards/time-controls/DefaultTimeRangeMenuItem.svelte";
  import TimeRangeScrubChip from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeScrubChip.svelte";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import {
    ALL_TIME,
    DEFAULT_TIME_RANGES,
  } from "@rilldata/web-common/lib/time/config";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import {
    DashboardTimeControls,
    TimeRange,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { LIST_SLIDE_DURATION } from "../../../layout/config";
  import type { V1TimeGrain } from "../../../runtime-client";
  import CustomTimeRangeInput from "./CustomTimeRangeInput.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let metricViewName: string;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let minTimeGrain: V1TimeGrain;
  export let selectedRange: DashboardTimeControls;

  const dispatch = createEventDispatcher();
  const ctx = getStateManagers();
  const timeControlsStore = useTimeControlStore(ctx);
  const metaQuery = useMetaQuery(ctx);

  const {
    selectors: {
      timeRangeSelectors: { timeRangeSelectorState },
    },
  } = ctx;

  let open = false;
  let isCustomRangeOpen = false;

  $: dashboardStore = useDashboardStore(metricViewName);
  $: scrubEnd = $dashboardStore?.selectedScrubRange?.end;
  $: scrubStart = $dashboardStore?.selectedScrubRange?.start;
  $: timeZone = $dashboardStore?.selectedTimezone;

  $: selectedStart = $timeControlsStore?.selectedTimeRange?.start;
  $: selectedEnd = $timeControlsStore?.selectedTimeRange?.end;

  $: currentSelection = $dashboardStore?.selectedTimeRange?.name;
  $: intermediateSelection = currentSelection;

  function setIntermediateSelection(
    timeRangeName: TimeRangePreset | TimeComparisonOption | undefined,
  ) {
    if (!timeRangeName) {
      return () => {};
    }

    return () => {
      intermediateSelection = timeRangeName;
    };
  }

  function onSelectRelativeTimeRange(
    timeRange: TimeRange,
    closeMenu?: () => void,
  ) {
    dispatch("select-time-range", {
      name: timeRange.name,
      start: timeRange.start,
      end: timeRange.end,
    });
    if (closeMenu) {
      closeMenu();
    }
  }

  function onSelectCustomTimeRange(startDate: string, endDate: string) {
    setIntermediateSelection(TimeRangePreset.CUSTOM)();

    dispatch("select-time-range", {
      name: TimeRangePreset.CUSTOM,
      start: startDate,
      end: endDate,
    });
  }

  function zoomScrub(toggleFloatingElement: () => void) {
    if (!scrubStart || !scrubEnd) return;

    const { start, end } = getOrderedStartEnd(scrubStart, scrubEnd);

    onSelectRelativeTimeRange(
      {
        name: TimeRangePreset.CUSTOM,
        start,
        end,
      },
      toggleFloatingElement,
    );

    dispatch("remove-scrub");
  }

  function handleMenuOpen() {
    if (intermediateSelection !== TimeRangePreset.CUSTOM) {
      isCustomRangeOpen = false;
    }
  }

  function toggleFloatingElement() {
    open = !open;
  }
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={handleMenuOpen}
  closeOnItemClick={false}
>
  <DropdownMenu.Trigger asChild let:builder>
    {#if scrubStart && scrubEnd && timeZone}
      <div class="flex" use:builder.action {...builder}>
        <TimeRangeScrubChip
          on:remove={() => dispatch("remove-scrub")}
          active={open}
          start={scrubStart}
          end={scrubEnd}
          zone={timeZone}
        />
      </div>
    {:else}
      <button
        use:builder.action
        {...builder}
        class:bg-gray-200={open}
        class="flex items-center gap-x-2 rounded px-3 py-2 hover:bg-gray-200 hover:dark:bg-gray-600"
        aria-label="Select time range"
      >
        <span class="ui-copy-icon"><Calendar size="16px" /></span>
        <b>
          <!-- This conditional shouldn't be necessary because there should always be a selected (at least default) time range -->
          {#if intermediateSelection === TimeRangePreset.CUSTOM}
            Custom range
          {:else if currentSelection}
            {#if currentSelection in DEFAULT_TIME_RANGES}
              {DEFAULT_TIME_RANGES[currentSelection].label}
            {:else}
              Last {humaniseISODuration(currentSelection)}
            {/if}
          {:else}
            Select a time range
          {/if}
        </b>
        {#if selectedStart && selectedEnd && currentSelection && timeZone}
          <p>
            {prettyFormatTimeRange(
              selectedStart,
              selectedEnd,
              currentSelection,
              timeZone,
            )}
          </p>
        {/if}

        <IconSpaceFixer pullRight>
          <div class="transition-transform" class:-rotate-180={open}>
            <CaretDownIcon size="14px" />
          </div>
        </IconSpaceFixer>
      </button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content class="w-[300px]" align="start">
    {@const allTime = {
      name: TimeRangePreset.ALL_TIME,
      label: ALL_TIME.label,
      start: boundaryStart,
      end: new Date(boundaryEnd.getTime() + 1), // end is exclusive
    }}

    {#if scrubStart && scrubEnd}
      <DropdownMenu.Item
        class="justify-between"
        on:mouseenter={setIntermediateSelection(TimeRangePreset.CUSTOM)}
        on:click={() => {
          // toggleFloatingElement();
          zoomScrub(toggleFloatingElement);
        }}
      >
        <span> Zoom to subrange </span>
        <span class="ui-copy-muted">Z</span>
      </DropdownMenu.Item>
      <DropdownMenu.Separator class="bg-gray-200" />
    {/if}

    <DropdownMenu.Item
      on:mouseenter={setIntermediateSelection(allTime.name)}
      on:click={() => onSelectRelativeTimeRange(allTime, toggleFloatingElement)}
    >
      <span class:font-bold={intermediateSelection === allTime.name}>
        {allTime.label}
      </span>
    </DropdownMenu.Item>

    {#if $timeRangeSelectorState.showDefaultItem}
      <DefaultTimeRangeMenuItem
        on:mouseenter={setIntermediateSelection(
          $metaQuery.data?.defaultTimeRange,
        )}
        on:click={() => {
          if ($timeControlsStore.defaultTimeRange) {
            onSelectRelativeTimeRange($timeControlsStore.defaultTimeRange);
          }
        }}
        selected={intermediateSelection === $metaQuery.data?.defaultTimeRange}
        isoDuration={$metaQuery.data?.defaultTimeRange}
      />
    {/if}

    {#if $timeRangeSelectorState.latestWindowTimeRanges?.length}
      <DropdownMenu.Separator class="bg-gray-200" />
      {#each $timeRangeSelectorState.latestWindowTimeRanges as timeRange}
        {#if timeRange.name}
          <DropdownMenu.Item
            on:click={() =>
              onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
            on:mouseenter={setIntermediateSelection(timeRange.name)}
          >
            <span class:font-bold={intermediateSelection === timeRange.name}>
              {timeRange.label}
            </span>
          </DropdownMenu.Item>
        {/if}
      {/each}
    {/if}

    {#if $timeRangeSelectorState.periodToDateRanges?.length}
      <DropdownMenu.Separator class="bg-gray-200" />
      {#each $timeRangeSelectorState.periodToDateRanges as timeRange}
        {#if timeRange.name}
          <DropdownMenu.Item
            on:click={() =>
              onSelectRelativeTimeRange(timeRange, toggleFloatingElement)}
            on:mouseenter={setIntermediateSelection(timeRange.name)}
          >
            <span class:font-bold={intermediateSelection === timeRange.name}>
              {timeRange.label}
            </span>
          </DropdownMenu.Item>
        {/if}
      {/each}
    {/if}

    <DropdownMenu.Separator class="bg-gray-200" />

    <DropdownMenu.Item
      class="justify-between"
      on:click={() => {
        setIntermediateSelection(TimeRangePreset.CUSTOM)();
        isCustomRangeOpen = !isCustomRangeOpen;
      }}
    >
      <span class:font-bold={intermediateSelection === TimeRangePreset.CUSTOM}>
        Custom range
      </span>

      <div
        class="transition-transform duration-100"
        class:-rotate-180={isCustomRangeOpen}
      >
        <CaretDownIcon size="14px" />
      </div>
    </DropdownMenu.Item>

    {#if isCustomRangeOpen}
      <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
        <CustomTimeRangeInput
          {boundaryStart}
          {boundaryEnd}
          {minTimeGrain}
          zone={$dashboardStore?.selectedTimezone}
          defaultDate={selectedRange}
          on:apply={(e) => {
            toggleFloatingElement();
            onSelectCustomTimeRange(e.detail.startDate, e.detail.endDate);
          }}
        />
      </div>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
