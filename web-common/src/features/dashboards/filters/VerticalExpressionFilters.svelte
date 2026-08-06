<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import AddExpressionFilterButton from "@rilldata/web-common/features/dashboards/filters/AddExpressionFilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilter.svelte";
  import Button from "../../../components/button/Button.svelte";

  let {
    id = "expression-filter",
    expressionFilterManager,
    excludedDimensions,
    excludedMeasures,
  }: {
    id?: string;
    expressionFilterManager: ExpressionFilterManager;
    excludedDimensions?: Record<string, boolean>;
    excludedMeasures?: Record<string, boolean>;
  } = $props();

  let metricsViewsProvider = $derived(
    expressionFilterManager.metricsViewsProvider,
  );
  let yamlConfigProvider = $derived(expressionFilterManager.yamlConfigProvider);

  let allDimensions = $derived(metricsViewsProvider.dimensions);

  let hasFilters = $derived(
    expressionFilterManager.filterManagers.dimensions.length > 0 ||
      expressionFilterManager.filterManagers.measures.length > 0,
  );
</script>

<div class="flex justify-between gap-x-2">
  <InputLabel small label={m.canvas_filters()} {id} />

  <AddExpressionFilterButton
    {expressionFilterManager}
    {excludedDimensions}
    {excludedMeasures}
  />
</div>

<!-- TODO: complex filters -->

<div class="relative flex flex-col gap-x-2 gap-y-2 items-start">
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#each expressionFilterManager.filterManagers.dimensions as dimensionManager (dimensionManager.name)}
      <DimensionFilter
        manager={expressionFilterManager}
        {dimensionManager}
        {yamlConfigProvider}
        timeStart={new Date(0).toISOString()}
        timeEnd={new Date().toISOString()}
        timeControlsReady
        openOnMount={expressionFilterManager.temporaryFilterName ===
          dimensionManager.name}
      />
    {/each}

    {#each expressionFilterManager.filterManagers.measures as measureManager (measureManager.name)}
      <MeasureFilter
        {measureManager}
        {yamlConfigProvider}
        {allDimensions}
        openOnMount={expressionFilterManager.temporaryFilterName ===
          measureManager.name}
      />
    {/each}
  </div>

  <div class="ml-auto">
    {#if hasFilters}
      <Button type="text" onClick={() => expressionFilterManager.clear()}>
        {m.dashboard_clear_filters()}
      </Button>
    {/if}
  </div>
</div>
