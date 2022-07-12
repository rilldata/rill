<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import MetricsExploreTimeChart from "$lib/components/leaderboard/MetricsExploreTimeChart.svelte";
  import type { Readable } from "svelte/store";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getFallbackMeasureName } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  // import MeasureBigNumber from "$lib/components/leaderboard/MeasureBigNumber.svelte";
  import MeasureBigNumber from "./time-series-charts/MeasureBigNumber.svelte";
  import TimeSeriesBody from "./time-series-charts/TimeSeriesBody.svelte";

  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";

  import {
    getBigNumbersByIds,
    getBigNumberById,
  } from "$lib/redux-store/big-number/big-number-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import TimeSeriesChartContainer from "./time-series-charts/TimeSeriesChartContainer.svelte";
  import {
    WithBisector,
    WithTween,
  } from "$lib/components/data-graphic/functional-components";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "$lib/components/data-graphic/guides";

  export let metricsDefId;
  export let activeMeasureIds: string[] = [];

  // get all the measure ids that are available.

  $: allMeasures = getMeasuresByMetricsId(metricsDefId);

  // get the active big numbers

  $: bigNumbers = getBigNumberById(metricsDefId);
  // plot the data

  $: timeSeries = getTimeSeriesById(metricsDefId);
  $: formattedData = $timeSeries?.values
    ? convertTimestampPreview($timeSeries.values)
    : undefined;

  let mouseoverValue = undefined;
</script>

{#if formattedData}
  <WithTween
    value={formattedData}
    let:output={tweenedFormattedData}
    tweenProps={{ duration: 500, easing: cubicOut }}
  >
    <WithBisector
      data={tweenedFormattedData}
      callback={(datum) => datum.ts}
      value={mouseoverValue?.x}
      let:point
    >
      <TimeSeriesChartContainer>
        {#if $bigNumbers}
          <div />
          <!-- add the axis component -->
          <SimpleDataGraphic height={42} top={24} bottom={4}>
            <Axis side="top" />
          </SimpleDataGraphic>
          {#each $allMeasures as measure, index (measure.id)}
            <!-- FIXME: I can't select the big number by the measure id.
    -->
            {@const bigNum = $bigNumbers.bigNumbers[`measure_${index}`]}
            <!-- FIXME: I can't select a time series by measure id. 
    -->
            <MeasureBigNumber value={bigNum}>
              <svelte:fragment slot="name">
                {measure.label || measure.expression}
              </svelte:fragment>
            </MeasureBigNumber>

            {#if formattedData}
              <TimeSeriesBody
                bind:mouseoverValue
                data={tweenedFormattedData.slice(-200, -100)}
                accessor={`measure_${index}`}
                mouseover={point}
              />
            {/if}
          {/each}
        {/if}
      </TimeSeriesChartContainer>
    </WithBisector>
  </WithTween>
{/if}
