<script lang="ts">
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "$lib/components/data-graphic/functional-components";
  import { Axis } from "$lib/components/data-graphic/guides";
  import CrossIcon from "$lib/components/icons/CrossIcon.svelte";
  import Spinner from "$lib/components/Spinner.svelte";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import { getValidMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import type {
    TimeSeriesEntity,
    TimeSeriesValue,
  } from "$lib/redux-store/timeseries/timeseries-slice";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { removeTimezoneOffset } from "$lib/util/formatters";
  import { NicelyFormattedTypes } from "$lib/util/humanize-numbers";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import TimeSeriesBody from "./TimeSeriesBody.svelte";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  export let metricsDefId;

  $: allMeasures = getValidMeasuresByMetricsId(metricsDefId);

  let bigNumbers: Readable<BigNumberEntity>;
  $: bigNumbers = getBigNumberById(metricsDefId);

  let timeSeries: Readable<TimeSeriesEntity>;
  $: timeSeries = getTimeSeriesById(metricsDefId);
  $: formattedData = $timeSeries?.values
    ? convertTimestampPreview($timeSeries.values, true)
    : undefined;

  let mouseoverValue = undefined;

  $: key = `${startValue}` + `${endValue}`;

  function getMinTs(values: TimeSeriesValue[]): Date {
    if (!values) return new Date();
    let min = new Date(values[0].ts);
    for (let i = 1; i < values.length; i++) {
      if (new Date(values[i].ts).getTime() < min.getTime()) {
        min = new Date(values[i].ts);
      }
    }
    return min;
  }
  function getMaxTs(values: TimeSeriesValue[]): Date {
    if (!values) return new Date();
    let max = new Date(values[0].ts);
    for (let i = 1; i < values.length; i++) {
      if (new Date(values[i].ts).getTime() > max.getTime()) {
        max = new Date(values[i].ts);
      }
    }
    return max;
  }
  $: minDate = getMinTs($timeSeries?.values);
  $: maxDate = getMaxTs($timeSeries?.values);
  $: startValue = removeTimezoneOffset(minDate);
  $: endValue = removeTimezoneOffset(maxDate);
</script>

<WithBisector
  data={formattedData}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer start={startValue} end={endValue}>
    <!-- mouseover date elements-->

    <div />
    <div style:padding-left="24px">
      {#if point?.ts}
        <div
          class="absolute italic text-gray-600"
          transition:fly|local={{ duration: 100, y: 4 }}
        >
          {new Intl.DateTimeFormat("en-US", {
            dateStyle: "medium",
            timeStyle: "medium",
          }).format(point?.ts)}
        </div>
        &nbsp;
      {:else}
        &nbsp;
      {/if}
    </div>
    <!-- top axis element -->
    <div />
    <SimpleDataGraphic
      height={32}
      top={34}
      bottom={0}
      xMin={startValue}
      xMax={endValue}
    >
      <Axis superlabel side="top" />
    </SimpleDataGraphic>
    <!-- bignumbers and line charts -->
    {#each $allMeasures as measure, index (measure.id)}
      <!-- FIXME: I can't select the big number by the measure id. -->
      {@const bigNum = $bigNumbers?.bigNumbers?.[`measure_${index}`]}

      <!-- FIXME: I can't select a time series by measure id. -->
      <MeasureBigNumber
        value={bigNum}
        description={measure?.description ||
          measure?.label ||
          measure?.expression}
        formatPreset={measure?.formatPreset || NicelyFormattedTypes.HUMANIZE}
        status={$bigNumbers?.status}
      >
        <svelte:fragment slot="name">
          {measure?.label || measure?.expression}
        </svelte:fragment>
      </MeasureBigNumber>
      <div class="time-series-body" style:height="125px">
        {#if $timeSeries?.status === EntityStatus.Error}
          <div class="p-5"><CrossIcon /></div>
        {:else if formattedData}
          <TimeSeriesBody
            bind:mouseoverValue
            formatPreset={measure?.formatPreset ||
              NicelyFormattedTypes.HUMANIZE}
            data={formattedData}
            accessor={`measure_${index}`}
            mouseover={point}
            {key}
            start={startValue}
            end={endValue}
          />
        {:else}
          <div>
            <Spinner status={EntityStatus.Running} />
          </div>
        {/if}
      </div>
    {/each}
  </TimeSeriesChartContainer>
</WithBisector>
