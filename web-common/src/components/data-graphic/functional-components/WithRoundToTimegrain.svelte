<!-- @component 
Rounds a Date object to the nearest time unit. This is useful to prevent
mouseover Dates from constantly pushing updates to the bisector; when this
functional component feeds its slot prop into the bisector, it should only give

-->
<script lang="ts">
  import { roundToNearestTimeUnit } from "@rilldata/web-common/features/dashboards/time-series/round-to-nearest-time-unit";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  export let timeGrain: V1TimeGrain;
  export let value: Date;

  let prev;
  $: roundedValue = roundToNearestTimeUnit(value, TIME_GRAIN[timeGrain].label);
  $: if (+prev !== +roundedValue) {
    prev = roundedValue;
  }
</script>

<slot roundedValue={prev} />
