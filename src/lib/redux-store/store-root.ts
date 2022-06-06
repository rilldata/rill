import { configureStore } from "@reduxjs/toolkit";
import metricsDefinitionReducer from "./metrics-definition/metrics-definition-slice";

export const store = configureStore({
  reducer: {
    metricsDefinition: metricsDefinitionReducer,
  },
});
