<script lang="ts">
  import { scaleLinear, scaleTime } from "d3-scale";

  export let xExtents: [Date, Date];
  export let yExtents: [number, number];
  export let xScale = scaleTime();
  export let yScale = scaleLinear();

  $: yScale = scaleLinear()
    .domain([0, yExtents[1] * 1.15])
    .range([100, 0]);
  $: xScale = scaleTime().domain(xExtents).range([0, 100]);
</script>

<slot {xScale} {yScale} />
