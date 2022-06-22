import type { ActionCreatorWithPreparedPayload } from "@reduxjs/toolkit";
import type { MetricsDefSliceActionTypes } from "$lib/redux-store/metrics-definition-slice";
import { MetricsDefSliceActions } from "$lib/redux-store/metrics-definition-slice";
import type { MetricsLeaderboardSliceTypes } from "$lib/redux-store/metrics-leaderboard-slice";
import { MetricsLeaderboardSliceActions } from "$lib/redux-store/metrics-leaderboard-slice";

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
  MetricsDefSliceActionTypes & MetricsLeaderboardSliceTypes
>;

export const ReduxActions = {
  ...MetricsDefSliceActions,
  ...MetricsLeaderboardSliceActions,
};
