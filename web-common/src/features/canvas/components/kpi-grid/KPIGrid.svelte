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
  <div class="element h-fit" style:--item-count={kpis.length}>
    {#each kpis as kpi, i (i)}
      <div
        class:solo={kpis.length > 1}
        class="kpi-wrapper border-gray-200 size-full min-h-52 p-4 overflow-hidden"
      >
        <KPI rendererProperties={kpi} {timeAndFilterStore} />
      </div>
    {/each}
  </div>
{:else}
  <ComponentError error={schema.error} />
{/if}

<style lang="postcss">
  .element {
    @apply size-full grid;
    @apply px-0;
    grid-template-columns: repeat(var(--item-count), 1fr);
  }

  .kpi-wrapper {
    @apply w-full;
  }

  .kpi-wrapper:not(:last-of-type) {
    @apply border-r;
  }

  .element {
    container-type: inline-size;
    container-name: container;
  }

  @container container (inline-size < 600px) {
    .element {
      grid-template-columns: repeat(min(2, var(--item-count)), 1fr);
    }

    .kpi-wrapper:nth-child(2) {
      border-right-width: 0;
      border-bottom-width: 1px;
    }

    .kpi-wrapper.solo:nth-child(1) {
      border-bottom-width: 1px;
    }

    .kpi-wrapper:nth-child(3) {
      border-right-width: 1px;
    }
  }

  @container container (inline-size < 300px) {
    .element {
      grid-template-columns: repeat(1, 1fr);
    }

    .kpi-wrapper {
      border-right-width: 0 !important;
    }

    .kpi-wrapper:not(:last-of-type) {
      border-bottom-width: 1px;
    }
  }
</style>
