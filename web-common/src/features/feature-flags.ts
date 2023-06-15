import { writable } from "svelte/store";

export type FeatureFlags = {
  readOnly: boolean;
};
export const featureFlags = writable<FeatureFlags>({
  readOnly: undefined,
});
