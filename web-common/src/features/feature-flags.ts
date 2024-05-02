import type { BeforeNavigate } from "@sveltejs/kit";
import { writable } from "svelte/store";
import {
  createRuntimeServiceGetInstance,
  V1InstanceFeatureFlags,
} from "../runtime-client";
import { runtime } from "../runtime-client/runtime-store";
import { debounce } from "../lib/create-debouncer";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

class FeatureFlag {
  private state = writable(false);
  subscribe = this.state.subscribe;

  constructor(scope: "user" | "system" = "user", initial = false) {
    this._system = scope === "system";
    this.set(initial);
  }

  private _system = false;

  get system() {
    return this._system;
  }

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

    const staticFlags = new Set([...urlFlags, ...localFlags]);

    const updateFlags = debounce((userFlags: V1InstanceFeatureFlags) => {
      Object.keys(this).forEach((key) => {
        const flag = this[key] as FeatureFlag;
        if (flag.system) return;
        if (staticFlags.has(key))
          flag.set(true); // precedence to flags from url and localStorage
        else if (key in userFlags) flag.set(userFlags[key]);
      });
    }, 400);

    // Responsively update flags based rill.yaml
    runtime.subscribe((runtime) => {
      if (!runtime?.instanceId) return;

      createRuntimeServiceGetInstance(runtime.instanceId, {
        query: {
          select: (data) => data?.instance?.featureFlags,
          queryClient,
        },
      }).subscribe((features) => {
        if (features.data) updateFlags(features.data);
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
