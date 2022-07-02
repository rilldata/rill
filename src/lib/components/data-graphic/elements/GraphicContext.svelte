<!--
@component
A functional component that cascades its props to its children.
If a GraphicContext is a child of another GraphicContext, it will inherit its props and
honor any new props passed into it, thereby reconciling the props between it and its parent
for any of its children.
-->
<script lang="ts">
  import { guidGenerator } from "$lib/util/guid";

  import { contexts } from "../constants";
  import {
    cascadingContextStore,
    initializeMaxMinStores,
    initializeScale,
  } from "../state";
  import type {
    SimpleConfigurationStore,
    SimpleDataGraphicConfiguration,
    SimpleDataGraphicConfigurationArguments,
  } from "../state/types";

  export let width: number = undefined;
  export let height: number = undefined;
  export let top: number = undefined;
  export let bottom: number = undefined;
  export let left: number = undefined;
  export let right: number = undefined;

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

  export let shareXScale = true;
  export let shareYScale = true;

  const id = guidGenerator();

  const config = cascadingContextStore<
    SimpleDataGraphicConfigurationArguments,
    SimpleDataGraphicConfiguration
  >(
    contexts.config,
    {
      width,
      height,
      top,
      bottom,
      left,
      right,
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
      id,
    },
    /** these values are derived from the existing SimpleDataGraphicConfigurationArguments */
    {
      plotLeft: (config: SimpleDataGraphicConfiguration) => config.left,
      plotRight: (config: SimpleDataGraphicConfiguration) =>
        config.width - config.right,
      plotTop: (config: SimpleDataGraphicConfiguration) => config.top,
      plotBottom: (config: SimpleDataGraphicConfiguration) =>
        config.height - config.bottom,
      bodyLeft: (config: SimpleDataGraphicConfiguration) =>
        config.left + config.bodyBuffer || 0,
      bodyRight: (config: SimpleDataGraphicConfiguration) =>
        config.width - config.right - config.bodyBuffer || 0,
      bodyTop: (config: SimpleDataGraphicConfiguration) =>
        config.top + config.bodyBuffer || 0,
      bodyBottom: (config: SimpleDataGraphicConfiguration) =>
        config.height - config.bottom - config.bodyBuffer || 0,
      graphicWidth: (config: SimpleDataGraphicConfiguration) =>
        config.width -
        config.left -
        config.right -
        2 * (config.bodyBuffer || 0),
      graphicHeight: (config: SimpleDataGraphicConfiguration) =>
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
    id,
  });

  /** Reset any extremum values if we aren't sharing the scale or there is no parent cascade. */
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
    scaleType: xType || $config.yType,
    rangeMin: (config: SimpleDataGraphicConfiguration) => config.bodyLeft,
    rangeMax: (config: SimpleDataGraphicConfiguration) => config.bodyRight,
  });
  const yScale = initializeScale({
    namespace: "y",
    scaleType: yType || $config.yType,
    rangeMin: (config: SimpleDataGraphicConfiguration) => config.bodyBottom,
    rangeMax: (config: SimpleDataGraphicConfiguration) => config.bodyTop,
  });
</script>

<slot config={$config} xScale={$xScale} yScale={$yScale} />
