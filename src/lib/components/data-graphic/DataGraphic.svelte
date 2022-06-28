<script lang="ts">
  import { getContext } from "svelte";
  import {
    initializeCartesianScales,
    initializePlotConfigs,
    initializeScaleContext,
  } from "./state/initializers";
  import type { PlotConfig } from "./utils";

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

  let plotConfigContext = getContext("rill:data-graphic:plot-config");
  let plotConfig;
  if (plotConfigContext === undefined) {
    plotConfig = initializePlotConfigs();
  } else {
    plotConfig = plotConfigContext;
  }
  $: if (plotConfigContext === undefined) {
    plotConfig.set({
      width,
      height,
      top,
      bottom,
      left,
      right,
      buffer,
      fontSize,
      textGap,
      xType,
      yType,
    });
  }

  initializeScaleContext({
    type: xType,
    namespace: "x",
    // overriding domain min & max if these values are set
    domainMin: xMin,
    domainMax: xMax,
    rangeMin: (config: PlotConfig) => config.plotLeft,
    rangeMax: (config: PlotConfig) => config.plotRight,
  });

  // initialize Y
  initializeScaleContext({
    type: yType,
    namespace: "y",
    // overriding domain min & max
    domainMin: yMin,
    domainMax: yMax,
    rangeMin: (config: PlotConfig) => config.plotBottom,
    rangeMax: (config: PlotConfig) => config.plotTop,
  });
</script>

<svg {width} {height}>
  <slot plotConfig={$plotConfig} />
</svg>
