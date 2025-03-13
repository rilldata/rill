<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { KPIGridSpec } from ".";
  import ComponentError from "../ComponentError.svelte";
  import type { KPISpec } from "../kpi";
  import KPI from "../kpi/KPI.svelte";
  import { validateKPIGridSchema } from "./selector";
  import { getMinWidth } from "../kpi";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;

  let kpis: KPISpec[];

  $: kpiGridProperties = rendererProperties as KPIGridSpec;
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
        <div
          class="min-h-32 kpi-wrapper before:absolute before:z-20 before:top-full before:h-px before:w-full before:bg-gray-200 after:absolute after:left-full after:h-full after:w-px after:bg-gray-200"
        >
          <KPI rendererProperties={kpi} {timeAndFilterStore} />
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
    @apply relative p-4 grid;
  }

  .border-overlay {
    @apply border-[16px] pointer-events-none border-white absolute size-full z-50;
  }

  @container component-container (inline-size < 440px) {
    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
