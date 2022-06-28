import { guidGenerator } from "$lib/util/guid";
import { scaleLinear, ScaleLinear, scaleTime } from "d3-scale";
import { getContext, onMount, setContext } from "svelte";
import { derived, writable, get } from "svelte/store";
import type { Writable } from "svelte/store"
import { createExtremumResolutionStore } from "../extremum-resolution-store";
import type { PlotConfig } from "../utils";
import { contexts } from "../contexts"

const CARTESIAN_SCALES_DEFAULTS = {
  xMinValue: undefined,
  xMaxValue: undefined,
  yMinValue: undefined,
  yMaxValue: undefined,

}

const SCALES = {
  number: scaleLinear,
  date: scaleTime
}

type ScaleRangeArgument = number | Date | ((arg0: PlotConfig) => (number | Date))

enum ScaleType {
  number = 'scaleLinear',
  date = 'scaleTime'
}

interface ScaleInitializerArguments {
  type: ScaleType,
  namespace: string,
  domainMin?: number | Date,
  domainMax?: number | Date,
  rangeMin: ScaleRangeArgument,
  rangeMax: ScaleRangeArgument,
}

export function initializeScaleContext(args: ScaleInitializerArguments) {
  const plotConfig = getContext(contexts.config);

  const scaleContextNamespace = contexts.scale(args.namespace);
  const minContextNamespace = contexts.min(args.namespace);
  const maxContextNamespace = contexts.max(args.namespace);

  const min = createExtremumResolutionStore(args.domainMin, {
    duration: 0,
    direction: "min",
  });
  const max = createExtremumResolutionStore(args.domainMax, {
    duration: 0,
    direction: "max"
  });

  /** Set the domain min and / or max if it is specified.
   * This will ensure the scale stays pegged at this value.
   */
  if (args.domainMin !== undefined) min.setWithKey('global', args.domainMin, true);
  if (args.domainMax !== undefined) max.setWithKey('global', args.domainMax, true);

  const scaleType = SCALES[args.type];

  /** the scale itself is a derived store, taking in the extrema and the plot config.
   */
  const scale = derived([min, max, plotConfig], ([$min, $max, $plotConfig]) => {
    let minRangeValue: (number | Date);
    let maxRangeValue: (number | Date);
    if (typeof args.rangeMin === 'function') {
      minRangeValue = args.rangeMin($plotConfig);
    } else {
      minRangeValue = args.rangeMin;
    }

    if (typeof args.rangeMax === 'function') {
      maxRangeValue = args.rangeMax($plotConfig);
    } else {
      maxRangeValue = args.rangeMax;
    }

    return scaleType()
      .domain([$min, $max])
      .range([minRangeValue, maxRangeValue]);
  })
  setContext(minContextNamespace, min);
  setContext(maxContextNamespace, max);
  setContext(scaleContextNamespace, scale);
  return { min, max, scale };
}

export const PLOT_CONFIG = {
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
  plotBottom: 120 - 12 - 4,
  xType: 'number',
  yType: 'number'
}


export function initializePlotConfigs(propArgs = {}) {
  const props = { ...PLOT_CONFIG, ...propArgs };
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
    xType: props.xType,
    yType: props.yType
  });

  function setValues(props) {

    function updateStore(key: string, fcn = undefined) {
      plotConfig.update((config: PlotConfig) => {
        if (props[key] !== config[key] && props[key] !== undefined) config[key] = fcn ? (fcn(props)) : props[key];
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
    updateStore('xType');
    updateStore('yType');
  }

  setValues(props);

  // reactively update any new fields in the store.

  setContext(contexts.config, plotConfig);
  return {
    subscribe: plotConfig.subscribe,
    set: setValues,
  };
}