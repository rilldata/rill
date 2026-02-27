<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { TimeState } from "../../stores/time-state";
  import { ALL_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";

  export let id: string;
  export let localTimeControls: TimeState;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let canvasName: string;
  export let metricsView: string | null;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      timeManager: {
        defaultTimeRangeStore,
        timeRangeOptionsStore,
        minTimeGrainMap,
        availableTimeZonesStore,
        state: { rangeStore: globalRangeStore, minMaxTimeStamps },
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    interval: intervalStore,
    rangeStore,
    comparisonIntervalStore,
    showTimeComparisonStore,
    timeZoneStore,
    grainStore,
    comparisonRangeStore,
    set,
    searchParamsStore,
    clearAll,
  } = localTimeControls);

  $: minMax = $minMaxTimeStamps;

  $: globalRange = $globalRangeStore;
  $: availableTimeZones = $availableTimeZonesStore;

  $: minDate = minMax?.min;
  $: maxDate = minMax?.max;

  $: localFiltersEnabled = Boolean($searchParamsStore.size);

  $: selectedRangeAlias = $rangeStore;
  $: activeTimeGrain = $grainStore;
  $: defaultTimeRange = $defaultTimeRangeStore;
  $: timeRanges = $timeRangeOptionsStore;
  $: showTimeComparison = $showTimeComparisonStore;

  $: activeTimeZone = $timeZoneStore;
  $: minTimeGrain = metricsView ? $minTimeGrainMap.get(metricsView) : undefined;

  $: interval = $intervalStore;

  $: timeStart = interval?.start.toISO();
  $: timeEnd = interval?.end.toISO();

  $: comparisonInterval = $comparisonIntervalStore;
  $: comparisonRange = $comparisonRangeStore;
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
          set.range(globalRange ?? defaultTimeRange ?? ALL_TIME_RANGE_ALIAS);
        }
      }}
      small
    />
  </div>
  <div class="text-fg-secondary">
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
        {minDate}
        {maxDate}
        {selectedRangeAlias}
        showPivot={!showGrain}
        {minTimeGrain}
        {defaultTimeRange}
        {availableTimeZones}
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
          {minDate}
          {maxDate}
          {interval}
          selectedRange={selectedRangeAlias}
          {activeTimeGrain}
          showFullRange={false}
          {comparisonInterval}
          {comparisonRange}
          {showTimeComparison}
          {activeTimeZone}
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
