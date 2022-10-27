/**
 * The ApplicationStore contains the state of the general application.
 * It does not contain any of the state for the entities; instead, it contains information
 * about things like the active entity and the application status.
 */
import { clientFactory } from "@rilldata/web-local/common/clientFactory";
import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type {
  ApplicationEntity,
  ApplicationState,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type {
  EntityRecord,
  EntityState,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataModelerSocketService } from "@rilldata/web-local/common/socket/DataModelerSocketService";
import type {
  ClientToServerEvents,
  ServerToClientEvents,
} from "@rilldata/web-local/common/socket/SocketInterfaces";
import type { Socket } from "socket.io";
import { writable, Writable } from "svelte/store";

export interface AppStore<
  Entity extends EntityRecord = EntityRecord,
  State extends EntityState<Entity> = EntityState<Entity>
> extends Pick<Writable<State>, "subscribe"> {
  socket?: Socket<ServerToClientEvents, ClientToServerEvents>;
}

export const config = new RootConfig({});

const clientInstances = clientFactory(config);
export const dataModelerService = clientInstances.dataModelerService;
export const metricsService = clientInstances.metricsService;
export const dataModelerStateService = clientInstances.dataModelerStateService;
dataModelerService.init();

export type ApplicationStore = AppStore<ApplicationEntity, ApplicationState>;

export function createStore(): ApplicationStore {
  return {
    subscribe: dataModelerStateService.getEntityStateService(
      EntityType.Application,
      StateType.Derived
    ).store.subscribe as any,
    // FIXME: what is happening with these types
    // @ts-ignore
    socket: (dataModelerService as DataModelerSocketService).getSocket(),
  };
}

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None
);

export const duplicateSourceName: Writable<string> = writable(null);

export type RuntimeState = {
  repoId?: string;
  instanceId: string;
};
export const runtimeStore = writable<RuntimeState>({
  repoId: null,
  instanceId: null,
});
