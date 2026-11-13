<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";

  // Only renders the widget's time filters.
  // The component writes the changes back to its `time_filters`, see `BaseCanvasComponent`.
  let {
    id,
    component,
    showComparison,
    showGrain,
  }: {
    id: string;
    component: BaseCanvasComponent;
    showComparison: boolean;
    showGrain: boolean;
  } = $props();

  let {
    localTimeFilterManager,
    metricsViewsProvider,
    parent: {
      dashboardProvider: { yamlConfigProvider },
    },
  } = $derived(component);

  let { urlTimeRange, urlComparisonTimeRange } = $derived(
    localTimeFilterManager,
  );

  let hasLocalTimeRange = $derived(
    !!urlTimeRange && urlTimeRange !== INHERIT_TIME_RANGE_ALIAS,
  );
  let hasLocalComparison = $derived(
    !!urlComparisonTimeRange &&
      urlComparisonTimeRange !== INHERIT_TIME_RANGE_ALIAS,
  );
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <InputLabel
    capitalize={false}
    small
    label={m.canvas_time_range_label()}
    {id}
  />
  <TimeFilters
    timeFilterManager={localTimeFilterManager}
    {metricsViewsProvider}
    {yamlConfigProvider}
    context="filters-input"
    config={{
      showInheritRange: true,
      hidePan: true,
      showFullRange: false,
      lockTimeZone: !hasLocalTimeRange,
      allowCustomTimeRange: yamlConfigProvider.allowCustomTimeRange,
      showComparisonSelector: showComparison,
      showGrainPicker: showGrain,
    }}
  />
  <div class="text-fg-secondary">
    {#if hasLocalTimeRange}
      {m.canvas_time_range_override_hint()}
    {:else}
      {m.canvas_time_range_inherit_hint()}
    {/if}
  </div>
  {#if showComparison}
    <div class="text-fg-secondary">
      {#if hasLocalComparison}
        {m.canvas_comparison_override_hint()}
      {:else}
        {m.canvas_comparison_inherit_hint()}
      {/if}
    </div>
  {/if}
</div>
