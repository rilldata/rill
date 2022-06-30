<script lang="ts">
  import { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import type { Readable } from "svelte/store";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";

  export let metricsDefId: string;

  let timeSeries: Readable<TimeSeriesEntity>;
  $: if (metricsDefId) {
    timeSeries = getTimeSeriesById(metricsDefId);
  }
</script>

{#if $timeSeries?.values}
  <TimestampSpark
    data={convertTimestampPreview($timeSeries.values)}
    xAccessor="ts"
    yAccessor="count"
    width={800}
    height={300}
    top={0}
    bottom={0}
    left={0}
    right={0}
    leftBuffer={0}
    rightBuffer={0}
    area
    tweenIn
  />
{/if}
