<script lang="ts">
  import { reduxReadable } from "$lib/redux-store/store-root";
  import { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import { selectTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-selectors";
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";

  export let metricsDefId: string;

  let timeSeries: TimeSeriesEntity;
  $: if (metricsDefId) {
    timeSeries = selectTimeSeriesById(metricsDefId)($reduxReadable);
  }
</script>

{#if timeSeries?.values}
  <TimestampSpark
    data={convertTimestampPreview(timeSeries.values)}
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
