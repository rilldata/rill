import { get, writable } from "svelte/store";
import { setContext, getContext, hasContext } from "svelte";
import type { SimpleDataGraphicConfiguration, SimpleDataGraphicConfigurationArguments } from "./types";

export function pruneProps<T extends object>(props: T): T {
  return Object.keys(props).reduce((next, prop) => {
    if (props[prop] !== undefined) next[prop] = props[prop];
    return next;
  }, {}) as T;
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
