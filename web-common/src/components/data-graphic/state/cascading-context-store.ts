import { get, writable } from "svelte/store";
import { setContext, getContext, hasContext } from "svelte";
import type { SimpleDataGraphicConfiguration, SimpleDataGraphicConfigurationArguments } from "./types";
import { ScaleType } from ".";

export function pruneProps<T extends object>(props: T): T {
  return Object.keys(props).reduce((next, prop) => {
    if (props[prop] !== undefined) next[prop] = props[prop];
    return next;
  }, {}) as T;
}

 export const SIMPLE_DATA_GRAPHIC_DEFAULTS: SimpleDataGraphicConfigurationArguments ={
// These defaults are legimate defaults
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
  devicePixelRatio: 1,

  /**
   * all the values below do not represent "real"
   * defaults, but are included here to make sure
   * that the type of SIMPLE_DATA_GRAPHIC_DEFAULTS
   * aligns with the contract implied at the single call
   * as of 2024-02, any downstream values at least recieve
   * _some_ input.
   * 
   * if any of the default values below is actually used,
   * we will log a warning, because in that case something
   * has gone wrong!
   */

  // these values should always be overidded by props to
  // <GraphicsContext> (there are default props if none
  // are explicity passed)
  xType: ScaleType.DATE,
  yType: ScaleType.NUMBER,


  // this value should always be overridden by 
  // `const id = guidGenerator()` in `GraphicsContext.sveltez
  id: "DUMMY_ID",

  // Based on current types, these values may be undefined!
  // That should not be the case based on the, expected contract,
  // so it may be that more careful initialization is required,
  // or that SimpleDataGraphicConfigurationArguments should be
  // amended to allow these to be null.
  xMin: 0,
  xMax: 1,
  yMin: 0,
  yMax: 1,
}


/**
 * This function tries to clean up the typing of the creation
 * of paramaters for the cascadingContextStore. As of 2024-02,
 * `cascadingContextStore` was only called in one place, and
 * it's expected `parameters` argument at that call site was
 * supposed to have type SimpleDataGraphicConfigurationArguments.
 * 
 * `parameters` was created by (a) optionally using a set
 * of defaults, and (b) overriding those defaults with any 
 * available props.
 * 
 * We implement that that same strategy here, but written out
 * in detail to allow more careful typing, and with warnings
 * in case of a missing field, since Dhiraj noted that he
 * recalls that happening in some cases
 * 
 */
export function makeContextStoreProps(props:Partial<SimpleDataGraphicConfigurationArguments>,useDefault:boolean):SimpleDataGraphicConfigurationArguments {
  // init with defaults
  const finalProps = {...SIMPLE_DATA_GRAPHIC_DEFAULTS};

  // overwrite with input props
  for (const key in finalProps) {
    if (props[key] !== undefined) {
      // overwrite the default with the input props, if available
      finalProps[key] = props[key];
    } else if (!useDefault) {
      // if the there is no input prop for this key, 
      // but we are not supposed to be using defaults,
      // log a warning
      console.warn(`makeContextStoreProps: no input prop for key ${key} and useDefault is false`)
    } else{
      // if we _are_ using defaults, and there is no input prop
      // for this key but this is one of the keys that should
      // never use a default, log a warning
      console.warn(`makeContextStoreProps: used default for key "${key}", which should never take a defualt value`)
    }
  }
  return finalProps;
}

function addDerivations(store, derivations) {
  store.update((state) => {
    Object.keys(derivations).forEach((key) => {
      state[key] = derivations[key](state);
    });
    return state;
  });
}

/** Creates a store that passes itself down as a context.
 * If any children of the parent that created the store create a cascadingContextStore,
 * the store value will look like {...parentProps, ...childProps}.
 * In this case, the child component calling the new cascadingContextStore will pass the
 * new store down to its children, reconciling any differences downstream.
 *
 * this may seem complicated, but it does enable a lot of important
 * reactive data viz component compositions.
 * Most consumers of the data graphic components won't need to worry about this store.
 */
export function cascadingContextStore(
  namespace: string,
  props: SimpleDataGraphicConfigurationArguments,
) {
  const derivations = {
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
  };

  // check to see if namespace exists.
  const hasParentCascade = hasContext(namespace);

  const prunedProps = pruneProps<SimpleDataGraphicConfigurationArguments>(props);

  let lastProps = props;
  let lastParentState = {};

  const store = writable<SimpleDataGraphicConfigurationArguments | SimpleDataGraphicConfiguration>(prunedProps);
  let parentStore;

  if (hasParentCascade) {
    parentStore = getContext(namespace);
    store.set({
      ...get(parentStore),
      ...prunedProps,
    });

    /** When the parent updates, we need to take care
     * to reconcile parent and child + any changed props.
     */
    parentStore.subscribe((parentState) => {
      lastParentState = { ...parentState };
      store.set({
        ...parentState, // the parent state
        ...pruneProps(lastProps), // last props to be reconciled overrides clashing keys with current state
      });
      // add the derived values into the final store.
      addDerivations(store, derivations);
    });
  }
  addDerivations(store, derivations);
  // always reset the context for all children.
  setContext(namespace, store);
  return {
    hasParentCascade,
    subscribe: store.subscribe,
    reconcileProps(props: SimpleDataGraphicConfigurationArguments) {
      lastProps = { ...props };

      /** let's update the store with the latest props. */
      store.set({ ...lastParentState, ...pruneProps(lastProps) });
      addDerivations(store, derivations);
    },
  };
}
