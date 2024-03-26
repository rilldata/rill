<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
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
  import { createEventDispatcher } from "svelte";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
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
  const metricsView = useMetricsView(ctx);
  const {
    selectors: {
      timeRangeSelectors: { timeRangeSelectorState },
    },
  } = ctx;

  let open = false;

  $: dashboardStore = useDashboardStore(metricViewName);
  $: currentSelection =
    $dashboardStore?.selectedTimeRange?.name ?? TimeRangePreset.ALL_TIME;

  $: defaultTimeRange = $metricsView.data?.defaultTimeRange
    ? ($metricsView.data?.defaultTimeRange as TimeRangePreset)
    : undefined;

  $: selectedSubRange =
    $dashboardStore?.selectedScrubRange?.start &&
    $dashboardStore?.selectedScrubRange?.end
      ? {
          start: $dashboardStore.selectedScrubRange.start,
          end: $dashboardStore.selectedScrubRange.end,
        }
      : null;

  function onSelectRelativeTimeRange(timeRange: TimeRange | undefined) {
    if (!timeRange) {
      return;
    }
    dispatch("select-time-range", {
      name: timeRange.name,
      start: timeRange.start,
      end: timeRange.end,
    });
  }

  function onSelectCustomTimeRange(
    e: CustomEvent<{ startDate: Date; endDate: Date }>,
  ) {
    const { startDate, endDate } = e.detail;
    open = false;

    dispatch("select-time-range", {
      name: TimeRangePreset.CUSTOM,
      start: startDate,
      end: endDate,
    });
  }

  function zoomScrub() {
    if (
      !$dashboardStore?.selectedScrubRange?.start ||
      !$dashboardStore?.selectedScrubRange?.end
    ) {
      return;
    }
    const { start, end } = getOrderedStartEnd(
      $dashboardStore?.selectedScrubRange?.start,
      $dashboardStore?.selectedScrubRange?.end,
    );
    onSelectRelativeTimeRange({
      name: TimeRangePreset.CUSTOM,
      start,
      end,
    });
    dispatch("remove-scrub");
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger aria-label="Select a time range">
    {#if selectedSubRange}
      <TimeRangeScrubChip
        on:remove={() => {
          open = false;
          dispatch("remove-scrub");
        }}
        active={open}
        start={selectedSubRange.start}
        end={selectedSubRange.end}
        zone={$dashboardStore?.selectedTimezone}
      />
    {:else}
      <div
        class:bg-gray-200={open}
        class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-baseline"
        aria-label="Select time range"
      >
        <div class="flex flew-row gap-x-3">
          <div class="font-bold flex flex-row items-center gap-x-3">
            <span class="ui-copy-icon"><Calendar size="16px" /></span>
            <span style:transform="translateY(1px)">
              {#if currentSelection in DEFAULT_TIME_RANGES}
                {DEFAULT_TIME_RANGES[currentSelection].label}
              {:else}
                Last {humaniseISODuration(currentSelection)}
              {/if}
            </span>
          </div>
          <span style:transform="translateY(1px)">
            {prettyFormatTimeRange(
              $timeControlsStore?.selectedTimeRange?.start,
              $timeControlsStore?.selectedTimeRange?.end,
              $timeControlsStore?.selectedTimeRange?.name,
              $dashboardStore?.selectedTimezone,
            )}
          </span>
        </div>
        <IconSpaceFixer pullRight>
          <div class="transition-transform" class:-rotate-180={open}>
            <CaretDownIcon size="14px" />
          </div>
        </IconSpaceFixer>
      </div>
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-80">
    {@const allTime = {
      name: TimeRangePreset.ALL_TIME,
      label: ALL_TIME.label,
      start: boundaryStart,
      end: new Date(boundaryEnd.getTime() + 1), // end is exclusive
    }}
    {#if selectedSubRange}
      <DropdownMenu.Item on:click={zoomScrub}>
        <span> Zoom to subrange </span>
        <span slot="right">Z</span>
      </DropdownMenu.Item>
      <DropdownMenu.Separator />
    {/if}
    <DropdownMenu.Item on:click={() => onSelectRelativeTimeRange(allTime)}>
      <span class:font-bold={currentSelection === allTime.name}>
        {allTime.label}
      </span>
    </DropdownMenu.Item>
    {#if $timeRangeSelectorState.showDefaultItem && defaultTimeRange}
      <DropdownMenu.Item
        on:click={() =>
          onSelectRelativeTimeRange($timeControlsStore.defaultTimeRange)}
      >
        <div class:font-bold={currentSelection === defaultTimeRange}>
          Last {humaniseISODuration(defaultTimeRange)}
        </div>
      </DropdownMenu.Item>
    {/if}
    {#if $timeRangeSelectorState.latestWindowTimeRanges?.length}
      <DropdownMenu.Separator />
      {#each $timeRangeSelectorState.latestWindowTimeRanges as timeRange}
        <DropdownMenu.Item
          on:click={() => onSelectRelativeTimeRange(timeRange)}
        >
          <span class:font-bold={currentSelection === timeRange.name}>
            {timeRange.label}
          </span>
        </DropdownMenu.Item>
      {/each}
    {/if}
    {#if $timeRangeSelectorState.periodToDateRanges?.length}
      <DropdownMenu.Separator />
      {#each $timeRangeSelectorState.periodToDateRanges as timeRange}
        <DropdownMenu.Item
          on:click={() => onSelectRelativeTimeRange(timeRange)}
        >
          <span class:font-bold={currentSelection === timeRange.name}>
            {timeRange.label}
          </span>
        </DropdownMenu.Item>
      {/each}
    {/if}
    {#if $timeRangeSelectorState.previousCompleteDateRanges?.length}
      <DropdownMenu.Separator />
      {#each $timeRangeSelectorState.previousCompleteDateRanges as timeRange}
        <DropdownMenu.Item
          on:click={() => onSelectRelativeTimeRange(timeRange)}
        >
          <span class:font-bold={currentSelection === timeRange.name}>
            {timeRange.label}
          </span>
        </DropdownMenu.Item>
      {/each}
    {/if}
    <DropdownMenu.Separator />

    <CustomTimeRangeInput
      {boundaryStart}
      {boundaryEnd}
      {minTimeGrain}
      zone={$dashboardStore?.selectedTimezone}
      defaultDate={selectedRange}
      on:apply={onSelectCustomTimeRange}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
