import type { Socket } from "socket.io";
import type { Writable } from "svelte/store";
import type { DataModelerState } from "./types";
import { clientFactory } from "$common/clientFactory";
import { RootConfig } from "$common/config/RootConfig";
import type { DataModelerSocketService } from "$common/socket/DataModelerSocketService";

interface ServerToClientEvents {
	['app-state']: (state:DataModelerState) => void;
};

interface ClientToServerEvents {
  }

export interface AppStore extends Pick<Writable<DataModelerState>, "subscribe"> {
	socket:Socket<ServerToClientEvents, ClientToServerEvents>;
}

const clientInstances = clientFactory(RootConfig.getDefaultConfig());
export const dataModelerService = clientInstances.dataModelerService;
export const dataModelerStateService = clientInstances.dataModelerStateService;
dataModelerService.init();

export function createStore() : AppStore {
	return {
		subscribe: dataModelerStateService.store.subscribe,
		// @ts-ignore
		socket: (dataModelerService as DataModelerSocketService).getSocket(),
	}
}
