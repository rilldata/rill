<!--
@component
Takes in a bisection callback, data, and a value, and returns the nearest point in the data
to the value.

Useful for finding the nearest value to the current mouseover
-->
<script lang="ts">
  import { bisector } from "d3-array";

  type T = $$Generic;
  type K = $$Generic;
  /** the dataset that will be used for bisection */
  export let data: T[];
  /** The callback function that returns the value for bisection */
  export let callback: (arg0: T) => K;
  /** The value that will be used for the bisection */
  export let value: K;
  /** The direction that will be used for the bisection */
  export let direction: "left" | "right" | "center" = "left";

  const bisect = bisector(callback)[direction];

  /** provide a bind site for the output */
  export let point: T | undefined = undefined;

  $: point = value !== undefined ? data[bisect(data, value)] : undefined;
</script>

<slot {point} />
