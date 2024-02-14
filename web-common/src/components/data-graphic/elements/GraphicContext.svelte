<!--
@component
A functional component that cascades its props to its children.
If a GraphicContext is a child of another GraphicContext, it will inherit its props and
honor any new props passed into it, thereby reconciling the props between it and its parent
for any of its children.
-->
<script lang="ts">
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { getContext, hasContext, onMount } from "svelte";
  import { contexts } from "../constants";
  import {
    ScaleType,
    cascadingContextStore,
    initializeMaxMinStores,
    initializeScale,
    pruneProps,
  } from "../state";
  import type {
    ExtremumResolutionStore,
    SimpleDataGraphicConfiguration,
  } from "@rilldata/web-common/components/data-graphic/state/types";

  export let width: number | undefined = undefined;
  export let height: number | undefined = undefined;
  export let top: number | undefined = undefined;
  export let bottom: number | undefined = undefined;
  export let left: number | undefined = undefined;
  export let right: number | undefined = undefined;

  export let fontSize: number | undefined = undefined;
  export let textGap: number | undefined = undefined;

  export let bodyBuffer: number | undefined = undefined;
  export let marginBuffer: number | undefined = undefined;

  export let xType: ScaleType = ScaleType.DATE;
  export let yType: ScaleType = ScaleType.NUMBER;

  export let xMin: number | undefined | Date = undefined;
  export let xMax: number | undefined | Date = undefined;
  export let yMin: number | undefined | Date = undefined;
  export let yMax: number | undefined | Date = undefined;

  export let xMinTweenProps = { duration: 0 };
  export let xMaxTweenProps = { duration: 0 };
  export let yMinTweenProps = { duration: 0 };
  export let yMaxTweenProps = { duration: 0 };

  export let shareXScale = true;
  export let shareYScale = true;

  const id = guidGenerator();

  let devicePixelRatio = 1;
  onMount(() => {
    devicePixelRatio = window.devicePixelRatio;
  });

  const props = {
    width,
    height,
    top,
    bottom,
    left,
    right,
    fontSize,
    textGap,
    devicePixelRatio,
    xType,
    yType,
    xMin,
    xMax,
    yMin,
    yMax,
    bodyBuffer,
    marginBuffer,
    id,
  };

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
        devicePixelRatio,
      };

  $: parameters = {
    ...DEFAULTS,
    ...pruneProps(props),
  };

  const config = cascadingContextStore(contexts.config, parameters);

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

  $: if (yMaxTweenProps && yMaxStore) {
    yMaxStore.setTweenProps(yMaxTweenProps);
  }

  $: if (xMin !== undefined || $config?.xMin)
    xMinStore.setWithKey("global", xMin || $config.xMin, true);
  $: if (xMax !== undefined || $config?.xMax)
    xMaxStore.setWithKey("global", xMax || $config.xMax, true);
  $: if (yMin !== undefined || $config?.yMin)
    yMinStore.setWithKey("global", yMin || $config.yMin, true);
  $: if (yMax !== undefined || $config?.yMax)
    yMaxStore.setWithKey("global", yMax || $config.yMax, true);
</script>

<slot config={$config} xScale={$xScale} yScale={$yScale} />
