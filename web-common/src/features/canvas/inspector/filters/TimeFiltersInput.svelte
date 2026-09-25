<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { INHERIT_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import { onMount } from "svelte";

  let {
    id,
    localTimeFilters,
    metricsViewsProvider,
    yamlConfigProvider,
    showComparison,
    updateLocalTimeFilterString,
  }: {
    id: string;
    localTimeFilters: TimeFilterManager;
    metricsViewsProvider: MetricsViewsProvider;
    yamlConfigProvider: YAMLConfigProvider;
    showComparison: boolean;
    updateLocalTimeFilterString: (newFilterString: string) => void;
  } = $props();

  let { urlTimeRange, urlComparisonTimeRange } = $derived(localTimeFilters);

  let hasLocalTimeRange = $derived(
    (urlTimeRange && urlTimeRange !== INHERIT_TIME_RANGE_ALIAS) ||
      (urlComparisonTimeRange &&
        urlComparisonTimeRange !== INHERIT_TIME_RANGE_ALIAS),
  );

  onMount(() => {
    return localTimeFilters.storeSync.on("internal-change", (newUrlParams) => {
      if (
        localTimeFilters.urlTimeRange === INHERIT_TIME_RANGE_ALIAS &&
        localTimeFilters.urlComparisonTimeRange === INHERIT_TIME_RANGE_ALIAS
      ) {
        updateLocalTimeFilterString("");
      } else {
        updateLocalTimeFilterString(newUrlParams.toString());
      }
    });
  });
</script>

<div class="flex flex-col gap-y-1 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label={m.canvas_time_range_label()}
      {id}
    />
  </div>
  <div class="text-fg-secondary">
    {#if hasLocalTimeRange}
      {m.canvas_time_range_override_hint()}
    {:else}
      {m.canvas_time_range_inherit_hint()}
    {/if}
  </div>

  <TimeFilters
    timeFilterManager={localTimeFilters}
    {metricsViewsProvider}
    {yamlConfigProvider}
    context="filter-input"
    config={{
      showFullRange: false,
      hidePan: true,
      showComparisonSelector: showComparison,
      showInheritRange: true,
    }}
  />
</div>
