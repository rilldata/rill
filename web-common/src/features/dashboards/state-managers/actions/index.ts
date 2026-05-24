import { filterActions } from "@rilldata/web-common/features/dashboards/state-managers/actions/filters";
import { measureFilterActions } from "@rilldata/web-common/features/dashboards/state-managers/actions/measure-filters";
import type { ExploreState } from "web-common/src/features/dashboards/stores/explore-state";
import { comparisonActions } from "./comparison";
import { contextColActions } from "./context-columns";
import { dimensionFilterActions } from "./dimension-filters";
import { dimensionTableActions } from "./dimension-table";
import { dimensionActions } from "./dimensions";
import { measureActions } from "./measures";
import { sortActions } from "./sorting";
import type {
  DashboardCallbackExecutor,
  DashboardMutatorFn,
  DashboardMutatorFns,
  DashboardUpdaters,
} from "./types";
import { leaderboardActions } from "./leaderboard";

export type StateManagerActions = ReturnType<typeof createStateManagerActions>;

/**
 * DashboardConnectedMutators object contains functions that
 * are closed over references connected to the live dashboard,
 * so calling these functions will actually update the dashboard.
 */
type DashboardConnectedMutators = {
  /**
   * Used to update the dashboard.
   */
  updateDashboard: DashboardCallbackExecutor;
};

export const createStateManagerActions = (
  actionArgs: DashboardConnectedMutators,
) => {
  return {
    /**
     * Actions related to the sorting state of the dashboard.
     */
    sorting: createDashboardUpdaters(actionArgs, sortActions),

    /**
     * Actions related to the dashboard comparison state.
     */
    comparison: createDashboardUpdaters(actionArgs, comparisonActions),

    /**
     * Actions related to the dashboard context columns.
     */
    contextColumn: createDashboardUpdaters(actionArgs, contextColActions),

    /**
     * Actions related to dimensions.
     */
    dimensions: createDashboardUpdaters(actionArgs, dimensionActions),

    /**
     * Actions related to measures.
     */
    measures: createDashboardUpdaters(actionArgs, measureActions),

    /**
     * Common filter actions
     */
    filters: createDashboardUpdaters(actionArgs, filterActions),

    /**
     * Actions related to dimensions filters
     */
    dimensionsFilter: createDashboardUpdaters(
      actionArgs,
      dimensionFilterActions,
    ),

    /**
     * Actions related to the dimension table.
     */
    dimensionTable: createDashboardUpdaters(actionArgs, dimensionTableActions),

    /**
     * Actions related to measure filters
     */
    measuresFilter: createDashboardUpdaters(actionArgs, measureFilterActions),

    /**
     * Actions related to the leaderboard.
     */
    leaderboard: createDashboardUpdaters(actionArgs, leaderboardActions),
  };
};

/**
 * `dashboardMutatorToUpdater` takes a DashboardConnectedMutators
 * object, and returns a DashboardMutatorFn that directly updates
 * the dashboard by calling the included DashboardCallbackExecutor.
 **/
function dashboardMutatorToUpdater<T extends unknown[]>(
  connectedMutators: DashboardConnectedMutators,
  mutator: DashboardMutatorFn<T>,
): (...params: T) => void {
  return (...x) => {
    const callback = (exploreState: ExploreState) =>
      mutator(
        {
          dashboard: exploreState,
        },
        ...x,
      );
    connectedMutators.updateDashboard(callback);
  };
}

/**
 * Takes an object containing `DashboardMutatorFn`s,
 * and returns an object of functions that directly update the dashboard.
 */
function createDashboardUpdaters<T extends DashboardMutatorFns>(
  connectedMutators: DashboardConnectedMutators,
  mutators: T,
): DashboardUpdaters<T> {
  return Object.fromEntries(
    Object.entries(mutators).map(([key, mutator]) => [
      key,
      dashboardMutatorToUpdater(connectedMutators, mutator),
    ]),
  ) as DashboardUpdaters<T>;
}
