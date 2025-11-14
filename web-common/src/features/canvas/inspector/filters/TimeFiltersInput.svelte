<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { DateTime, Interval } from "luxon";
  import type { TimeControls } from "../../stores/time-control";

  export let id: string;
  export let localTimeControls: TimeControls;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: { spec },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    allTimeRange,
    timeRangeStateStore,
    comparisonRangeStateStore,
    selectedTimezone,
    minTimeGrain: _minTimeGrain,
    set,
    searchParamsStore,
    clearAll,
  } = localTimeControls);

  $: ({ selectedTimeRange, timeStart, timeEnd } = $timeRangeStateStore || {});

  $: localFiltersEnabled = Boolean($searchParamsStore.size);

  $: selectedComparisonTimeRange =
    $comparisonRangeStateStore?.selectedComparisonTimeRange;

  $: selectedRangeAlias = selectedTimeRange?.name;
  $: activeTimeGrain = selectedTimeRange?.interval;
  $: defaultTimeRange = $spec?.defaultPreset?.timeRange;
  $: timeRanges = $spec?.timeRanges ?? [];

  $: activeTimeZone = $selectedTimezone;
  $: minTimeGrain = $_minTimeGrain;

  $: interval = selectedTimeRange
    ? Interval.fromDateTimes(
        DateTime.fromJSDate(selectedTimeRange.start).setZone(activeTimeZone),
        DateTime.fromJSDate(selectedTimeRange.end).setZone(activeTimeZone),
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
      faint={!localFiltersEnabled}
    />
    <Switch
      checked={localFiltersEnabled}
      on:click={() => {
        if (localFiltersEnabled) {
          clearAll();
        } else {
          set.range("P14D");
          set.zone("UTC");
          set.grain("TIME_GRAIN_HOUR");
        }
      }}
      small
    />
  </div>
  <div class="text-gray-500">
    {#if localFiltersEnabled}
      Overriding inherited time filters from canvas.
    {:else}
      Overrides inherited time filters from canvas when ON.
    {/if}
  </div>

  {#if localFiltersEnabled}
    <div class="flex flex-row flex-wrap pt-2 gap-y-1.5 items-center">
      <SuperPill
        context="filters-input"
        allTimeRange={$allTimeRange}
        {selectedRangeAlias}
        showPivot={!showGrain}
        {minTimeGrain}
        {defaultTimeRange}
        availableTimeZones={[]}
        {timeRanges}
        complete={false}
        {interval}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        {activeTimeZone}
        hidePan
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
          {minTimeGrain}
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
