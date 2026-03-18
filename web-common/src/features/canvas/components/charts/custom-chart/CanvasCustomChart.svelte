<script lang="ts">
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/agentic/CustomChartRenderer.svelte";
  import AgenticChartPrompt from "./AgenticChartPrompt.svelte";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;
  export let editable: boolean = false;

  $: ({ specStore, timeAndFilterStore } = component);

  $: ({ metrics_sql, vega_spec } = $specStore);

  // Show the prompt UI when there's no valid spec yet
  $: hasValidSpec =
    Array.isArray(metrics_sql) &&
    metrics_sql.length > 0 &&
    metrics_sql.every((q) => q.trim().length > 0) &&
    typeof vega_spec === "string" &&
    vega_spec.trim().length > 0;
</script>

{#if hasValidSpec}
  <CustomChartRenderer
    name={component.id}
    spec={vega_spec}
    whereFilter={$timeAndFilterStore?.where}
    timeRange={$timeAndFilterStore?.timeRange}
    metricsSQL={metrics_sql}
    showDataTable={editable}
  />
{:else}
  <AgenticChartPrompt {component} />
{/if}
