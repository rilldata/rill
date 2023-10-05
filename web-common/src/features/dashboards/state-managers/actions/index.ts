import { sortActions } from "./sorting";
import type {
  MetricsExplorerMutatorFn,
  UpdateDashboard2ndOrderCallback,
} from "../state-managers";
import { contextColActions } from "./context-columns";

export type StateManagerActions = ReturnType<typeof createStateManagerActions>;

export const createStateManagerActions = (
  updateDashboard: UpdateDashboard2ndOrderCallback
) => {
  return {
    sorting: createDashboardUpdatersFromMutatorProducers(
      updateDashboard,
      sortActions
    ),
    context: createDashboardUpdatersFromMutatorProducers(
      updateDashboard,
      contextColActions
    ),
    setLeaderboardMeasureName: createDashboardUpdaterFn(
      updateDashboard,
      (name: string) => (metricsExplorer) => {
        metricsExplorer.leaderboardMeasureName = name;
      }
    ),
  };
};

/**
 * A MetricsExplorerMutatorProducer is a higher order function that takes
 * parameters relevant to a specif mutation and returns
 * a MetricsExplorerMutatorFn.
 */
type MetricsExplorerMutatorProducer<T extends unknown[]> = (
  ...params: T
) => MetricsExplorerMutatorFn;

type MutatorProducers = {
  [key: string]: MetricsExplorerMutatorProducer<unknown[]>;
};

/**
 * A DashboardUpdaters object is a collection of functions that
 * directly update the dashboard.
 */
type DashboardUpdaters<T extends MutatorProducers> = {
  [P in keyof T]: (...params: Parameters<T[P]>) => void;
};

/**
 * `createDashboardUpdaterFn` take a MetricsExplorerMutatorProducer
 * and returns a that directly updates the dashboard.
 **/
function createDashboardUpdaterFn<T extends unknown[]>(
  updateDashboard: UpdateDashboard2ndOrderCallback,
  callback: MetricsExplorerMutatorProducer<T>
): (...params2: T) => void {
  return (...x) => updateDashboard(callback(...x));
}

/**
 * Takes an object containing mutator producer functions,
 * and returns an object of funcionts that directly update the dashboard.
 * @param updateDashboard
 * @param mutatorProducers
 * @returns
 */
function createDashboardUpdatersFromMutatorProducers<
  T extends MutatorProducers
>(
  updateDashboard: UpdateDashboard2ndOrderCallback,
  mutatorProducers: T
): DashboardUpdaters<T> {
  return Object.fromEntries(
    Object.entries(mutatorProducers).map(([key, value]) => [
      key,
      createDashboardUpdaterFn(updateDashboard, value),
    ])
  ) as DashboardUpdaters<T>;
}
