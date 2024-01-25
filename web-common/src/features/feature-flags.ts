import type { BeforeNavigate } from "@sveltejs/kit";
import { writable, derived } from "svelte/store";
import type { Readable } from "svelte/store";

export type DerivedReadables<T> = {
  [K in keyof T]: Readable<T[K]>;
};

const urlParams = new URLSearchParams(window.location.search);

const features = urlParams.get("features")?.split(",") ?? [];

const flags = {
  adminServer: false,
  readOnly: false,
  pivot: features?.includes("pivot") || false,
};

export type FeatureFlags = typeof flags;

export const featureFlags = (() => {
  const data = writable<FeatureFlags>(flags);
  const { subscribe, update } = data;

  const derivedGetters = Object.keys(flags).reduce(
    (acc, flag: keyof FeatureFlags) => {
      acc[flag] = derived(data, ($flags) => $flags[flag]);
      return acc;
    },
    {} as DerivedReadables<FeatureFlags>,
  );

  return {
    subscribe,
    set(bool: boolean, ...toggleFlags: (keyof FeatureFlags)[]) {
      update((flags) => {
        toggleFlags.forEach((n) => {
          flags[n] = bool;
        });

        return flags;
      });
    },
    toggle(...toggleFlags: (keyof FeatureFlags)[]) {
      update((flags) => {
        toggleFlags.forEach((n) => {
          flags[n] = !flags[n];
        });

        return flags;
      });
    },

    ...derivedGetters,
  };
})();

export function retainFeaturesFlags(navigation: BeforeNavigate) {
  const featureFlags = navigation?.from?.url?.searchParams.get("features");
  if (!featureFlags) return;

  navigation?.to?.url.searchParams.set("features", featureFlags);
}
