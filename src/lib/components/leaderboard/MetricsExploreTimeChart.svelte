<script lang="ts">
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import type { Readable } from "svelte/store";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";

  export let metricsDefId: string;
  export let yAccessor: string;

  let timeSeries: Readable<TimeSeriesEntity>;
  $: if (metricsDefId) {
    timeSeries = getTimeSeriesById(metricsDefId);
  }
</script>

{#if $timeSeries?.values}
  <TimestampSpark
    data={convertTimestampPreview($timeSeries.values)}
    xAccessor="ts"
    {yAccessor}
    width={345}
    height={120}
    top={0}
    bottom={0}
    left={0}
    right={45}
    leftBuffer={0}
    rightBuffer={0}
    area
    tweenIn
  />
{/if}
