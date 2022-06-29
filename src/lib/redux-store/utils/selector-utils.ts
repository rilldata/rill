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

// @aditya, note that the generated selectors have a slightly different signature
// than in yuor generators: in these, the redux state is always the first argument.
// this made writing the svelte readable wrapper a bit  easier.
// I added this version alongside yours for now because I wasn't sure whether you needed the curried
// function style version for specific some application.
export function generateEntitySelectors<T>(sliceKey: keyof RillReduxState) {
  return {
    manySelector: (state: RillReduxState) =>
      state[sliceKey].ids.map((id) => <T>state[sliceKey].entities[id]),
    singleSelector: (state: RillReduxState, id: string) =>
      <T>state[sliceKey].entities[id],
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
