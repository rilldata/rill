import type { EntityStatus } from "../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityState, PayloadAction } from "@reduxjs/toolkit";

/**
 * Prepare method for a setStatus reducer.
 * Creating the reducer object fails type definition.
 * TODO: figure out the style issue and create the full object.
 */
export const setStatusPrepare = (id: string, status: EntityStatus) => ({
  payload: { id, status },
});
/**
 * Reducer method to modify state for a setStatus reducer.
 */
export const setStatusReducer = <T extends { status: EntityStatus }>(
  state: EntityState<T>,
  {
    payload: { id, status },
  }: PayloadAction<{ id: string; status: EntityStatus }>
) => {
  if (!state.entities[id]) return;
  state.entities[id].status = status;
};
