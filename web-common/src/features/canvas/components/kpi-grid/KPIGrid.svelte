<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import type { KPIGridSpec } from ".";
  import type { KPISpec } from "../kpi";
  import KPI from "../kpi/KPI.svelte";
  import { validateKPIGridSchema } from "./selector";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  let kpis: KPISpec[];

  $: kpiGridProperties = rendererProperties as KPIGridSpec;
  $: schema = validateKPIGridSchema(kpiGridProperties);

  // Convert measures to KPI specs
  $: kpis = (kpiGridProperties.measures || []).map((measure) => ({
    metrics_view: kpiGridProperties.metrics_view,
    measure,
    sparkline: kpiGridProperties.sparkline,
    comparison: kpiGridProperties.comparison,
  }));

  // Calculate individual KPI width based on container width and number of KPIs
  // $: kpiWidth = containerWidth ? Math.floor(containerWidth / kpis.length) : 0;
</script>

{#if schema.isValid}
  <div
    class="element h-fit p-4"
    style:grid-template-columns="repeat({kpis.length}, 1fr)"
  >
    {#each kpis as kpi, i (i)}
      <div
        class="kpi-wrapper border-gray-200 size-full min-h-40 overflow-hidden p-4"
      >
        <KPI rendererProperties={kpi} topPadding={false} />
      </div>
    {/each}
  </div>
{:else}
  <ComponentError error={schema.error} />
{/if}

<style lang="postcss">
  /* Add any custom styles here if needed */

  .element {
    @apply size-full grid;

    @apply px-0;
  }

  .kpi-wrapper {
    @apply py-0;
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
      grid-template-columns: repeat(2, 1fr);
    }

    /* remove border for second item */
    .kpi-wrapper:nth-child(2) {
      border-right-width: 0;
      border-bottom-width: 1px;
    }

    .kpi-wrapper:nth-child(1) {
      border-bottom-width: 1px;
    }

    .element {
      padding: 16px;
    }

    /* first two hsould have bottom padding of 16px */
    .kpi-wrapper:nth-child(1),
    .kpi-wrapper:nth-child(2) {
      padding-bottom: 16px;
    }

    .kpi-wrapper:nth-child(3),
    .kpi-wrapper:nth-child(4) {
      padding-top: 16px;
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
