<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { Interval } from "luxon";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { TimeControls } from "../../stores/time-control";

  export let id: string;
  export let localTimeControls: TimeControls;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      spec: { canvasSpec },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    minMaxTimeStamps,
    _comparisonInterval,
    _comparisonRange,
    _interval,
    _range,
    _grain,
    _showTimeComparison,
    _zone,
    minTimeGrain,
    set,
    clearAll,
  } = localTimeControls);

  $: minMax = $minMaxTimeStamps;

  $: interval = $_interval;

  $: comparisonInterval = $_comparisonInterval;
  $: comparisonRange = $_comparisonRange;
  $: showTimeComparison = $_showTimeComparison;

  $: selectedRangeAlias = $_range;
  $: activeTimeGrain = $_grain;

  $: localFiltersEnabled = Boolean(selectedRangeAlias);

  $: ({ defaultPreset: { timeRange: defaultTimeRange } = {}, timeRanges = [] } =
    $canvasSpec ?? {});

  $: activeTimeZone = $_zone;

  $: timeStart = interval?.start.toISO();
  $: timeEnd = interval?.end.toISO();

  $: allTimeRange = {
    start: minMax ? minMax.min.toJSDate() : new Date(0),
    end: minMax ? minMax.max.toJSDate() : new Date(),
  };

  $: selectedComparisonTimeRange = comparisonInterval
    ? {
        name: comparisonRange,
        start: comparisonInterval.start.toJSDate(),
        end: comparisonInterval.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;

  $: selectedTimeRange = interval
    ? {
        name: selectedRangeAlias,
        start: interval?.start.toJSDate(),
        end: interval?.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;
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
        {allTimeRange}
        {selectedRangeAlias}
        showPivot={!showGrain}
        minTimeGrain={$minTimeGrain}
        {defaultTimeRange}
        availableTimeZones={[]}
        {timeRanges}
        complete={false}
        interval={interval || Interval.invalid("No interval")}
        {timeStart}
        {timeEnd}
        {activeTimeGrain}
        {activeTimeZone}
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
          {allTimeRange}
          {selectedTimeRange}
          showFullRange={false}
          {selectedComparisonTimeRange}
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
