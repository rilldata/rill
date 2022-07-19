import type { MetricsExplorerSliceTypes } from "$lib/redux-store/explore/explore-slice";
import { MetricsExplorerSliceActions } from "$lib/redux-store/explore/explore-slice";
import type { MetricsDefSliceActionTypes } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import { MetricsDefSliceActions } from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import type { ActionCreatorWithPreparedPayload } from "@reduxjs/toolkit";

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

export type ReduxActionDefinitions = ReduxSliceActionDefinitions<
  MetricsDefSliceActionTypes & MetricsExplorerSliceTypes
>;

export const ReduxActions = {
  ...MetricsDefSliceActions,
  ...MetricsExplorerSliceActions,
};
