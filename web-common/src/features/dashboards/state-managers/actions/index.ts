import { sortActions } from "./sorting";
import type { DashboardCallbackExecutor } from "../state-managers";
import { contextColActions } from "./context-columns";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";
import { setLeaderboardMeasureName } from "./core-actions";
import { dimTableActions } from "./dimension-table";

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

// Note: the types below are helper types to simplify the type inference
// used in the creation of StateManagerActions, so that we can have nice
// autocomplete and type checking in the IDE, while still keeping the
// code that is used to define actions organized and readable.

/**
 * A DashboardMutatorFn is a function mutates the data
 * model of a single dashboard.
 * It takes a reference to a dashboard as its first parameter,
 * and may take any number of additional parameters relevant to the mutation.
 */
type DashboardMutatorFn<T extends unknown[]> = (
  dash: MetricsExplorerEntity,
  ...params: T
) => void;

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

type DashboardMutatorFns = {
  [key: string]: DashboardMutatorFn<unknown[]>;
};

/**
 * A helper type that drops the first element from a tuple.
 */
type DropFirst<T extends unknown[]> = T extends [unknown, ...infer U]
  ? U
  : never;

/**
 * A DashboardUpdaters object is a collection of functions that
 * directly update the live dashboard.
 */
type DashboardUpdaters2<T extends DashboardMutatorFns> = Expand<{
  [P in keyof T]: (...params: DropFirst<Parameters<T[P]>>) => void;
}>;

/**
 * Takes an object containing `DashboardMutatorFn`s,
 * and returns an object of functions that directly update the dashboard.
 */
function createDashboardUpdaters<T extends DashboardMutatorFns>(
  updateDashboard: DashboardCallbackExecutor,
  mutators: T
): DashboardUpdaters2<T> {
  return Object.fromEntries(
    Object.entries(mutators).map(([key, mutator]) => [
      key,
      dashboardMutatorToUpdater(updateDashboard, mutator),
    ])
  ) as DashboardUpdaters2<T>;
}
