import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { sourceIngestionTelemetry } from "@rilldata/web-common/features/sources/source-ingestion-telemetry";
import type { TelemetryParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
import type { V1ResourceEvent } from "@rilldata/web-common/runtime-client";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { Readable, writable } from "svelte/store";

export enum EntityAction {
  Create,
  Update,
  Rename,
  Delete,
}

export enum ChainAction {
  ModelFromSource,
  DashboardFromModel,
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
  chain?: Array<ChainAction>;

  // telemetry
  params: TelemetryParams;
};
type EntityActionQueueReducers = {
  add: (
    name: string,
    action: EntityAction,
    params: TelemetryParams,
    chain?: Array<ChainAction>
  ) => void;
  resolved: (resource: V1Resource, event: V1ResourceEvent) => void;
};
export type EntityActionQueueStore = Readable<EntityActionQueueState> &
  EntityActionQueueReducers;

// TODO: how does reconcile handle cases like, create source => rename soon after before ingestion is completed
const { update, subscribe } = writable<EntityActionQueueState>({
  entities: {},
});

export const entityActionQueueStore: EntityActionQueueStore = {
  subscribe,

  add(
    name: string,
    action: EntityAction,
    params: TelemetryParams,
    chain?: Array<ChainAction>
  ) {
    update((state) => {
      state.entities[name] ??= [];
      state.entities[name].push({
        action,
        params,
        chain,
      });
      return state;
    });
  },

  resolved(resource: V1Resource, _: V1ResourceEvent) {
    update((state) => {
      if (!state.entities[resource.meta.name.name]?.length) return state;

      if (resource.meta.renamedFrom) {
        // TODO: rename telemetry
        return state;
      }

      const action = state.entities[resource.meta.name.name].shift();

      switch (resource.meta.name.kind) {
        case ResourceKind.Source:
          sourceIngestionTelemetry(resource, action);
          break;
      }

      if (action.chain?.length) {
        switch (action.chain[0]) {
          case ChainAction.DashboardFromModel:
            this.add();
            break;
          // TODO: others
        }
      }

      return state;
    });
  },
};
