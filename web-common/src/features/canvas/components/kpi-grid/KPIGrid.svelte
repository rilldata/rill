<script lang="ts">
  import type { KPIGridComponent } from ".";
  import ComponentError from "../ComponentError.svelte";
  import type { KPISpec } from "../kpi";
  import { validateKPIGridSchema } from "./selector";
  import { getMinWidth } from "../kpi";
  import KPIProvider from "../kpi/KPIProvider.svelte";

  export let component: KPIGridComponent;

  let kpis: KPISpec[];

  $: ({
    specStore,
    timeAndFilterStore,
    parent: { name: canvasName },
  } = component);
  $: kpiGridProperties = $specStore;
  $: schema = validateKPIGridSchema(kpiGridProperties);

  // Convert measures to KPI specs
  $: kpis = (kpiGridProperties.measures || []).map((measure) => ({
    metrics_view: kpiGridProperties.metrics_view,
    measure,
    sparkline: kpiGridProperties.sparkline,
    comparison: kpiGridProperties.comparison,
    dimension_filters: kpiGridProperties.dimension_filters,
    time_filters: kpiGridProperties.time_filters,
  }));

  $: sparkline = kpiGridProperties.sparkline;

  $: minWidth = getMinWidth(sparkline);
</script>

{#if schema.isValid}
  <div class="h-fit p-0 grow relative" class:!p-0={kpis.length === 1}>
    <span class="border-overlay" />
    <div
      style:grid-template-columns="repeat(auto-fit, minmax(min({minWidth}px,
      100%), 1fr))"
      class="grid-wrapper gap-px overflow-hidden size-full"
    >
      {#each kpis as kpi, i (i)}
        <div class="min-h-32 kpi-wrapper">
          <KPIProvider spec={kpi} {timeAndFilterStore} {canvasName} />
        </div>
      {/each}
    </div>
  </div>
{:else}
  <ComponentError error={schema.error} />
{/if}

<style lang="postcss">
  .grid-wrapper {
    @apply size-full grid;
    grid-auto-rows: auto;
  }

  .kpi-wrapper {
    @apply relative p-4 grid outline outline-1 outline-gray-200;
  }

  .border-overlay {
    @apply absolute border-[12.5px] pointer-events-none border-white size-full;
    z-index: 50;
  }

  @container component-container (inline-size < 440px) {
    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
