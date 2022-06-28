import { ScaleLinear, scaleLinear, ScaleTime, scaleTime } from "d3-scale";
import { getContext, setContext, getAllContexts } from "svelte";
import { derived, Readable, Writable, get } from "svelte/store";
import { contexts } from "../contexts";
import { createExtremumResolutionStore } from "../extremum-resolution-store";
import type { PlotConfig } from "../utils";

const SCALES = {
  number: scaleLinear,
  date: scaleTime
}

/** We operate on the domain through these stores. */
export function initializeMaxMinStores({
  namespace,
  domainMin = undefined, domainMax = undefined
}) {
  // initialize
  const minStore = createExtremumResolutionStore(domainMin, { direction: 'min', duration: 0 });
  const maxStore = createExtremumResolutionStore(domainMax, { direction: 'max', duration: 0 });
  if (domainMin !== undefined) minStore.setWithKey('global', domainMin, true);
  if (domainMax !== undefined) maxStore.setWithKey('global', domainMax, true);
  // set the contexts.
  setContext(contexts.min(namespace), minStore);
  setContext(contexts.max(namespace), maxStore);
  return { minStore, maxStore };
}

export function initializeScale(args) {
  const minStore = getContext(contexts.min(args.namespace)) as Writable<(number | Date)>;
  const maxStore = getContext(contexts.max(args.namespace)) as Writable<(number | Date)>;
  const config = getContext(contexts.config) as PlotConfig;
  const scaleStore = derived([minStore, maxStore, config], ([$min, $max, $config]) => {
    const scale = SCALES[args.scaleType];
    const minRangeValue: (number | Date) = typeof args.rangeMin === 'function' ? args.rangeMin($config) : args.rangeMin;
    const maxRangeValue: (number | Date) = typeof args.rangeMax === 'function' ? args.rangeMax($config) : args.rangeMax;
    return scale()
      .domain([$min, $max]).range([minRangeValue, maxRangeValue])
  });
  scaleStore.type = args.scaleType;
  setContext(contexts.scale(args.namespace), scaleStore);
  return scaleStore;
}