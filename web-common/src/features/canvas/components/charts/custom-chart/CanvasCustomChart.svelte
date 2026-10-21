<script lang="ts">
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/custom/CustomChartRenderer.svelte";
  import { onDestroy } from "svelte";
  import AgenticChartPrompt from "./AgenticChartPrompt.svelte";
  import { clearComponentConversation } from "./chart-ai-agent";
  import type { CustomChartComponent, QueryFieldMeta } from "./index";

  let {
    component,
    editable,
  }: {
    component: CustomChartComponent;
    editable: boolean;
  } = $props();

  onDestroy(() => {
    clearComponentConversation(component.id);
  });

  let { metricsViewName, specStore, expressionFilters, timeFilters } =
    $derived(component);

  let whereFilter = $derived(
    expressionFilters.exprByMetricsView[metricsViewName],
  );

  let timeRange = $derived(timeFilters.apiTimeRange);

  let hasValidSpec = $derived(component.isValid($specStore));
  let hasContent = $derived(component.hasContent($specStore));

  function handleMetaChange(meta: Record<string, unknown> | undefined) {
    if (!meta?.fields || !Array.isArray(meta.fields)) {
      component.queryFieldsMeta.set([]);
      return;
    }
    component.queryFieldsMeta.set(meta.fields as QueryFieldMeta[]);
  }
</script>

{#if hasValidSpec || hasContent}
  <CustomChartRenderer
    name={component.id}
    spec={$specStore.vega_spec}
    {whereFilter}
    {timeRange}
    metricsSQL={$specStore.metrics_sql}
    showDataTable={editable}
    onMetaChange={handleMetaChange}
  />
{:else if editable}
  <!-- The prompt edits the component by its YAML path, which only matches the file in the editor:
       viewers get a canvas without the content hidden from them, so their positions can differ. -->
  <AgenticChartPrompt {component} />
{/if}
