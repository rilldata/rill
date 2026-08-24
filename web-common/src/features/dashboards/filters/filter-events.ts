import type { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

/**
 * What caused a filter change. `undefined` is the filter bar itself.
 * Anything else identifies the component that applied the filter,
 * a canvas component id for a click to filter interaction for example.
 */
export type FilterChangeSource = string | undefined;

export type FilterEvents = {
  /** The filter param settled on a new value. Diffed and emitted by `ExpressionFilterManager.createListener`. */
  "state-changed": boolean;
  /** A filter was mutated. Emitted synchronously by the manager that was mutated. */
  "filter-changed": { source: FilterChangeSource };
};

/**
 * Every manager under an `ExpressionFilterManager` shares the root's emitter, so a leaf manager
 * can report a change without the managers in between having to forward it.
 *
 * The manager tree is rebuilt from the filter param on every change, and `$derived` has no
 * disposal hook, so a chain of per-manager subscriptions would have no place to be torn down.
 */
export type FilterEventEmitter = EventEmitter<FilterEvents>;
