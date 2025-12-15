import { WatchFilesClient } from "./WatchFilesClient";
import { WatchResourcesClient } from "./WatchResourcesClient";

export const fileWatcher = new WatchFilesClient().client;
export const resourceWatcher = new WatchResourcesClient().client;
