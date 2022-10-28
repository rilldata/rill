import { writable } from "svelte/store";

export interface SourceEntity {
  id: string;
  sqlDraft: string;
}

export interface SourceStoreType {
  entities: Record<string, SourceEntity>;
}

export const sourceStore = writable({
  entities: {},
} as SourceStoreType);
