import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, writable } from "svelte/store";
import {
  createRuntimeServiceGetInstance,
  runtimeServiceGetInstance,
  type V1InstanceFeatureFlags,
} from "../runtime-client";
import type { RuntimeClient } from "../runtime-client/v2";

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

  // Flags with "rill" scope describe the capabilities of the environment the frontend is running in,
  // as opposed to "user" scope flags which are project settings from rill.yaml.
  // They are not reset by project feature flags and must be set by the app
  // synchronously at the top of its root layout's script (not in onMount or after an await),
  // because some children read them once on mount (e.g. ExportMenu decides whether to import the admin client).

  // True when the frontend has a user session with the admin server (i.e. Rill Cloud).
  adminServer = new FeatureFlag("rill", false);
  // True in embeds that were given an admin token and user email.
  // Embeds have no user session with the admin server, but can call the subset of admin APIs
  // that embed admin tokens are permitted to call (see ProjectPermissionsForEmbedAuthToken in admin/permissions.go).
  // Only use it to gate features that embed admin tokens support.
  embedAdminActions = new FeatureFlag("rill", false);
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
  // Do not check reports directly to gate UI; use reportsEnabled instead,
  // which also checks that the environment can reach the admin APIs for them.
  alerts = new FeatureFlag("user", true);
  reports = new FeatureFlag("user", true);
  chat = new FeatureFlag("user", true);
  chatMemory = new FeatureFlag("user", true);
  dashboardChat = new FeatureFlag("user", false);
  developerChat = new FeatureFlag("user", false);
  deploy = new FeatureFlag("user", true);
  stickyDashboardState = new FeatureFlag("user", false);
  cloudEditing = new FeatureFlag("user", false);
  customCharts = new FeatureFlag("user", false);
  customComponents = new FeatureFlag("user", false);
  disablePersistentDashboardState = new FeatureFlag("user", false);
  personalCanvases = new FeatureFlag("user", false);

  private flagsUnsub?: () => void;

  constructor() {
    this.ready = new Promise<void>((resolve) => {
      this._resolveReady = resolve;
    });
  }

  clearRuntimeClient() {
    this.flagsUnsub?.();
    this.flagsUnsub = undefined;
  }

  setRuntimeClient(client: RuntimeClient) {
    this.flagsUnsub?.();

    this.flagsUnsub = createRuntimeServiceGetInstance(
      client,
      {},
      {
        query: {
          select: (data) => data?.instance?.featureFlags,
        },
      },
      queryClient,
    ).subscribe((features) => {
      if (features.data) this.updateFlags(features.data);
    });
  }

  private updateFlags(userFlags: V1InstanceFeatureFlags) {
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

// Reports are managed through the admin server,
// so they need either a Rill Cloud session or an embed with an admin token.
// They are never available in Rill Developer.
export const reportsEnabled = derived(
  [
    featureFlags.reports,
    featureFlags.adminServer,
    featureFlags.embedAdminActions,
  ],
  ([reports, adminServer, embedAdminActions]) =>
    reports && (adminServer || embedAdminActions),
);

export async function getFeatureFlags(
  client: RuntimeClient,
): Promise<V1InstanceFeatureFlags> {
  try {
    const data = await runtimeServiceGetInstance(client, {});
    return data.instance?.featureFlags ?? {};
  } catch {
    return {};
  }
}
