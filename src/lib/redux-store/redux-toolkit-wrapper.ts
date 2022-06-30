import * as reduxToolkit from "@reduxjs/toolkit";

export const {
  createEntityAdapter,
  createSlice,
  createAsyncThunk,
  configureStore,
} = ((reduxToolkit as any)?.default ?? reduxToolkit) as typeof reduxToolkit;
