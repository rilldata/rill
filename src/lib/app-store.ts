import type { Socket } from "socket.io";
import { io } from "socket.io-client";
import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { DataModelerState } from "../types";

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

export function createStore() : AppStore {
	const socket = io("http://localhost:3001");
	socket.on("connect", () => {});
	const store:Writable<DataModelerState> = writable({queries:[], sources:[], metricsModels:[], exploreConfigurations: [], status: undefined});
	socket.on("app-state", (state:DataModelerState) => {
		store.set(state);
	});
	return {
		subscribe: store.subscribe,
		// @ts-ignore
		socket,
		reset() {
			socket.emit('reset');
		},
		action(name:string, args:any) {
			socket.emit(name, args);
		}
	}
}
