import { writable, type Writable } from "svelte/store";

export { localStorageStore } from "./local-storage";
export { previousValueStore } from "./previous-value-store";

export function maybeWritable<T>(value?: T): Writable<T | undefined> {
  return writable(value);
}
