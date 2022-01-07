import type { Socket } from "socket.io";
import { io } from "socket.io-client";
import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { DataModellerState } from "../types";

interface ServerToClientEvents {
	['app-state']: (state:DataModellerState) => void;
};

interface ClientToServerEvents {
  }

export interface AppStore extends Pick<Writable<DataModellerState>, "subscribe"> {
	socket:Socket<ServerToClientEvents, ClientToServerEvents>;
	reset:Function;
	action:Function;
}

export function createStore() : AppStore {
	const socket = io("http://localhost:3001");
	socket.on("connect", () => {});
	const store:Writable<DataModellerState> = writable({queries:[], sources:[], status: undefined});
	socket.on("app-state", (state:DataModellerState) => {
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
