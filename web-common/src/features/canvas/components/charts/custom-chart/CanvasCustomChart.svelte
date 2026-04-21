<script lang="ts">
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/custom/CustomChartRenderer.svelte";
  import { onDestroy } from "svelte";
  import AgenticChartPrompt from "./AgenticChartPrompt.svelte";
  import { clearComponentConversation } from "./chart-ai-agent";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;
  export let editable: boolean = false;

  onDestroy(() => {
    clearComponentConversation(component.id);
  });

  $: ({ specStore, timeAndFilterStore } = component);

  $: hasValidSpec = component.isValid($specStore);
  $: hasContent = component.hasContent($specStore);
</script>

{#if hasValidSpec || hasContent}
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
