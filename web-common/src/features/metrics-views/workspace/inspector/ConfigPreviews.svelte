<script lang="ts">
  import {
    useMetaQuery,
    useModelAllTimeRange,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { datesToFormattedTimeRange } from "@rilldata/web-common/lib/formatters";
  import {
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import DimensionPreview from "./DimensionPreview.svelte";
  import MeasurePreview from "./MeasurePreview.svelte";
  $: instanceId = $runtime.instanceId;

  export let metricsDefName;
  export let modelName;
  $: metaQuery = useMetaQuery(instanceId, metricsDefName);
  let allTimeRangeQuery;

  // FIXME: fix this useModelAllTimeRange to be dependent on other query params.
  $: if ($metaQuery?.data?.timeDimension) {
    allTimeRangeQuery = useModelAllTimeRange(
      $runtime.instanceId,
      modelName,
      $metaQuery?.data?.timeDimension
    );
  }

  $: start = $allTimeRangeQuery?.data?.start;
  $: end = $allTimeRangeQuery?.data?.end;

  /** get the big numbers */
  let totalsQuery;
  $: totalsQueryParams = {
    measureNames: $metaQuery?.data?.measures?.map((m) => m.name),
    //filter: { include: [], exclude: [] },
    //timeStart: start?.toISOString(), //$dashboardStore.selectedTimeRange?.start.toISOString(),
    //timeEnd: end?.toISOString(), //$dashboardStore.selectedTimeRange?.end.toISOString(),
  };
  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricsDefName,
    totalsQueryParams,
    {
      query: { enabled: $metaQuery?.data?.measures?.length > 0 },
    }
  );

  // }

  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    instanceId,
    metricsDefName,
    {
      measureNames: $metaQuery?.data?.measures?.map((m) => m.name),
      filter: { include: [], exclude: [] }, //$dashboardStore?.filters,
      // timeStart: $dashboardStore.selectedTimeRange?.start.toISOString(),
      // timeEnd: $dashboardStore.selectedTimeRange?.end.toISOString(),
      // timeGranularity: $dashboardStore.selectedTimeRange?.interval,
    }
  );

  let mouseoverValue;

  let showMeasures = true;
  let showDimensions = true;
</script>

<div>
  {#if $metaQuery?.data && $allTimeRangeQuery?.data && start && end}
    {@const formattedTimeRange = datesToFormattedTimeRange(start, end)}
    <div class="px-4 py-4">{formattedTimeRange} of data</div>
  {/if}

  {#if $metaQuery?.data?.measures}
    <div>
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="measures"
          bind:active={showMeasures}
        >
          Measures
        </CollapsibleSectionTitle>
      </div>
      {#if showMeasures}
        <ul
          class="p-4 w-full"
          transition:slide={{ duration: LIST_SLIDE_DURATION }}
        >
          {#each $metaQuery?.data?.measures as measure (measure?.expression + measure?.name)}
            {@const value = $totalsQuery?.data?.data?.[measure.name]}
            {@const trend = $timeSeriesQuery?.data?.data?.map((di) => {
              const pi = {
                value: di.records[measure.name],
                ts: new Date(di.ts),
              };
              return pi;
            })}
            <li class="w-full">
              <MeasurePreview
                label={measure.label}
                format={measure.format}
                {value}
                {trend}
                {start}
                {end}
                bind:mouseoverValue
              />
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {/if}

  {#if $metaQuery?.data?.dimensions}
    <div>
      <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle
          tooltipText="dimensions"
          bind:active={showDimensions}
        >
          Dimensions
        </CollapsibleSectionTitle>
      </div>
      {#if showDimensions}
        <ul
          class="space-y-2 w-full"
          transition:slide={{ duration: LIST_SLIDE_DURATION }}
        >
          {#each $metaQuery?.data?.dimensions as dimension}
            <DimensionPreview dimensionName={dimension.name} {metricsDefName} />
          {/each}
        </ul>
      {/if}
    </div>
  {/if}
</div>
