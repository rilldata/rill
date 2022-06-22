import * as reduxToolkit from "@reduxjs/toolkit";
import { createLogger } from "redux-logger";
import { metricsDefinitionReducer } from "./metrics-definition-slice";
import { metricsLeaderboardReducer } from "./metrics-leaderboard-slice";
import { readable } from "svelte/store";
import { browser } from "$app/env";

const { configureStore } = reduxToolkit;

export const store = configureStore({
  reducer: {
    metricsDefinition: metricsDefinitionReducer,
    metricsLeaderboard: metricsLeaderboardReducer,
  },
  middleware: (getDefaultMiddleware) => {
    // if (browser && process.env.NODE_ENV === `development`) {
    //   // calling `createLogger()` outside the browser causes SSR errors
    //   const logger = createLogger();
    //   return getDefaultMiddleware().concat(logger);
    // }
    return getDefaultMiddleware();
  },
});

export const reduxReadable = readable(store.getState(), (set) => {
  return store.subscribe(() => {
    set(store.getState());
  });
});

export const createReadableStoreWithSelector = (selector) =>
  readable(selector(store.getState()), (set) => {
    return store.subscribe(() => {
      set(selector(store.getState()));
    });
  });
