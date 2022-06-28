<script lang="ts">
  import type { PlotConfig } from "../utils";
  import { contexts } from "../contexts";
  import { cascadingContextStore } from "../state/cascading-context-store";
  import {
    initializeMaxMinStores,
    initializeScale,
  } from "../state/scale-stores";
  import { hasContext } from "svelte";

  export let width: number = undefined;
  export let height: number = undefined;
  export let top: number = undefined;
  export let bottom: number = undefined;
  export let left: number = undefined;
  export let right: number = undefined;
  export let buffer: number = undefined;

  export let fontSize: number = undefined;
  export let textGap: number = undefined;

  export let bodyBuffer: number = undefined;
  export let marginBuffer: number = undefined;

  export let xType = undefined;
  export let yType = undefined;

  export let xMin: number | Date = undefined;
  export let xMax: number | Date = undefined;
  export let yMin: number | Date = undefined;
  export let yMax: number | Date = undefined;

  export let shareXScale: boolean = true;
  export let shareYScale: boolean = true;

  const config = cascadingContextStore(
    contexts.config,
    {
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
      xMin,
      xMax,
      yMin,
      yMax,
      bodyBuffer,
      marginBuffer,
    },
    {
      plotLeft: (config) => config.left,
      plotRight: (config) => config.width - config.right,
      plotTop: (config) => config.top,
      plotBottom: (config) => config.height - config.bottom,
      bodyLeft: (config) => config.left + config.bodyBuffer || 0,
      bodyRight: (config) =>
        config.width - config.right - config.bodyBuffer || 0,
      bodyTop: (config) => config.top + config.bodyBuffer || 0,
      bodyBottom: (config) =>
        config.height - config.bottom - config.bodyBuffer || 0,
      graphicWidth: (config) =>
        config.width -
        config.left -
        config.right -
        2 * (config.bodyBuffer || 0),
      graphicHeight: (config) =>
        config.height -
        config.top -
        config.bottom -
        2 * (config.bodyBuffer || 0),
    }
  );

  $: config.reconcileProps({
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
    xMin,
    xMax,
    yMin,
    yMax,
    bodyBuffer,
    marginBuffer,
  });

  /** we will need to (1) reset the derived scale store and (2) (if we're not sharing the x scale) we need
   * FIRST to regenerate the extrema.
   */

  if (!config.hasParentCascade || !shareYScale) {
    initializeMaxMinStores({
      namespace: "y",
      domainMin: yMin,
      domainMax: yMax,
    });
  }

  if (!config.hasParentCascade || !shareXScale) {
    initializeMaxMinStores({
      namespace: "x",
      domainMin: xMin,
      domainMax: xMax,
    });
  }

  /** Now that the extrema are created, let's update the actual scales,
   * which are a store derived from the config & the extremum stores.
   */
  const xScale = initializeScale({
    namespace: "x",
    scaleType: "date",
    rangeMin: (config: PlotConfig) => config.bodyLeft,
    rangeMax: (config: PlotConfig) => config.bodyRight,
  });
  const yScale = initializeScale({
    namespace: "y",
    scaleType: "number",
    rangeMin: (config: PlotConfig) => config.bodyBottom,
    rangeMax: (config: PlotConfig) => config.bodyTop,
  });
</script>

<slot config={{ ...$config }} xScale={$xScale} yScale={$yScale} />
