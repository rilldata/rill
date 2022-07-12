import type {
  RillReduxEntities,
  RillReduxState,
} from "$lib/redux-store/store-root";

/**
 * Generates
 * 1. Single entity selector by id
 * 2. Selector for multiple entities by list of ids.
 */
function generateCommonSelectors<Entity extends RillReduxEntities>(
  sliceKey: keyof RillReduxState
) {
  const singleSelector = (state: RillReduxState, id: string) =>
    <Entity>state[sliceKey].entities[id];
  return {
    manySelectorByIds: (state: RillReduxState, ids: Array<string>) => {
      console.log(ids);
      console.log(state);
      console.log(ids.map((id) => singleSelector(state, id)));
      return ids.map((id) => singleSelector(state, id));
    },
    singleSelector,
  };
}

/**
 * Generates selectors from {@link generateCommonSelectors}
 * Also generates a selector for all entities.
 */
export function generateEntitySelectors<Entity extends RillReduxEntities>(
  sliceKey: keyof RillReduxState
) {
  return {
    manySelector: (state: RillReduxState) =>
      state[sliceKey].ids.map((id) => <Entity>state[sliceKey].entities[id]),
    ...generateCommonSelectors<Entity>(sliceKey),
  };
}

/**
 * Generates selectors from {@link generateCommonSelectors}
 * Also generates a selector for multiple entities by a filter criteria supplied by 'filter' param.
 */
export function generateFilteredEntitySelectors<
  FilterArgs extends Array<unknown>,
  Entity extends RillReduxEntities
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
