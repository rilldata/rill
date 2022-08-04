import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  setStatusPrepare,
  setStatusReducer,
} from "$lib/redux-store/utils/loading-utils";
import {
  setFieldPrepare,
  setFieldReducer,
} from "$lib/redux-store/utils/slice-utils";
import type { PayloadAction } from "@reduxjs/toolkit";

export interface BigNumberEntity {
  id: string;
  bigNumbers: Record<string, number>;
  referenceValues?: Record<string, number>;
  status: EntityStatus;
}

const bigNumberAdapter = createEntityAdapter<BigNumberEntity>();

const bigNumberSlice = createSlice({
  name: "bigNumber",
  initialState: bigNumberAdapter.getInitialState(),
  reducers: {
    setBigNumber: {
      reducer: setFieldReducer("bigNumbers"),
      prepare: setFieldPrepare<BigNumberEntity, "bigNumbers">("bigNumbers"),
    },
    setReferenceValues: {
      reducer: (
        state,
        {
          payload: { id, referenceValues },
        }: PayloadAction<{
          id: string;
          referenceValues: Record<string, number>;
        }>
      ) => {
        if (!state.entities[id]) {
          bigNumberAdapter.addOne(state, {
            id,
            bigNumbers: {},
            referenceValues,
            status: EntityStatus.Idle,
          });
        } else {
          state.entities[id].referenceValues = referenceValues;
        }
      },
      prepare: setFieldPrepare<BigNumberEntity, "referenceValues">(
        "referenceValues"
      ),
    },

    setBigNumberStatus: {
      reducer: setStatusReducer,
      prepare: setStatusPrepare,
    },
  },
});

export const { setBigNumber, setReferenceValues, setBigNumberStatus } =
  bigNumberSlice.actions;

export const bigNumberReducer = bigNumberSlice.reducer;
