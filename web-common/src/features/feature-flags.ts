import type { BeforeNavigate } from "@sveltejs/kit";
import type { Readable } from "svelte/store";
import { writable } from "svelte/store";
import { parse } from "yaml";
import { createRuntimeServiceGetFile } from "../runtime-client";
import { runtime } from "../runtime-client/runtime-store";
import { debounce } from "../lib/create-debouncer";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export type DerivedReadables<T> = {
  [K in keyof T]: Readable<T[K]>;
};

class FeatureFlag {
  private _system = false;
  private state = writable(false);

  constructor(scope: "user" | "system" = "user", initial = false) {
    this._system = scope === "system";
    this.set(initial);
  }

  get system() {
    return this._system;
  }

  subscribe = this.state.subscribe;
  toggle = () => this.state.update((n) => !n);
  set = (n: boolean) => this.state.set(n);
}

type FeatureFlagKey = keyof Omit<FeatureFlags, "set">;

class FeatureFlags {
  adminServer = new FeatureFlag("system");
  readOnly = new FeatureFlag("system");
  ai = new FeatureFlag("system", true);
  pivot = new FeatureFlag();
  cloudDataViewer = new FeatureFlag();
  customDashboards = new FeatureFlag();

  constructor() {
    const urlParams = new URLSearchParams(window.location.search);
    const urlFlags = (urlParams.get("features") ?? "").split(",");

    const localFlags = (localStorage.getItem("features") ?? "").split(",");

    const staticFlags = [...urlFlags, ...localFlags];

    const updateFlags = debounce((userFlags: Set<string>) => {
      Object.keys(this).forEach((key) => {
        const flag = this[key] as FeatureFlag;
        if (flag.system) return;
        if (userFlags.has(key)) flag.set(true);
      });
    }, 400);

    // Responsively update flags based rill.yaml
    runtime.subscribe((runtime) => {
      if (!runtime?.instanceId) return;

      createRuntimeServiceGetFile(runtime.instanceId, "rill.yaml", {
        query: {
          select: (data) => {
            let features: string[] = [];
            try {
              const projectData = parse(data?.blob ?? "") as {
                features?: string[];
              };
              features = projectData?.features ?? [];
            } catch (e) {
              console.error(e);
            }
            return features;
          },
          queryClient,
        },
      }).subscribe((features) => {
        if (!Array.isArray(features?.data)) return;
        const yamlFlags = features.data;
        updateFlags(new Set([...staticFlags, ...yamlFlags]));
      });
    });
  }

  get set() {
    return (bool: boolean, ...toggleFlags: FeatureFlagKey[]) => {
      toggleFlags.forEach((n) => {
        const flag = this[n] as FeatureFlag | undefined;
        if (!flag) return;
        flag.set(bool);
      });
    };
  }
}

export const featureFlags = new FeatureFlags();

export function retainFeaturesFlags(navigation: BeforeNavigate) {
  const featureFlags = navigation?.from?.url?.searchParams.get("features");
  if (!featureFlags) return;
  navigation?.to?.url.searchParams.set("features", featureFlags);
}
