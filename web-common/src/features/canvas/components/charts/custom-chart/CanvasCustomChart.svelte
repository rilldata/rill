<script lang="ts">
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/agentic/CustomChartRenderer.svelte";
  import AgenticChartPrompt from "./AgenticChartPrompt.svelte";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;
  export let editable: boolean = false;

  $: ({ specStore, timeAndFilterStore } = component);

  $: hasValidSpec = component.isValid($specStore);
</script>

{#if hasValidSpec}
  <CustomChartRenderer
    name={component.id}
    spec={$specStore.vega_spec}
    whereFilter={$timeAndFilterStore?.where}
    timeRange={$timeAndFilterStore?.timeRange}
    metricsSQL={$specStore.metrics_sql}
    showDataTable={editable}
  />
{:else}
  <AgenticChartPrompt {component} />
{/if}
