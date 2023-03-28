/**
 * @module extremum-resolution-store
 * This specialized store handles the resolution of plot bounds based
 * on multiple extrema. If multiple components set a maximum value within
 * the same namespace, the store will automatically pick the largest
 * for the store value. This enables us to determine, for instance, if
 * multiple lines are on the same chart, which ones determine the bounds.
 */
import { max, min } from "d3-array";
import { cubicOut } from "svelte/easing";
import { tweened } from "svelte/motion";
import { derived, get, writable, Writable } from "svelte/store";
import type { EasingFunction } from "svelte/transition";

const LINEAR_SCALE_STORE_DEFAULTS = {
  duration: 0,
  easing: cubicOut,
  direction: "min",
  namespace: undefined,
  alwaysOverrideInitialValue: false,
};

interface extremumArgs {
  duration?: number;
  easing?: EasingFunction;
  direction?: string;
  alwaysOverrideInitialValue?: boolean;
}

interface Extremum {
  value: number | Date;
  override?: boolean;
}

interface ExtremaStoreValue {
  [key: string]: Extremum;
}

const extremaFunctions = { min, max };

export function createExtremumResolutionStore(
  initialValue: number | Date = undefined,
  passedArgs: extremumArgs = {}
) {
  const args = { ...LINEAR_SCALE_STORE_DEFAULTS, ...passedArgs };
  const storedValues: Writable<ExtremaStoreValue> = writable({});
  let tweenProps = {
    duration: args.duration,
    easing: args.easing,
  };
  const valueTween = tweened(initialValue, tweenProps);
  function _update(key: string, value: number | Date, override = false) {
    // FIXME: there's an odd bug where if I don't check for equality first, I tend
    // to get an infinite loop with dates and the downstream scale.
    // This is easily fixed by only updating if the value has in fact changed.
    const extremum = get(storedValues)[key];
    if (extremum?.value === value && extremum?.override === override) return;
    storedValues.update((storeValue) => {
      if (!(key in storeValue))
        storeValue[key] = { value: undefined, override: false };
      storeValue[key].value = value;
      storeValue[key].override = override;
      return storeValue;
    });
  }
  /** add the initial value as its own key, if set by user. */
  if (initialValue && args.alwaysOverrideInitialValue === false) {
    _update("__initial_value__", initialValue);
  }

  function _remove(key: string) {
    storedValues.update((storeValue) => {
      delete storeValue[key];
      return storeValue;
    });
  }

  const domainExtremum = derived(
    storedValues,
    ($storedValues) => {
      let extremum;
      const extrema: Extremum[] = [...Object.values($storedValues)];
      for (const entry of extrema) {
        if (entry.override) {
          extremum = entry.value;
          break;
        } else {
          extremum = extremaFunctions[args.direction]([entry.value, extremum]);
        }
      }
      return extremum;
    },
    initialValue
  );

  // set the final tween with the value.
  domainExtremum.subscribe((value) => {
    if (value !== undefined) {
      valueTween.set(value, tweenProps);
    }
  });

  const returnedStore = {
    subscribe: valueTween.subscribe,
    setWithKey(key, value = undefined, override = undefined) {
      _update(key, value, override);
    },
    removeKey(key: string) {
      _remove(key);
    },
    setTweenProps(tweenPropsArgs) {
      tweenProps = tweenPropsArgs;
    },
  };
  return returnedStore;
}
