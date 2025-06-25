<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { DateTime, Interval } from "luxon";
  import type { TimeControls } from "../../stores/time-control";

  export let id: string;
  export let timeFilter: string;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let canvasName: string;
  export let localTimeControls: TimeControls;
  export let onChange: (filter: string) => void = () => {};

  $: ({
    canvasEntity: {
      spec: { canvasSpec },
    },
  } = getCanvasStore(canvasName));

  $: showLocalFilters = Boolean(timeFilter && timeFilter !== "");

  $: filterText = $timeRangeText?.toString() || "";

  $: if (showLocalFilters) {
    // console.log({ filterText });
    // onChange(filterText);
  }

  $: ({
    allTimeRange,
    timeRangeText,
    timeRangeStateStore,
    comparisonRangeStateStore,
    selectedTimezone,
    minTimeGrain,
    set,
  } = localTimeControls);

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $canvasSpec?.defaultPreset?.timeRange;
  $: timeRanges = $canvasSpec?.timeRanges ?? [];

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone($selectedTimezone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone($selectedTimezone),
      )
    : Interval.fromDateTimes($allTimeRange.start, $allTimeRange.end);
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label="Local time range"
      {id}
      faint={!showLocalFilters}
    />
    <Switch
      checked={showLocalFilters}
      on:click={() => {
        showLocalFilters = !showLocalFilters;
        console.log({ showLocalFilters, filterText });
        onChange(showLocalFilters ? filterText : "");
      }}
      small
    />
  </div>
  <div class="text-gray-500">
    {#if showLocalFilters}
      Overriding inherited time filters from canvas.
    {:else}
      Overrides inherited time filters from canvas when ON.
    {/if}
  </div>

  {#if showLocalFilters}
    <div class="flex flex-row flex-wrap pt-2 gap-y-1.5 items-center">
      <SuperPill
        allTimeRange={$allTimeRange}
        {selectedRangeAlias}
        showPivot={!showGrain}
        minTimeGrain={$minTimeGrain}
        {defaultTimeRange}
        availableTimeZones={[]}
        {timeRanges}
        complete={false}
        {interval}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        activeTimeZone={$selectedTimezone}
        canPanLeft={false}
        canPanRight={false}
        showFullRange={false}
        showDefaultItem={false}
        applyRange={(timeRange) => {
          const string = `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`;
          set.range(string);
        }}
        onSelectRange={set.range}
        onTimeGrainSelect={set.grain}
        onSelectTimeZone={set.zone}
        onPan={() => {}}
      />

      {#if showComparison}
        <CanvasComparisonPill
          allTimeRange={$allTimeRange}
          {selectedTimeRange}
          showFullRange={false}
          {selectedComparisonTimeRange}
          showTimeComparison={$comparisonRangeStateStore?.showTimeComparison ??
            false}
          activeTimeZone={$selectedTimezone}
          onDisplayTimeComparison={set.comparison}
          onSetSelectedComparisonRange={(range) => {
            if (range.name === "CUSTOM_COMPARISON_RANGE") {
              const stringRange = `${range.start.toISOString()},${range.end.toISOString()}`;
              set.comparison(stringRange);
            } else if (range.name) {
              set.comparison(range.name);
            }
          }}
        />
      {/if}
    </div>
  {/if}
</div>
