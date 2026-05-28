<script lang="ts">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
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

  $: ({
    title,
    description,
    show_description_as_tooltip,
    time_filters,
    dimension_filters,
  } = $specStore);
</script>

{#if hasValidSpec || hasContent}
  <div class="size-full flex flex-col overflow-hidden">
    <ComponentHeader
      {title}
      {description}
      showDescriptionAsTooltip={show_description_as_tooltip}
      filters={{ time_filters, dimension_filters }}
      {component}
    />
    <CustomChartRenderer
      name={component.id}
      spec={$specStore.vega_spec}
      whereFilter={$timeAndFilterStore?.where}
      timeRange={$timeAndFilterStore?.timeRange}
      metricsSQL={$specStore.metrics_sql}
      showDataTable={editable}
    />
  </div>
{:else}
  <AgenticChartPrompt {component} />
{/if}
