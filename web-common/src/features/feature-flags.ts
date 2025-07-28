import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { writable } from "svelte/store";
import {
  createRuntimeServiceGetInstance,
  type V1InstanceFeatureFlags,
} from "../runtime-client";
import { runtime } from "../runtime-client/runtime-store";

class FeatureFlag {
  private _internal = false;
  private state = writable(false);
  subscribe = this.state.subscribe;

  constructor(scope: "user" | "rill", initial: boolean) {
    this._internal = scope === "rill";
    this.set(initial);
  }

  get internalOnly() {
    return this._internal;
  }

  toggle = () => this.state.update((n) => !n);
  set = (n: boolean) => this.state.set(n);
}

type FeatureFlagKey = keyof Omit<FeatureFlags, "set">;

class FeatureFlags {
  ready: Promise<void>;
  private _resolveReady!: () => void;

  adminServer = new FeatureFlag("rill", false);
  readOnly = new FeatureFlag("rill", false);
  // Until we figure out a good way to test managed github we need to use the legacy archive method.
  // Right now this is true only in an E2E environment.
  legacyArchiveDeploy = new FeatureFlag(
    "rill",
    !!import.meta.env.VITE_PLAYWRIGHT_TEST,
  );

  ai = new FeatureFlag("user", !import.meta.env.VITE_PLAYWRIGHT_TEST);
  exports = new FeatureFlag("user", true);
  cloudDataViewer = new FeatureFlag("user", false);
  dimensionSearch = new FeatureFlag("user", false);
  clickhouseModeling = new FeatureFlag("user", false);
  twoTieredNavigation = new FeatureFlag("user", false);
  rillTime = new FeatureFlag("user", false);
  hidePublicUrl = new FeatureFlag("user", false);
  exportHeader = new FeatureFlag("user", false);
  alerts = new FeatureFlag("user", true);
  reports = new FeatureFlag("user", true);
  darkMode = new FeatureFlag("user", false);
  chat = new FeatureFlag("user", false);

  constructor() {
    this.ready = new Promise<void>((resolve) => {
      this._resolveReady = resolve;
    });

    const updateFlags = (userFlags: V1InstanceFeatureFlags) => {
      this._resolveReady();
      for (const key in userFlags) {
        const flag = this[key] as FeatureFlag | undefined;
        if (!flag || flag.internalOnly) continue;
        flag.set(userFlags[key]);
      }
    };

    // Responsively update flags based rill.yaml
    runtime.subscribe((runtime) => {
      if (!runtime?.instanceId) return;

      createRuntimeServiceGetInstance(
        runtime.instanceId,
        undefined,
        {
          query: {
            select: (data) => data?.instance?.featureFlags,
          },
        },
        queryClient,
      ).subscribe((features) => {
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
