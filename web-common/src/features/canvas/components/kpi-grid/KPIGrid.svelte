<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import type { KPIGridComponent } from ".";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import {
    getMinWidth,
    getOptimalColumns,
    resolveKPISize,
    type KPISpec,
  } from "../kpi";
  import KPIProvider from "../kpi/KPIProvider.svelte";
  import { canvasExactSurface } from "../exact-measures";
  import { validateKPIGridSchema } from "./selector";
  import type { KpiExportValue } from "@rilldata/web-common/features/exports/xlsx/payload";

  export let component: KPIGridComponent;

  let kpis: KPISpec[];

  // The values each KPI reports, by position, for the workbook export; see
  // BaseCanvasComponent.exportPayload.
  let kpiValues: (KpiExportValue | undefined)[] = [];
  function setKpiValue(index: number, value: KpiExportValue | undefined) {
    if (kpiValues[index] === value) return;
    const next = kpiValues.slice();
    next[index] = value;
    kpiValues = next;
  }

  $: ({
    specStore,
    timeAndFilterStore,
    parent: { name: canvasName },
    dataEnabled: visible,
  } = component);
  $: kpiGridProperties = $specStore;
  $: schema = validateKPIGridSchema(kpiGridProperties);

  // Convert measures to KPI specs
  $: kpis = (kpiGridProperties.measures || []).map((measure) => ({
    metrics_view: kpiGridProperties.metrics_view,
    measure,
    adhoc_measures: kpiGridProperties.adhoc_measures,
    sparkline: kpiGridProperties.sparkline,
    size: kpiGridProperties.size,
    hide_time_range: kpiGridProperties.hide_time_range,
    comparison: kpiGridProperties.comparison,
    dimension_filters: kpiGridProperties.dimension_filters,
    time_filters: kpiGridProperties.time_filters,
  }));

  $: filters = {
    time_filters: kpiGridProperties.time_filters,
    dimension_filters: kpiGridProperties.dimension_filters,
  };

  $: component.exportPayload.set(
    kpis.length && kpis.every((_, i) => kpiValues[i])
      ? {
          kind: "kpi",
          values: kpiValues.slice(0, kpis.length) as KpiExportValue[],
        }
      : undefined,
  );

  $: sparkline = kpiGridProperties.sparkline;
  $: size = resolveKPISize(kpiGridProperties.size);

  $: minWidth = getMinWidth(sparkline, size);

  let containerWidth = 0;

  $: optimalColumns = getOptimalColumns(kpis.length, containerWidth, minWidth);

  $: ({ title, description, show_description_as_tooltip } = kpiGridProperties);
</script>

<ComponentHeader
  {component}
  {title}
  {description}
  showDescriptionAsTooltip={show_description_as_tooltip}
  {filters}
/>

{#if schema.isValid}
  <div
    class="h-fit p-0 grow relative"
    class:!p-0={kpis.length === 1}
    data-size={size}
    bind:clientWidth={containerWidth}
  >
    <span class="border-overlay"></span>
    <div
      style:grid-template-columns="repeat({optimalColumns}, minmax(0, 1fr))"
      class="grid-wrapper gap-px overflow-hidden size-full"
    >
      {#each kpis as kpi, i (i)}
        <div class="kpi-wrapper">
          {#if $timeAndFilterStore}
            <KPIProvider
              spec={kpi}
              {timeAndFilterStore}
              {canvasName}
              visible={$visible}
              surface={canvasExactSurface(component.id)}
              onValues={(value: KpiExportValue | undefined) =>
                setKpiValue(i, value)}
            />
          {/if}
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
    @apply relative p-4 min-h-32 grid outline outline-1 outline-border;
  }

  .border-overlay {
    @apply absolute border-[12.5px] pointer-events-none border-surface-card size-full;
    z-index: 50;
  }

  /*
   * Size presets: the card's padding and minimum height follow the text sizes
   * KPI.svelte applies for the same preset.
   */
  [data-size="small"] .kpi-wrapper {
    @apply p-3 min-h-20;
  }

  /* The overlay covers the cards' outer padding, so it shrinks with it. */
  [data-size="small"] .border-overlay {
    @apply border-[8.5px];
  }

  [data-size="large"] .kpi-wrapper {
    @apply p-5 min-h-40;
  }

  [data-size="xlarge"] .kpi-wrapper {
    @apply p-6 min-h-48;
  }

  @container component-container (inline-size < 440px) {
    .grid-wrapper {
      grid-template-columns: repeat(1, 1fr) !important;
    }
  }
</style>
