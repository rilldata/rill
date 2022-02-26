import { writable } from "svelte/store";

export function createQueryHighlightStore() {
    const { subscribe, set } = writable(undefined);
    return {
        subscribe,
        set,
    }
}