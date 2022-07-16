<script lang="ts">
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { fly } from "svelte/transition";
  import MeasureBigNumber from "./MeasureBigNumber.svelte";
  import TimeSeriesBody from "./TimeSeriesBody.svelte";

  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";

  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithBisector } from "$lib/components/data-graphic/functional-components";
  import { Axis } from "$lib/components/data-graphic/guides";
  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import TimeSeriesChartContainer from "./TimeSeriesChartContainer.svelte";

  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { Readable } from "svelte/store";
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import type { BigNumberEntity } from "$lib/redux-store/big-number/big-number-slice";
  import CrossIcon from "$lib/components/icons/CrossIcon.svelte";
  import Spinner from "$lib/components/Spinner.svelte";

  export let metricsDefId;
  export let start: Date;
  export let end: Date;

  // get all the measure ids that are available.

  $: allMeasures = getMeasuresByMetricsId(metricsDefId);

  // get the active big numbers

  let bigNumbers: Readable<BigNumberEntity>;
  $: bigNumbers = getBigNumberById(metricsDefId);
  // plot the data

  let timeSeries: Readable<TimeSeriesEntity>;
  $: timeSeries = getTimeSeriesById(metricsDefId);
  $: formattedData = $timeSeries?.values
    ? convertTimestampPreview($timeSeries.values, true)
    : undefined;

  let mouseoverValue = undefined;

  function initializeToMidnight(dt) {
    let newDt = new Date(dt);
    newDt.setHours(0, 0, 0, 0);
    return newDt;
  }

  $: key = `${start}` + `${end}`;

  $: startValue = initializeToMidnight(new Date(start));
  $: endValue = new Date(end);
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
      <!-- FIXME: I can't select the big number by the measure id.
    -->
      {@const bigNum = $bigNumbers?.bigNumbers?.[`measure_${index}`]}

      <!-- FIXME: I can't select a time series by measure id. 
    -->
      <MeasureBigNumber
        value={bigNum}
        description={measure?.description ||
          measure?.label ||
          measure?.expression}
        formatPreset={measure?.formatPreset}
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
            formatPreset={measure?.formatPreset}
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
