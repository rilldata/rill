import type { RillReduxState } from "$lib/redux-store/store-root";

function generateCommonSelectors<Entity>(sliceKey: keyof RillReduxState) {
  const singleSelector = (state: RillReduxState, id: string) =>
    <Entity>state[sliceKey].entities[id];
  return {
    manySelectorByIds: (state: RillReduxState, ids: Array<string>) =>
      ids.map((id) => singleSelector(state, id)),
    singleSelector,
  };
}

export function generateEntitySelectors<Entity>(
  sliceKey: keyof RillReduxState
) {
  return {
    manySelector: (state: RillReduxState) =>
      state[sliceKey].ids.map((id) => <Entity>state[sliceKey].entities[id]),
    ...generateCommonSelectors<Entity>(sliceKey),
  };
}

export function generateFilteredEntitySelectors<
  FilterArgs extends Array<unknown>,
  Entity
>(
  sliceKey: keyof RillReduxState,
  filter: (entity: unknown, ...args: FilterArgs) => boolean
) {
  return {
    manySelector: (state: RillReduxState, ...args: FilterArgs) =>
      state[sliceKey].ids
        .filter((id) => filter(state[sliceKey].entities[id], ...args))
        .map((id) => <Entity>state[sliceKey].entities[id]),
    ...generateCommonSelectors<Entity>(sliceKey),
  };
}
