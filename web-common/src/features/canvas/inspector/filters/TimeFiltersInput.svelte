<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";
  import type { ComponentFilterProperties } from "../../components/types";
  import {
    resolveTimeFilters,
    TIME_FILTER_INHERIT,
  } from "../../components/time-filters";
  import ComparisonRangeInput from "./ComparisonRangeInput.svelte";

  export let id: string;
  export let component: BaseCanvasComponent;
  export let showComparison: boolean;
  export let showGrain: boolean;
  export let canvasName: string;
  export let metricsView: string | null;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);

  $: ({ localTimeControls, specStore } = component);

  $: ({
    canvasEntity: {
      timeManager: {
        defaultTimeRangeStore,
        timeRangeOptionsStore,
        minTimeGrainMap,
        availableTimeZonesStore,
        state: {
          rangeStore: globalRangeStore,
          minMaxTimeStamps,
          interval: globalIntervalStore,
          grainStore: globalGrainStore,
          timeZoneStore: globalTimeZoneStore,
          showTimeComparisonStore: globalShowTimeComparisonStore,
          comparisonRangeStore: globalComparisonRangeStore,
        },
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    interval: localIntervalStore,
    rangeStore: localRangeStore,
    timeZoneStore: localTimeZoneStore,
    grainStore: localGrainStore,
    set,
  } = localTimeControls);

  $: minMax = $minMaxTimeStamps;
  $: minDate = minMax?.min;
  $: maxDate = minMax?.max;

  $: defaultTimeRange = $defaultTimeRangeStore;
  $: timeRanges = $timeRangeOptionsStore;
  $: availableTimeZones = $availableTimeZonesStore;
  $: minTimeGrain = metricsView ? $minTimeGrainMap.get(metricsView) : undefined;

  $: resolved = resolveTimeFilters(
    ($specStore as ComponentFilterProperties).time_filters,
  );
  $: ({ hasLocalTimeRange } = resolved);

  // The pill always shows the range in effect: the local one when set, the canvas one otherwise.
  // Comparison options are computed against the same range.
  $: interval = hasLocalTimeRange ? $localIntervalStore : $globalIntervalStore;
  $: selectedRangeAlias = hasLocalTimeRange
    ? $localRangeStore
    : $globalRangeStore;
  $: activeTimeGrain = hasLocalTimeRange ? $localGrainStore : $globalGrainStore;
  $: activeTimeZone = hasLocalTimeRange
    ? $localTimeZoneStore
    : $globalTimeZoneStore;

  $: timeStart = interval?.start.toUTC().toISO();
  $: timeEnd = interval?.end.toUTC().toISO();

  $: inheritedComparisonLabel = $globalShowTimeComparisonStore
    ? (TIME_COMPARISON[$globalComparisonRangeStore as TimeComparisonOption]
        ?.label ?? m.time_custom_range())
    : m.canvas_comparison_off();
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <InputLabel
    capitalize={false}
    small
    label={m.canvas_time_range_label()}
    {id}
  />
  <div class="flex flex-row flex-wrap gap-y-1.5 items-center">
    <SuperPill
      context="filters-input"
      {minDate}
      {maxDate}
      {selectedRangeAlias}
      showPivot={!showGrain || !hasLocalTimeRange}
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
      lockTimeZone={!hasLocalTimeRange}
      showFullRange={false}
      showDefaultItem={false}
      inheritOption={{
        label: m.canvas_inherit_from_canvas(),
        selected: !hasLocalTimeRange,
        onSelect: () => set.range(TIME_FILTER_INHERIT),
      }}
      applyRange={(timeRange) => {
        const string = `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`;
        set.range(string);
      }}
      onSelectRange={set.range}
      onTimeGrainSelect={set.grain}
      onSelectTimeZone={set.zone}
      onPan={() => {}}
    />
  </div>
  <div class="text-fg-secondary">
    {#if hasLocalTimeRange}
      {m.canvas_time_range_override_hint()}
    {:else}
      {m.canvas_time_range_inherit_hint()}
    {/if}
  </div>

  {#if showComparison}
    <div class="flex flex-col gap-y-1 pt-3">
      <InputLabel
        capitalize={false}
        small
        label={m.canvas_comparison_range_label()}
        id="{id}-comparison"
      />
      <ComparisonRangeInput
        resolved={resolved.comparison}
        inheritedLabel={inheritedComparisonLabel}
        {interval}
        {selectedRangeAlias}
        {activeTimeGrain}
        {activeTimeZone}
        {minTimeGrain}
        {minDate}
        {maxDate}
        onSelect={(value) => component.setComparisonRange(value)}
      />
      <div class="text-fg-secondary">
        {#if resolved.comparison.mode === "inherit"}
          {m.canvas_comparison_inherit_hint()}
        {:else}
          {m.canvas_comparison_override_hint()}
        {/if}
      </div>
    </div>
  {/if}
</div>
