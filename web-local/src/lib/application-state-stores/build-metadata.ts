import { writable } from "svelte/store";

export type ApplicationBuildMetadata = {
  version: string;
  commitHash: string;
};

export function createApplicationBuildMetadataStore() {
  const { subscribe, set } = writable<ApplicationBuildMetadata>({
    version: "",
    commitHash: "",
  });
  return {
    subscribe,
    set,
  };
}
