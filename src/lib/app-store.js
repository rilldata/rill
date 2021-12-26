import { io } from "socket.io-client";
import { writable } from "svelte/store";

export function createStore() {
	const socket = io("http://localhost:3001");
	socket.on("connect", () => {});
	const { set, subscribe } = writable({});
	socket.on("app-state", (state) => set(state));
	return {
		subscribe,
		socket,
		reset() {
			socket.emit('reset');
		},
		action(name, args) {
			socket.emit(name, args);
		}
	}
}
