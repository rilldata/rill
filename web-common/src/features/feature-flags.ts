import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { writable } from "svelte/store";
import {
  getRuntimeServiceGetInstanceQueryKey,
  runtimeServiceGetInstance,
  type V1InstanceFeatureFlags,
} from "../runtime-client";

class FeatureFlag {
  private _internal = false;
  private _default: boolean;
  private state = writable(false);
  subscribe = this.state.subscribe;

  constructor(scope: "user" | "rill", defaultValue: boolean) {
    this._internal = scope === "rill";
    this._default = defaultValue;
    this.set(defaultValue);
  }

  get internalOnly() {
    return this._internal;
  }

  get defaultValue() {
    return this._default;
  }

  toggle = () => this.state.update((n) => !n);
  set = (n: boolean) => this.state.set(n);
  resetToDefault = () => this.set(this._default);
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

  // These are fallback defaults in case of issues in parsing rill.yaml.
  // Full defaults are in defaultFeatureFlags in runtime/drivers/registry.go
  ai = new FeatureFlag("user", !import.meta.env.VITE_PLAYWRIGHT_TEST);
  exports = new FeatureFlag("user", true);
  cloudDataViewer = new FeatureFlag("user", false);
  dimensionSearch = new FeatureFlag("user", false);
  twoTieredNavigation = new FeatureFlag("user", false);
  rillTime = new FeatureFlag("user", true);
  hidePublicUrl = new FeatureFlag("user", false);
  exportHeader = new FeatureFlag("user", false);
  alerts = new FeatureFlag("user", true);
  reports = new FeatureFlag("user", true);
  darkMode = new FeatureFlag("user", true);
  chat = new FeatureFlag("user", true);
  dashboardChat = new FeatureFlag("user", false);
  developerChat = new FeatureFlag("user", false);
  deploy = new FeatureFlag("user", true);
  generateCanvas = new FeatureFlag("user", false);

  constructor() {
    this.ready = new Promise<void>((resolve) => {
      this._resolveReady = resolve;
    });
  }

  updateFlags = (userFlags: V1InstanceFeatureFlags) => {
    this._resolveReady();

    // First, reset all user flags to their defaults
    const userFlagKeys = Object.keys(this).filter((key) => {
      const flag = this[key];
      return flag instanceof FeatureFlag && !flag.internalOnly;
    });

    for (const key of userFlagKeys) {
      const flag = this[key] as FeatureFlag;
      flag.resetToDefault();
    }

    // Then apply project-specific overrides
    for (const key in userFlags) {
      const flag = this[key] as FeatureFlag | undefined;
      if (!flag || flag.internalOnly) continue;
      flag.set(userFlags[key]);
    }
  };

  setInstanceId = async (instanceId: string) => {
    const response = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetInstanceQueryKey(instanceId),
      queryFn: () => runtimeServiceGetInstance(instanceId),
    });

    if (response?.instance?.featureFlags) {
      this.updateFlags(response?.instance?.featureFlags);
    }
  };

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
