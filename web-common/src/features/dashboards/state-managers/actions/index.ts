import { sortActions } from "./sorting";
import { contextColActions } from "./context-columns";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import { setLeaderboardMeasureName } from "./core-actions";
import { dimTableActions } from "./dimension-table";
import type {
  DashboardCallbackExecutor,
  DashboardMutatorFn,
  DashboardMutatorFns,
  DashboardUpdaters,
} from "./types";

export type StateManagerActions = ReturnType<typeof createStateManagerActions>;

export const createStateManagerActions = (
  updateDashboard: DashboardCallbackExecutor
) => {
  return {
    /**
     * Actions related to the sorting state of the dashboard.
     */
    sorting: createDashboardUpdaters(updateDashboard, sortActions),
    /**
     * Actions related to the dashboard context columns.
     */
    contextCol: createDashboardUpdaters(updateDashboard, contextColActions),
    /**
     * Actions related to the dimension table.
     */
    dimTable: createDashboardUpdaters(updateDashboard, dimTableActions),
    // Note: for now, some core actions are kept in the root of the
    // actions object. Can revisit that later if we want to move them.
    setLeaderboardMeasureName: dashboardMutatorToUpdater(
      updateDashboard,
      setLeaderboardMeasureName
    ),
  };
};

/**
 * `dashboardMutatorToUpdater` take a DashboardCallbackExecutor
 * and returns a DashboardMutatorFn that directly updates the dashboard
 * by calling the DashboardCallbackExecutor.
 **/
function dashboardMutatorToUpdater<T extends unknown[]>(
  updateDashboard: DashboardCallbackExecutor,
  mutator: DashboardMutatorFn<T>
): (...params: T) => void {
  return (...x) => {
    const callback = (dash: MetricsExplorerEntity) => mutator(dash, ...x);
    updateDashboard(callback);
  };
}

/**
 * Takes an object containing `DashboardMutatorFn`s,
 * and returns an object of functions that directly update the dashboard.
 */
function createDashboardUpdaters<T extends DashboardMutatorFns>(
  updateDashboard: DashboardCallbackExecutor,
  mutators: T
): DashboardUpdaters<T> {
  return Object.fromEntries(
    Object.entries(mutators).map(([key, mutator]) => [
      key,
      dashboardMutatorToUpdater(updateDashboard, mutator),
    ])
  ) as DashboardUpdaters<T>;
}
