<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewAggregation,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Readable } from "svelte/motion";
  import type { GaugeSpec } from ".";
  import { Gauge } from ".";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { validateGaugeSchema } from "./selector";

  export let spec: GaugeSpec;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let canvasName: string;
  export let visible: boolean;

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: ({
    metricsView: { getMeasureForMetricView, getMetricsViewFromName },
  } = ctx.canvasEntity);

  $: ({ instanceId } = $runtime);

  $: ({
    metrics_view: metricsViewName,
    measure: measureName,
  } = spec);

  $: ({
    timeRange: { timeZone, start, end },
    where,
  } = $timeAndFilterStore);

  $: schema = validateGaugeSchema(ctx, spec);
  $: ({ isValid } = $schema);

  $: measureStore = getMeasureForMetricView(measureName, metricsViewName);
  $: measure = $measureStore;

  $: metricsViewQuery = getMetricsViewFromName(metricsViewName);
  $: metricsViewSpec = $metricsViewQuery?.metricsView;

  // Check if the measure has targets configured
  $: hasTargets = Boolean(
    metricsViewSpec?.targets?.some((target) =>
      target.measures?.includes(measureName),
    ),
  );

  $: queryMeasures = [{ name: measureName }];

  $: totalQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: queryMeasures,
      timeRange: {
        start,
        end,
        timeZone,
      },
      where,
      priority: 50,
      includeTargets: hasTargets,
    },
    {
      query: {
        enabled: isValid && visible && (!!start && !!end),
      },
    },
  );
</script>

<Gauge
  {measure}
  totalResult={$totalQuery}
/>

