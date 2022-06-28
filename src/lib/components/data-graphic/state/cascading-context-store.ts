import { get, writable } from "svelte/store";
import { setContext, getContext, hasContext } from "svelte";

function prune(props) {
  let next = {};
  Object.keys(props).forEach(prop => {
    if (props[prop] !== undefined) next[prop] = props[prop];
  })
  return next;
}

function addDerivations(store, derivations) {
  store.update(state => {
    Object.keys(derivations).forEach(key => {
      state[key] = derivations[key](state);
    });
    return state;
  })
}

export function cascadingContextStore(namespace, props, derivations = {}) {
  // check to see if namespace exists.
  const hasParentCascade = hasContext(namespace);

  const prunedProps = prune(props);

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
    reconcileProps(props) {
      lastProps = { ...props };

      /** let's update the store with the latest props. */
      store.set({ ...get(store), ...prune(lastProps) })
      addDerivations(store, derivations);
    }
  }
}