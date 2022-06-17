import type { ActionCreatorWithPreparedPayload } from "@reduxjs/toolkit";
import type { MetricsDefSliceActionTypes } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import { MetricsDefSliceActions } from "$lib/redux-store/metrics-definition/metrics-definition-slice";

type ReduxActionArgs<Action> = Action extends ActionCreatorWithPreparedPayload<
  infer Args,
  unknown
>
  ? Args
  : never;
type ReduxSliceActionDefinitions<ReduxSliceActionTypes> = {
  [Action in keyof ReduxSliceActionTypes]: ReduxActionArgs<
    ReduxSliceActionTypes[Action]
  >;
};

export type ReduxActionDefinitions =
  ReduxSliceActionDefinitions<MetricsDefSliceActionTypes>;

export const ReduxActions = {
  ...MetricsDefSliceActions,
};
