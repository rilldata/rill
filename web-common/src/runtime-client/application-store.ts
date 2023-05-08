/**
 * The ApplicationStore contains the state of the general application.
 * It does not contain any of the state for the entities; instead, it contains information
 * about things like the active entity and the application status.
 */
import { writable } from "svelte/store";

// Store to show application running status based on Query Queue
export const appQueryStatusStore = writable<boolean>(false);
