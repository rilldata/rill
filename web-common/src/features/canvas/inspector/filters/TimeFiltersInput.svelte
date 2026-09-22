<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ALL_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";
  import type { ComponentFilterProperties } from "../../components/types";
  import { resolveComparisonRange } from "../../components/comparison-range";
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
    interval: intervalStore,
    rangeStore,
    timeZoneStore,
    grainStore,
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

  $: activeTimeZone = $timeZoneStore;
  $: minTimeGrain = metricsView ? $minTimeGrainMap.get(metricsView) : undefined;

  $: interval = $intervalStore;

  $: timeStart = interval?.start.toUTC().toISO();
  $: timeEnd = interval?.end.toUTC().toISO();

  // Comparison is computed against the component's effective time range,
  // which is the local one when set and the canvas one otherwise.
  $: effectiveInterval = localFiltersEnabled ? interval : $globalIntervalStore;
  $: effectiveRangeAlias = localFiltersEnabled ? selectedRangeAlias : globalRange;
  $: effectiveTimeGrain = localFiltersEnabled
    ? activeTimeGrain
    : $globalGrainStore;
  $: effectiveTimeZone = localFiltersEnabled
    ? activeTimeZone
    : $globalTimeZoneStore;

  $: resolvedComparison = resolveComparisonRange(
    $specStore as ComponentFilterProperties,
  );
  $: inheritedComparisonLabel = $globalShowTimeComparisonStore
    ? (TIME_COMPARISON[$globalComparisonRangeStore as TimeComparisonOption]
        ?.label ?? m.time_custom_range())
    : m.canvas_comparison_off();
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label={m.canvas_local_time_range()}
      {id}
      faint={!localFiltersEnabled}
    />
    <Switch
      checked={localFiltersEnabled}
      onclick={() => {
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
      {m.canvas_overriding_inherited_time_filters()}
    {:else}
      {m.canvas_override_inherited_time_filters_hint()}
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
    </div>
  {/if}

  {#if showComparison}
    <div class="flex flex-col gap-y-1 pt-3">
      <InputLabel
        capitalize={false}
        small
        label={m.canvas_comparison_range_label()}
        id="{id}-comparison"
      />
      <ComparisonRangeInput
        resolved={resolvedComparison}
        inheritedLabel={inheritedComparisonLabel}
        interval={effectiveInterval}
        selectedRangeAlias={effectiveRangeAlias}
        activeTimeGrain={effectiveTimeGrain}
        activeTimeZone={effectiveTimeZone}
        {minTimeGrain}
        {minDate}
        {maxDate}
        onSelect={(value) => component.setComparisonRange(value)}
      />
      <div class="text-fg-secondary">
        {#if resolvedComparison.mode === "inherit"}
          {m.canvas_comparison_inherit_hint()}
        {:else}
          {m.canvas_comparison_override_hint()}
        {/if}
      </div>
    </div>
  {/if}
</div>
