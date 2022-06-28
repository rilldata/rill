import { readable } from "svelte/store";
import type {
  RillReduxState,
  RillReduxStore,
} from "$lib/redux-store/store-root";

type OptionalRestArgs = undefined | unknown[];

type SelectorWithOptionalArgs<T, U extends OptionalRestArgs> = (
  state: RillReduxState,
  ...args: U
) => T;

export const createReadableFactoryWithSelector = <
  T,
  U extends OptionalRestArgs
>(
  store: RillReduxStore,
  selector: SelectorWithOptionalArgs<T, U>
) => {
  return (...selectorArgs: U) =>
    readable(selector(store.getState(), ...selectorArgs), (set) => {
      // redux `store.subscribe()` returns an un
      const unsubscribe = store.subscribe(() => {
        set(selector(store.getState(), ...selectorArgs));
      });
      return unsubscribe;
    });
};
