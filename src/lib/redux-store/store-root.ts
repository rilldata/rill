import { configureStore } from "@reduxjs/toolkit";
// import type { EnhancedStore } from "@reduxjs/toolkit";
import logger from "redux-logger";
import metricsDefinitionReducer from "./metrics-definition/metrics-definition-slice";

import { readable } from "svelte/store";
// import metricsDefinitionsMetadataReducer from "./metrics-definition/metrics-definitions-metatdata-slice";

// const middlewares = [];

// if (process.env.NODE_ENV === `development`) {
//   middlewares.push(logger);
// }

export const store = configureStore({
  reducer: {
    metricsDefinition: metricsDefinitionReducer,
    // metricsDefinitionsMetadata: metricsDefinitionsMetadataReducer,
  },
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(logger),
});

export const reduxReadable = readable(store.getState(), (set) => {
  const unsubscribe = store.subscribe(() => {
    set(store.getState());
  });
  return unsubscribe;
});

export const createReadableStoreWithSelector = (selector) =>
  readable(selector(store.getState()), (set) => {
    const unsubscribe = store.subscribe(() => {
      set(selector(store.getState()));
    });
    return unsubscribe;
  });
