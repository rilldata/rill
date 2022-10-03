import * as reduxToolkit from "@reduxjs/toolkit";

// we need this to get redux toolkit working in ESM environment.
export const {
  createEntityAdapter,
  createSlice,
  createAsyncThunk,
  configureStore,
} = ((reduxToolkit as any)?.default ?? reduxToolkit) as typeof reduxToolkit;
