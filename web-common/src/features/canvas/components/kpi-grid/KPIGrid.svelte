<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { KPIGridSpec } from ".";
  import ComponentError from "../ComponentError.svelte";
  import type { KPISpec } from "../kpi";
  import KPI from "../kpi/KPI.svelte";
  import { validateKPIGridSchema } from "./selector";

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
</script>

{#if schema.isValid}
  <div class="h-fit p-4 grow" style:--item-count={kpis.length}>
    <div class="grid-wrapper gap-px overflow-hidden size-full min-h-32">
      {#each kpis as kpi, i (i)}
        <div
          class="kpi-wrapper before:absolute before:top-full before:h-px before:w-full before:bg-gray-200 after:absolute after:left-full after:h-full after:w-px after:bg-gray-200"
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
    grid-template-columns: repeat(auto-fit, minmax(min(200px, 100%), 1fr));
    grid-auto-rows: auto;
  }

  .kpi-wrapper {
    @apply relative p-4 grid;
  }

  @container component-container (inline-size < 600px) {
    .kpi-wrapper:nth-of-type(odd) {
      padding-left: 0px;
    }

    .kpi-wrapper:nth-of-type(even) {
      padding-right: 0px;
    }

    .grid-wrapper {
      grid-template-columns: repeat(min(2, var(--item-count)), 1fr);
    }
  }

  @container component-container (inline-size < 440px) {
    .kpi-wrapper {
      padding-left: 0px;
      padding-right: 0px;
    }

    .kpi-wrapper:last-of-type {
      padding-bottom: 0px;
    }

    .kpi-wrapper:first-of-type {
      padding-top: 0px;
    }

    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr);
    }
  }
</style>
