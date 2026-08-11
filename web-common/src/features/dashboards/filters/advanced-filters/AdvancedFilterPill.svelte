<script lang="ts">
  import type { AdvancedFilter } from "@rilldata/web-common/features/dashboards/filters/advanced-filters/advanced-filter.ts";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import {
    convertExpressionToFilterParam,
    convertFilterParamToExpression,
  } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
  import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";

  let {
    advancedFilter,
    expressionFilterManager,
    onChange,
  }: {
    advancedFilter: AdvancedFilter;
    expressionFilterManager: ExpressionFilterManager | undefined;
    onChange: (advancedFilter: AdvancedFilter) => void;
  } = $props();

  let { expr, dimensionsWithInlistFilter } = $derived(
    convertFilterParamToExpression(advancedFilter.sql),
  );
  let dimOrMesManager = $derived(
    expr
      ? expressionFilterManager?.createDimensionOrMeasureManagerForExpr(
          expr,
          dimensionsWithInlistFilter,
        )
      : expressionFilterManager?.createDimensionOrMeasureManagerForName(
          advancedFilter.name,
        ),
  );

  let sql = "";
  $effect(() => {
    if (!dimOrMesManager) return;
    const newSql = dimOrMesManager.expr
      ? convertExpressionToFilterParam(dimOrMesManager.expr, [])
      : "";
    if (newSql === sql) return;

    sql = newSql;
    onChange({ name: advancedFilter.name, sql });
  });
</script>

<div>
  {#if expressionFilterManager && dimOrMesManager}
    {#if dimOrMesManager instanceof DimensionFilterManager}
      <DimensionFilter
        dimensionManager={dimOrMesManager}
        manager={expressionFilterManager}
        yamlConfigProvider={expressionFilterManager.yamlConfigProvider}
        timeControlsReady
        smallChip
        removable={false}
      />
    {:else if dimOrMesManager instanceof MeasureFilterManager}
      <MeasureFilter
        measureManager={dimOrMesManager}
        yamlConfigProvider={expressionFilterManager.yamlConfigProvider}
        allDimensions={expressionFilterManager.metricsViewsProvider.dimensions}
      />
    {/if}
  {:else}
    {advancedFilter.sql}
  {/if}
</div>
