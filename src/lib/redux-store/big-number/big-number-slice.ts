import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";

export interface BigNumberEntity {
  id: string;
  bigNumbers: Record<string, number>;
  referenceValues?: Record<string, number>;
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
  },
});

export const { updateBigNumber } = bigNumberSlice.actions;

export const bigNumberReducer = bigNumberSlice.reducer;
