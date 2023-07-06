<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import WithBisector from "@rilldata/web-common/components/data-graphic/functional-components/WithBisector.svelte";
  import { Grid } from "@rilldata/web-common/components/data-graphic/guides";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import MultiMetricMouseoverLabel from "@rilldata/web-common/components/data-graphic/marks/MultiMetricMouseoverLabel.svelte";
  import { humanizeDataType } from "@rilldata/web-common/features/dashboards/humanize-numbers";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import {
    MetricsViewMeasure,
    createQueryServiceMetricsViewTimeSeries,
    createQueryServiceMetricsViewTotals,
    createRuntimeServiceGetCatalogEntry,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";

  export let metricsDefName: string;

  $: measuresQuery = createRuntimeServiceGetCatalogEntry(
    $runtime?.instanceId,
    metricsDefName,
    {
      query: {
        keepPreviousData: true,
      },
    }
  );
  let measures: MetricsViewMeasure[] = [];
  $: if ($measuresQuery?.isSuccess)
    measures = $measuresQuery?.data?.entry?.metricsView?.measures || [];
  $: console.log(measures);
  $: measureNames = measures?.map((measure) => measure.name) || [];

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    $runtime?.instanceId,
    metricsDefName,
    {
      measureNames,
      filter: {},
    },
    { query: { enabled: $measuresQuery?.isSuccess } }
  );

  $: timeSeriesQuery = createQueryServiceMetricsViewTimeSeries(
    $runtime.instanceId,
    metricsDefName,
    {
      measureNames,
      filter: {},
    },
    { query: { enabled: $measuresQuery?.isSuccess, keepPreviousData: true } }
  );
  let formattedData;

  $: if ($timeSeriesQuery?.isSuccess)
    formattedData = $timeSeriesQuery?.data?.data?.map((data) => ({
      ...data,
      ...Object.keys(data.records).reduce((acc, v) => {
        acc[v] = data.records[v];
        return acc;
      }, {}),
      ts: new Date(data.ts),
    }));

  $: xMin = formattedData?.[0]?.ts;
  $: xMax = formattedData?.[formattedData.length - 1]?.ts;

  let showMeasures = true;

  let measureSize = "lg";
  function setMeasureSize(size) {
    measureSize = size;
  }

  const { observedNode, listenToNodeResize } =
    createResizeListenerActionFactory();
  $: width = $observedNode?.getBoundingClientRect()?.width;

  let mouseoverValue;
  $: console.log(mouseoverValue);
</script>

<div class="pl-4 pb-4 pr-4" use:listenToNodeResize>
  <CollapsibleSectionTitle tooltipText="measures" bind:active={showMeasures}>
    Measures
    <svelte:fragment slot="contextual-information">
      <button
        on:click={() => {
          setMeasureSize("sm");
        }}>sm</button
      >
      <button
        on:click={() => {
          setMeasureSize("lg");
        }}>lg</button
      >
    </svelte:fragment>
  </CollapsibleSectionTitle>
</div>

{#if showMeasures}
  {#if formattedData}
    <WithBisector
      data={formattedData}
      value={mouseoverValue?.x}
      callback={(pt) => pt.ts}
      let:point
    >
      {#each measures as measure (measure.name)}
        {@const y = formattedData?.map((d) => d[measure.name])}
        {@const yMax = Math.max(...y)}
        {@const yMin = Math.min(...y, 0)}
        {@const tallestPoint = formattedData?.reduce((acc, v) => {
          const pt = { ...v };
          if (pt[measure.name] === yMax) {
            return pt;
          }
          return acc;
        }, formattedData[0])}

        {#if measureSize === "sm"}
          <div
            class="pop px-4 flex gap-x-4 items-center"
            transition:slide={{ duration: LIST_SLIDE_DURATION }}
          >
            <div class="grow truncate">{measure?.label || measure?.name}</div>
            <div class="text-right font-medium ui-copy-number">
              {humanizeDataType(
                $totalsQuery?.data?.data[measure.name],
                measure?.format
              )}
            </div>
            <SimpleDataGraphic
              width={80}
              height={19}
              left={0}
              right={0}
              top={0}
              bottom={0}
              xType="date"
              yType="number"
              {yMin}
              {yMax}
              {xMin}
              {xMax}
            >
              <ChunkedLine
                data={formattedData}
                xAccessor="ts"
                yAccessor={measure.name}
              />
            </SimpleDataGraphic>
          </div>
        {:else if measureSize === "lg"}
          <div
            class="flex gap-x-4"
            transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
          >
            <div class="px-4 truncate" style:width="120px">
              <div class="grow">{measure?.label || measure?.name}</div>
              <div class=" font-medium ui-copy-number" style:width="72px">
                {humanizeDataType(
                  $totalsQuery?.data?.data[measure.name],
                  measure?.format
                )}
              </div>
            </div>
            <SimpleDataGraphic
              width={width - 120 - 28}
              height={40}
              left={2}
              right={2}
              top={0}
              bottom={0}
              xType="date"
              yType="number"
              {yMin}
              {yMax}
              {xMin}
              {xMax}
              bind:mouseoverValue
            >
              <Grid />
              <ChunkedLine
                data={formattedData}
                xAccessor="ts"
                yAccessor={measure.name}
              />
              <MultiMetricMouseoverLabel
                formatValue={(v) => humanizeDataType(v, measure.format)}
                showPointLabels={false}
                xBuffer={4}
                point={[
                  {
                    key: "pt",
                    x: point?.ts || tallestPoint.ts,
                    y: point?.[measure.name] || tallestPoint[measure.name],
                    valueStyleClass: point?.ts
                      ? "font-semibold"
                      : "font-regular",
                    pointColorClass: "fill-blue-500",
                  },
                ]}
              />
            </SimpleDataGraphic>
          </div>
        {/if}
      {/each}
    </WithBisector>
  {/if}
{/if}
