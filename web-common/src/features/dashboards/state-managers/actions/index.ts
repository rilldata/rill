import { sortActions } from "./sorting";
import type {
  // DashboardMutatorCallback,
  DashboardCallbackExecutor,
} from "../state-managers";
import { contextColActions } from "./context-columns";
import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export type StateManagerActions = ReturnType<typeof createStateManagerActions>;

//////////////////////////

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
 * A DashboardUpdaterFn is a function that takes a parameters
 * relevant to a specific mutation, and applies that mutation to
 * the dashboard. In order to apply the mutation, the updaterFn
 * is a closure over a reference to a DashboardCallbackExecutor
 */
type DashboardUpdaterFn<T extends unknown[]> = (...params: T) => void;

// type UpdaterFromMutator<T extends DashboardMutatorFn<S extends unknown[]>> = DashboardUpdaterFn<S>

/**
 * `createDashboardUpdaterFn` take a DashboardMutatorCallbackProducer
 * and returns a that directly updates the dashboard.
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

// dashboardMutatorToUpdater(

type DashboardMutatorFns = {
  [key: string]: DashboardMutatorFn<unknown[]>;
};

type DropFirst<T extends unknown[]> = T extends [unknown, ...infer U]
  ? U
  : never;
/**
 * A DashboardUpdaters object is a collection of functions that
 * directly update the live dashboard.
 */
type DashboardUpdaters2<T extends DashboardMutatorFns> = Expand<{
  // [P in keyof T]: (...params: Parameters<T[P]>) => void;
  [P in keyof T]: (...params: DropFirst<Parameters<T[P]>>) => void;
}>;

// type TestMutators = {
//   foo: (dash: MetricsExplorerEntity) => void;
//   bar: (dash: MetricsExplorerEntity, a: number) => void;
//   bat: (dash: MetricsExplorerEntity, b: number, c: string) => void;
// };

// type TestUpdaters = DashboardUpdaters2<TestMutators>;

// type TestUpdaters2 = DashboardUpdaters2<typeof sortActions>;

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

export const createStateManagerActions = (
  updateDashboard: DashboardCallbackExecutor
) => {
  return {
    sorting: createDashboardUpdaters(updateDashboard, sortActions),
    context: createDashboardUpdaters(updateDashboard, contextColActions),
    setLeaderboardMeasureName: dashboardMutatorToUpdater(
      updateDashboard,
      (dash, name: string) => {
        dash.leaderboardMeasureName = name;
      }
    ),
  };
};
