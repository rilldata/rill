import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  setStatusPrepare,
  setStatusReducer,
} from "$lib/redux-store/utils/loading-utils";

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
    updateBigNumber: {
      reducer: bigNumberAdapter.upsertOne,
      prepare: (bigNumberEntity: BigNumberEntity) => ({
        payload: bigNumberEntity,
      }),
    },

    setBigNumberStatus: {
      reducer: setStatusReducer,
      prepare: setStatusPrepare,
    },
  },
});

export const { updateBigNumber, setBigNumberStatus } = bigNumberSlice.actions;

export const bigNumberReducer = bigNumberSlice.reducer;
