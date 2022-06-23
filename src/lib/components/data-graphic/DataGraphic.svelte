<script lang="ts">
  import {
    initializeCartesianScales,
    initializePlotConfigs,
  } from "./initializers";

  export let top = 40;
  export let bottom = 40;
  export let left = 42;
  export let right = 42;
  export let buffer = 0;
  export let width = 360;
  export let height = 120;
  export let fontSize = 12;
  export let textGap = 4;
  export let xType = undefined;
  export let yType = undefined;

  export let xMin = undefined;
  export let xMax = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  const plotConfig = initializePlotConfigs();
  $: plotConfig.set({
    width,
    height,
    top,
    bottom,
    left,
    right,
    buffer,
    fontSize,
    textGap,
  });

  const { yMin: yMinStore } = initializeCartesianScales({
    xMinValue: xMin,
    xMaxValue: xMax,
    yMinValue: yMin,
    yMaxValue: yMax,
    xType,
    yType,
  });
</script>

{$yMinStore}
<svg {width} {height}>
  <slot plotConfig={$plotConfig} />
</svg>
