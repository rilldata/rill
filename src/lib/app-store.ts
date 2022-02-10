import type { Socket } from "socket.io";
import type { Writable } from "svelte/store";
import type { DataModelerState } from "./types";
import { clientFactory } from "$common/clientFactory";
import { RootConfig } from "$common/config/RootConfig";

interface ServerToClientEvents {
	['app-state']: (state:DataModelerState) => void;
};

interface ClientToServerEvents {
  }

export interface AppStore extends Pick<Writable<DataModelerState>, "subscribe"> {
	socket:Socket<ServerToClientEvents, ClientToServerEvents>;
	reset:Function;
	action:Function;
}

const clientInstances = clientFactory(RootConfig.getDefaultConfig());
export const dataModelerService = clientInstances.dataModelerService;
export const dataModelerStateService = clientInstances.dataModelerStateService;
dataModelerService.init();

export function createStore() : AppStore {
	return {
		subscribe: dataModelerStateService.store.subscribe,
		// @ts-ignore
		socket: null,
		reset() {
			// socket.emit('reset');
		},
		action(name: any, args: any) {
			dataModelerService.dispatch(name, args);
		}
	}
}
