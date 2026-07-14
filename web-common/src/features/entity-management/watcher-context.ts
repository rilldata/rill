import type { Writable } from "svelte/store";
import type {
  ConnectionStatus,
  FileAndResourceWatcher,
} from "./file-and-resource-watcher";

export const WATCHER_CONTEXT_KEY = Symbol("file-and-resource-watcher");

export interface WatcherContext {
  watcher: FileAndResourceWatcher;
  status: Writable<ConnectionStatus>;
}
