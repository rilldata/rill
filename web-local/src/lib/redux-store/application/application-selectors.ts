import type { RillReduxState } from "../store-root";

export const selectApplicationActiveEntity = (state: RillReduxState) =>
  state.application.activeEntity;
