import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

// Note: the types below are helper types to simplify the type inference
// used in the creation of StateManagerActions, so that we can have nice
// autocomplete and type checking in the IDE, while still keeping the
// code that is used to define actions organized and readable.

/**
 * A DashboardMutatorCallback is a function that mutates
 * a MetricsExplorerEntity, i.e., the data single dashboard.
 * This will often be a closure over other parameters
 * that are relevant to the mutation.
 */
export type DashboardMutatorCallback = (
  metricsExplorer: MetricsExplorerEntity
) => void;

/**
 * DashboardCallbackExecutor is a function that takes a
 * DashboardMutatorCallback and executes it. The
 * DashboardCallbackExecutor is a closure containing a reference
 * to the live dashboard, and therefore calling this function
 * on a DashboardMutatorCallback will actually update the dashboard.
 */
export type DashboardCallbackExecutor = (
  callback: DashboardMutatorCallback
) => void;

/**
 * A DashboardMutatorFn is a function mutates the data
 * model of a single dashboard.
 * It takes a reference to a dashboard as its first parameter,
 * and may take any number of additional parameters relevant to the mutation.
 */
export type DashboardMutatorFn<T extends unknown[]> = (
  dash: MetricsExplorerEntity,
  ...params: T
) => void;

export type DashboardMutatorFns = {
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
export type DashboardUpdaters<T extends DashboardMutatorFns> = Expand<{
  [P in keyof T]: (...params: DropFirst<Parameters<T[P]>>) => void;
}>;
