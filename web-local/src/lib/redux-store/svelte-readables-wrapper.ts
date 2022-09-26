import { readable } from "svelte/store";
import type { RillReduxState, RillReduxStore } from "./store-root";

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
      let prevState: T;
      // redux `store.subscribe()` returns an un
      return store.subscribe(() => {
        const curState = selector(store.getState(), ...selectorArgs);
        if (prevState !== curState) {
          prevState = curState;
          set(prevState);
        }
      });
    });
};
