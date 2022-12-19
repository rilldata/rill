<!--
@component
A functional component that cascades its props to its children.
If a GraphicContext is a child of another GraphicContext, it will inherit its props and
honor any new props passed into it, thereby reconciling the props between it and its parent
for any of its children.
-->
<script lang="ts">
  import { getContext, hasContext } from "svelte";
  import { guidGenerator } from "../../../util/guid";

  import { contexts } from "../constants";
  import {
    cascadingContextStore,
    initializeMaxMinStores,
    initializeScale,
    pruneProps,
  } from "../state";
  import type {
    ExtremumResolutionStore,
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

  export let xMinTweenProps = { duration: 0 };
  export let xMaxTweenProps = { duration: 0 };
  export let yMinTweenProps = { duration: 0 };
  export let yMaxTweenProps = { duration: 0 };

  export let shareXScale = true;
  export let shareYScale = true;

  const id = guidGenerator();

  const DEFAULTS = hasContext(contexts.config)
    ? {}
    : {
        width: 300,
        height: 200,
        top: 24,
        bottom: 24,
        left: 24,
        right: 24,
        fontSize: 12,
        textGap: 4,
        bodyBuffer: 4,
        marginBuffer: 4,
      };
  let parameters = {
    ...DEFAULTS,
    ...pruneProps({
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
    }),
  };

  $: parameters = {
    ...DEFAULTS,
    ...pruneProps({
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
    }),
  };

  const config = cascadingContextStore<
    SimpleDataGraphicConfigurationArguments,
    SimpleDataGraphicConfiguration
  >(
    contexts.config,
    parameters,
    /** these values are derived from the existing SimpleDataGraphicConfigurationArguments */
    {
      plotLeft: (config: SimpleDataGraphicConfiguration) => config.left,
      plotRight: (config: SimpleDataGraphicConfiguration) =>
        config.width - config.right,
      plotTop: (config: SimpleDataGraphicConfiguration) => config.top,
      plotBottom: (config: SimpleDataGraphicConfiguration) =>
        config.height - config.bottom,
      bodyLeft: (config: SimpleDataGraphicConfiguration) =>
        config.left + (config.bodyBuffer || 0),
      bodyRight: (config: SimpleDataGraphicConfiguration) =>
        config.width - config.right - (config.bodyBuffer || 0),
      bodyTop: (config: SimpleDataGraphicConfiguration) =>
        config.top + config.bodyBuffer || 0,
      bodyBottom: (config: SimpleDataGraphicConfiguration) =>
        config.height - config.bottom - (config.bodyBuffer || 0),
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

  $: config.reconcileProps(parameters);

  /** Reset any extremum values if we aren't sharing the scale or there is no parent cascade. */
  if (!config.hasParentCascade || !shareYScale) {
    initializeMaxMinStores({
      namespace: "y",
      domainMin: yMin,
      domainMax: yMax,
      domainMinTweenProps: yMinTweenProps,
      domainMaxTweenProps: yMaxTweenProps,
    });
  }

  if (!config.hasParentCascade || !shareXScale) {
    initializeMaxMinStores({
      namespace: "x",
      domainMin: xMin,
      domainMax: xMax,
      domainMinTweenProps: xMinTweenProps,
      domainMaxTweenProps: xMaxTweenProps,
    });
  }

  /** Now that the extrema are created, let's update the actual scales,
   * which are a store derived from the config & the extremum stores.
   */
  const xScale = initializeScale({
    namespace: "x",
    scaleType: xType || $config.xType,
    rangeMin: (config: SimpleDataGraphicConfiguration) => config.bodyLeft,
    rangeMax: (config: SimpleDataGraphicConfiguration) => config.bodyRight,
  });
  const yScale = initializeScale({
    namespace: "y",
    scaleType: yType || $config.yType,
    rangeMin: (config: SimpleDataGraphicConfiguration) => config.bodyBottom,
    rangeMax: (config: SimpleDataGraphicConfiguration) => config.bodyTop,
  });

  // update the xMin, xMax, yMin, yMax as needed
  const xMinStore = getContext(contexts.min("x")) as ExtremumResolutionStore;
  const xMaxStore = getContext(contexts.max("x")) as ExtremumResolutionStore;
  const yMinStore = getContext(contexts.min("y")) as ExtremumResolutionStore;
  const yMaxStore = getContext(contexts.max("y")) as ExtremumResolutionStore;

  $: if (yMaxTweenProps) {
    yMaxStore.setTweenProps(yMaxTweenProps);
  }

  $: if (xMin || $config?.xMin)
    xMinStore.setWithKey("global", xMin || $config.xMin, true);
  $: if (xMax || $config?.xMax)
    xMaxStore.setWithKey("global", xMax || $config.xMax, true);
  $: if (yMin || $config?.yMin)
    yMinStore.setWithKey("global", yMin || $config.yMin, true);
  $: if (yMax || $config?.yMax)
    yMaxStore.setWithKey("global", yMax || $config.yMax, true);
</script>

<slot config={$config} xScale={$xScale} yScale={$yScale} />
