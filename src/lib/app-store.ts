import type { Socket } from "socket.io";
import type { Writable } from "svelte/store";
import { clientFactory } from "$common/clientFactory";
import { RootConfig } from "$common/config/RootConfig";
import type { DataModelerSocketService } from "$common/socket/DataModelerSocketService";
import type { ClientToServerEvents, ServerToClientEvents } from "$common/socket/SocketInterfaces";
import type {
	EntityRecord,
	EntityState
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
	ApplicationEntity,
	ApplicationState
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";

export interface AppStore<
	Entity extends EntityRecord = EntityRecord, State extends EntityState<Entity> = EntityState<Entity>
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

export function createStore() : ApplicationStore {
	return {
		subscribe: dataModelerStateService
			.getEntityStateService(EntityType.Application, StateType.Derived).store.subscribe as any,
		// @ts-ignore
		socket: (dataModelerService as DataModelerSocketService).getSocket(),
	}
}
