import type { RillReduxState } from "$lib/redux-store/store-root";

export const selectApplicationActiveEntity = (state: RillReduxState) =>
  state.application.activeEntity;
