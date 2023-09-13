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

export type EntityCreateFunction = (
  resource: V1Resource,
  sourceName: string,
  pathPrefix?: string
) => Promise<void>;

export type ChainParams = {
  chainFunction: EntityCreateFunction;
  sourceName: string;
  pathPrefix?: string;
};

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
  telemetryParams: TelemetryParams;

  chainParams?: ChainParams;
};
type EntityActionQueueReducers = {
  add: (
    name: string,
    action: EntityAction,
    telemetryParams: TelemetryParams,
    chainParams?: ChainParams
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
    telemetryParams: TelemetryParams,
    chainParams?: ChainParams
  ) {
    update((state) => {
      state.entities[name] ??= [];
      state.entities[name].push({
        action,
        telemetryParams,
        chainParams,
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

      if (action.chainParams) {
        action.chainParams.chainFunction(
          resource,
          action.chainParams.sourceName,
          action.chainParams.pathPrefix
        );
      }

      return state;
    });
  },
};
