<!--
@component
Takes in a bisection callback, data, and a value, and returns the nearest point in the data
to the value.

Useful for finding the nearest value to the current mouseover
-->
<script lang="ts">
  import { bisector } from "d3-array";
  /** the dataset that will be used for bisection */
  export let data: ArrayLike<any>;
  /** The callback function that returns the value for bisection */
  export let callback: (arg0: any) => number | Date;
  /** The value that will be used for the bisection */
  export let value: any;
  const bisect = bisector(callback).center;

  /** provide a bind site for the output */
  export let point = undefined;

  $: point = value !== undefined ? data[bisect(data, value)] : undefined;
</script>

<slot {point} />
