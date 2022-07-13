<script lang="ts">
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import MeasureBigNumber from "./time-series-charts/MeasureBigNumber.svelte";
  import TimeSeriesBody from "./time-series-charts/TimeSeriesBody.svelte";

  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";

  import { getBigNumberById } from "$lib/redux-store/big-number/big-number-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import TimeSeriesChartContainer from "./time-series-charts/TimeSeriesChartContainer.svelte";
  import { WithBisector } from "$lib/components/data-graphic/functional-components";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { Axis } from "$lib/components/data-graphic/guides";

  import Spinner from "$lib/components/Spinner.svelte";
  import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  export let metricsDefId;
  export let activeMeasureIds: string[] = [];
  export let start;
  export let end;
  export let interval;

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

  $: key = start + end;

  $: hovering = !!mouseoverValue?.x;
</script>

<WithBisector
  data={formattedData}
  callback={(datum) => datum.ts}
  value={mouseoverValue?.x}
  let:point
>
  <TimeSeriesChartContainer start={new Date(start)} end={new Date(end)}>
    <div />
    <!-- add the axis component -->
    <SimpleDataGraphic height={40} top={24} bottom={0} let:xScale>
      <Axis side="top" />
    </SimpleDataGraphic>
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
      >
        <svelte:fragment slot="name">
          {measure?.label || measure?.expression}
        </svelte:fragment>
      </MeasureBigNumber>
      <div class="time-series-body" style:height="125px">
        {#if formattedData}
          <TimeSeriesBody
            bind:mouseoverValue
            formatPreset={measure?.formatPreset}
            data={formattedData}
            accessor={`measure_${index}`}
            mouseover={point}
            {key}
            start={new Date(start)}
            end={new Date(end)}
            {interval}
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
