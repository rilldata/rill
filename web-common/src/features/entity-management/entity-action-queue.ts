import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { sourceIngestionTelemetry } from "@rilldata/web-common/features/sources/source-ingestion-telemetry";
import type { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import type { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { Readable, writable } from "svelte/store";

export enum EntityAction {
  Create,
  Update,
  Rename,
  Delete,
}

/**
 * A global queue for entity actions.
 * This is used to emit telemetry for async response from reconcile.
 */
export type EntityActionQueueState = {
  entities: Record<string, Array<EntityActionInstance>>;
};
export type EntityActionInstance = {
  action: EntityAction;

  // telemetry
  screenName: MetricsEventScreenName;
  behaviourEventMedium: BehaviourEventMedium;
};
type EntityActionQueueReducers = {
  add: (name: string, actionInstance: EntityActionInstance) => void;
  resolved: (resource: V1Resource) => void;
};
export type EntityActionQueueStore = Readable<EntityActionQueueState> &
  EntityActionQueueReducers;

// TODO: how does reconcile handle cases like, create source => rename soon after before ingestion is completed
const { update, subscribe } = writable<EntityActionQueueState>({
  entities: {},
});

export const entityActionQueueStore: EntityActionQueueStore = {
  subscribe,

  add(name: string, actionInstance: EntityActionInstance) {
    update((state) => {
      state.entities[name] ??= [];
      state.entities[name].push(actionInstance);
      return state;
    });
  },

  resolved(resource: V1Resource) {
    update((state) => {
      if (!state.entities[resource.meta.name.name]?.length) return state;

      if (resource.meta.renamedFrom) {
        // TODO: rename telemetry
        return state;
      }

      switch (resource.meta.name.kind) {
        case ResourceKind.Source:
          sourceIngestionTelemetry(
            resource,
            state.entities[resource.meta.name.name].shift()
          );
      }

      return state;
    });
  },
};
