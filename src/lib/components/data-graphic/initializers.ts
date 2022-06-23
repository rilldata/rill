import { guidGenerator } from "$lib/util/guid";
import { scaleLinear, ScaleLinear, scaleTime } from "d3-scale";
import { getContext, onMount, setContext } from "svelte";
import { derived, writable, Writable } from "svelte/store";
import { createExtremumResolutionStore } from "./extremum-resolution-store";
import type { PlotConfig } from "./utils";

const CARTESIAN_SCALES_DEFAULTS = {
  xMinValue: undefined,
  xMaxValue: undefined,
  yMinValue: undefined,
  yMaxValue: undefined,
  xType: 'number',
  yType: 'number'
}

const SCALES = {
  number: scaleLinear,
  date: scaleTime
}

export function initializeCartesianScales(passedArguments = {}) {
  const args = { ...CARTESIAN_SCALES_DEFAULTS, ...passedArguments }
  // const xScale: Writable<ScaleLinear<number, number>> = writable(undefined);
  // const yScale: Writable<ScaleLinear<number, number>> = writable(undefined);
  const xTypeScale = SCALES[args.xType] || scaleLinear;
  const yTypeScale = SCALES[args.yType] || scaleLinear;

  const xMin = createExtremumResolutionStore(args.xMinValue, {
    duration: 0,
    direction: "min",
  });
  const xMax = createExtremumResolutionStore(args.xMaxValue, {
    duration: 0,
    direction: "max"
  });

  const yMin = createExtremumResolutionStore(args.yMinValue, {
    duration: 0,
    direction: "min",
  });
  const yMax = createExtremumResolutionStore(args.yMaxValue, {
    duration: 0,
    direction: "max"
  });

  if (args.xMinValue !== undefined) xMin.setWithKey('global', args.xMinValue, true);
  if (args.xMaxValue !== undefined) xMax.setWithKey('global', args.xMaxValue, true);
  if (args.yMinValue !== undefined) yMin.setWithKey('global', args.yMinValue, true);
  if (args.yMaxValue !== undefined) yMax.setWithKey('global', args.yMaxValue, true);
  setContext("rill:data-graphic:x-min", xMin);
  setContext("rill:data-graphic:x-max", xMax);
  setContext("rill:data-graphic:y-min", yMin);
  setContext("rill:data-graphic:y-max", yMax);


  // get the plotConfig.
  const plotConfig = getContext('rill:data-graphic:plot-config');

  const xScale = derived([xMin, xMax, plotConfig], ([$xMin, $xMax, $plotConfig]) => {
    return xTypeScale()
      .domain([$xMin, $xMax])
      .range([$plotConfig.plotLeft, $plotConfig.plotRight])
  })

  const yScale = derived([yMin, yMax, plotConfig], ([$yMin, $yMax, $plotConfig]) => {
    return yTypeScale()
      .domain([$yMin, $yMax])
      .range([$plotConfig.plotBottom, $plotConfig.plotTop])
  })

  setContext("rill:data-graphic:x-scale", xScale);
  setContext("rill:data-graphic:y-scale", yScale);
  return { xScale, yScale, xMin, xMax, yMin, yMax };
}

const PLOT_CONFIG = {
  left: 12,
  right: 12,
  top: 12,
  bottom: 12,
  buffer: 4,
  width: 360,
  height: 120,
  plotLeft: 12 + 4,
  plotRight: 360 - 12 - 4,
  plotTop: 12 + 4,
  plotBottom: 120 - 12 - 4
}


export function initializePlotConfigs(propArgs = {}) {
  const props = { ...PLOT_CONFIG, propArgs };
  const id = guidGenerator();

  let devicePixelRatio = 1;

  onMount(() => {
    devicePixelRatio = window.devicePixelRatio;
  });

  const plotConfig: Writable<PlotConfig> = writable({
    id,
    devicePixelRatio,
    top: props.top,
    bottom: props.bottom,
    left: props.left,
    right: props.right,
    buffer: props.buffer,
    width: props.width,
    height: props.height,
    plotTop: props.top + props.buffer,
    plotBottom: props.height - props.buffer - props.bottom,
    plotLeft: props.left + props.buffer,
    plotRight: props.width - props.right - props.buffer,
    fontSize: props.fontSize,
    textGap: props.textGap,
  });

  function setValues(props) {

    function updateStore(key: string, fcn = undefined) {
      plotConfig.update((config: PlotConfig) => {
        if (props[key] !== config[key]) config[key] = fcn ? (fcn(props)) : props[key];
        return config;
      });
    }
    updateStore('width');
    updateStore('height');
    updateStore('top');
    updateStore('bottom');
    updateStore('left');
    updateStore('right');
    updateStore('buffer');
    updateStore('plotTop', () => props.top + props.buffer);
    updateStore('plotBottom', () => props.height - props.buffer - props.bottom);
    updateStore('plotLeft', () => props.left + props.buffer);
    updateStore('plotRight', () => props.width - props.right - props.buffer);
    updateStore('fontSize');
    updateStore('textGap');
  }

  setValues(props);

  // reactively update any new fields in the store.

  setContext("rill:data-graphic:plot-config", plotConfig);
  return {
    subscribe: plotConfig.subscribe,
    set: setValues
  };
}