import { get, writable } from "svelte/store";
import { setContext, getContext, hasContext } from "svelte";
import type { CascadingContextStore } from "./types";

function prune<T>(props: T) {
  return Object.keys(props).reduce((next, prop) => {
    if (props[prop] !== undefined) next[prop] = props[prop];
    return next;
  }, {})
}

function addDerivations(store, derivations) {
  store.update(state => {
    Object.keys(derivations).forEach(key => {
      state[key] = derivations[key](state);
    });
    return state;
  })
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
export function cascadingContextStore<T, V>(namespace: string, props: T, derivations = {}): CascadingContextStore<T, V> {
  // check to see if namespace exists.
  const hasParentCascade = hasContext(namespace);

  const prunedProps = prune<T>(props);

  let lastProps;
  const store = writable(prunedProps);
  let parentStore;
  if (hasParentCascade) {
    parentStore = getContext(namespace);
    store.set({
      ...get(parentStore), ...prunedProps
    })
    /** When the parent updates, we need to take care
    * to reconcile parent and child + any changed props.
    */
    parentStore.subscribe(state => {
      store.set({
        ...get(store),
        ...state,
        ...prune((lastProps || {}))
      });
      addDerivations(store, derivations)
      // add all derived values.
    })
  } else {
    // no-op.
  }
  // always reset the context here.
  setContext(namespace, store);

  return {
    hasParentCascade,
    subscribe: store.subscribe,
    reconcileProps(props: T) {
      lastProps = { ...props };

      /** let's update the store with the latest props. */
      store.set({ ...get(store), ...prune(lastProps) })
      addDerivations(store, derivations);
    }
  }
}