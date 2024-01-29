import type { BeforeNavigate } from "@sveltejs/kit";
import { writable } from "svelte/store";

export type FeatureFlags = {
  adminServer?: boolean;
  readOnly?: boolean;
  alerts?: boolean;
};
export const featureFlags = writable<FeatureFlags>();

export function retainFeaturesFlags(navigation: BeforeNavigate) {
  const featureFlags = navigation?.from?.url?.searchParams.get("features");
  if (!featureFlags) return;

  navigation?.to?.url.searchParams.set("features", featureFlags);
}
