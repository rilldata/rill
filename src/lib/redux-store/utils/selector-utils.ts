import type { RillReduxState } from "$lib/redux-store/store-root";

export function generateBasicSelectors(sliceKey: keyof RillReduxState) {
  return {
    manySelector: (state: RillReduxState) =>
      state[sliceKey].ids.map((id) => state[sliceKey].entities[id]),
    singleSelector: (id: string) => {
      return (state: RillReduxState) => state[sliceKey].entities[id];
    },
  };
}

export function generateFilteredSelectors<FilterArgs extends Array<unknown>>(
  sliceKey: keyof RillReduxState,
  filter: (entity: unknown, ...args: FilterArgs) => boolean
) {
  return {
    manySelector: (...args: FilterArgs) => {
      return (state: RillReduxState) =>
        state[sliceKey].ids
          .filter((id) => filter(state[sliceKey].entities[id], ...args))
          .map((id) => state[sliceKey].entities[id]);
    },
    singleSelector: (id: string) => {
      return (state: RillReduxState) => state[sliceKey].entities[id];
    },
  };
}
