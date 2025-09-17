<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import CanvasComparisonPill from "@rilldata/web-common/features/canvas/filters/CanvasComparisonPill.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
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
    _range,
    _interval,
    grain,
    _comparisonRange,
    _comparisonInterval,
    _zone,
    minTimeGrain,
    set,
    _showTimeComparison,

    clearAll,
  } = localTimeControls);

  $: interval = $_interval;

  $: localFiltersEnabled = Boolean(interval);

  $: comparisonInterval = $_comparisonInterval;
  $: comparisonRange = $_comparisonRange;
  $: showTimeComparison = $_showTimeComparison;

  $: selectedRangeAlias = $_range;
  $: activeTimeGrain = $grain;

  $: ({ defaultPreset: { timeRange: defaultTimeRange } = {}, timeRanges = [] } =
    $canvasSpec ?? {});

  $: activeTimeZone = $_zone;
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
        minMaxTimeStamps={$minMaxTimeStamps}
        {selectedRangeAlias}
        showPivot={!showGrain}
        minTimeGrain={$minTimeGrain}
        {defaultTimeRange}
        availableTimeZones={[]}
        {timeRanges}
        complete={false}
        {interval}
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
          {comparisonInterval}
          {comparisonRange}
          {activeTimeGrain}
          minMaxTimeStamps={$minMaxTimeStamps}
          {interval}
          range={selectedRangeAlias}
          {showTimeComparison}
          {activeTimeZone}
          setComparison={set.comparison}
        />
      {/if}
    </div>
  {/if}
</div>
