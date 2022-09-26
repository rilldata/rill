import { createSlice } from "../redux-toolkit-wrapper";
import type { ActiveEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type { PayloadAction } from "@reduxjs/toolkit";

export interface ApplicationReduxState {
  activeEntity?: ActiveEntity;
}

/**
 * Keeps application store from the older direct state to redux store.
 */
export const applicationSlice = createSlice({
  name: "application",
  initialState: {} as ApplicationReduxState,
  reducers: {
    setApplicationActiveState: {
      reducer: (
        state,
        { payload: activeEntity }: PayloadAction<ActiveEntity>
      ) => {
        state.activeEntity = activeEntity;
      },
      prepare: (activeEntity: ActiveEntity) => ({ payload: activeEntity }),
    },
  },
});

export const { setApplicationActiveState } = applicationSlice.actions;
export const applicationReducer = applicationSlice.reducer;
