import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { useQueryClient } from "@tanstack/svelte-query";
import { derived, writable, type Readable } from "svelte/store";
import {
  createRuntimeServiceGetInstance,
  runtimeServiceGetInstance,
  type V1InstanceFeatureFlags,
} from "../runtime-client";
import type { RuntimeClient } from "../runtime-client/v2";
import { useRuntimeClient } from "../runtime-client/v2/context";

// ── Runtime feature flags (per-deployment) ─────────────────────────────
//
// These flags are hydrated from `instance.featureFlags` (sourced from the
// project's `rill.yaml`). Each `<RuntimeProvider>` owns a `RuntimeClient`,
// and `useFeatureFlags()` derives reactive stores from a `GetInstance` query
// against that client. Different runtimes have independent flag values, and
// switching runtimes (e.g. branch swap) updates flags without page refresh.
//
// Defaults below act as a fallback when a key is missing from the response.
// Source-of-truth defaults live in `runtime/drivers/registry.go`.

const RUNTIME_FLAG_DEFAULTS = {
  ai: !import.meta.env.VITE_PLAYWRIGHT_TEST,
  exports: true,
  cloudDataViewer: false,
  dimensionSearch: false,
  twoTieredNavigation: false,
  rillTime: true,
  hidePublicUrl: false,
  exportHeader: false,
  alerts: true,
  reports: true,
  chat: true,
  dashboardChat: false,
  developerChat: false,
  deploy: true,
  stickyDashboardState: false,
  cloudEditing: false,
  customCharts: false,
} as const satisfies Record<string, boolean>;

export type RuntimeFeatureFlagKey = keyof typeof RUNTIME_FLAG_DEFAULTS;

export type RuntimeFeatureFlags = {
  [K in RuntimeFeatureFlagKey]: Readable<boolean>;
};

const RUNTIME_FLAG_KEYS = Object.keys(
  RUNTIME_FLAG_DEFAULTS,
) as RuntimeFeatureFlagKey[];

/**
 * Returns reactive stores for the runtime feature flags of the nearest
 * `<RuntimeProvider>` ancestor's `RuntimeClient`. Must be called during
 * component initialization (top-level `<script>`).
 *
 * Backed by the `GetInstance` TanStack query. The query is cached per
 * `instanceId`, so multiple components share a single underlying request.
 */
export function useFeatureFlags(): RuntimeFeatureFlags {
  const client = useRuntimeClient();
  const qc = useQueryClient();
  const flagsQuery = createRuntimeServiceGetInstance(
    client,
    {},
    { query: { select: (data) => data?.instance?.featureFlags ?? {} } },
    qc,
  );
  return Object.fromEntries(
    RUNTIME_FLAG_KEYS.map((name) => [
      name,
      derived(
        flagsQuery,
        ($q) => ($q.data?.[name] ?? RUNTIME_FLAG_DEFAULTS[name]) as boolean,
      ),
    ]),
  ) as RuntimeFeatureFlags;
}

/**
 * One-shot fetch of the current runtime feature flags. Useful in load
 * functions and other non-Svelte contexts where reactive subscription is
 * unnecessary. Returns the raw map; callers apply defaults themselves.
 */
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

// ── Legacy singleton (deprecated) ──────────────────────────────────────
//
// Kept temporarily so app-wide and runtime callsites can migrate piecemeal.
// Slated for deletion once all callers have moved to `useFeatureFlags()`
// and `app-flags.ts`.

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
  legacyArchiveDeploy = new FeatureFlag(
    "rill",
    !!import.meta.env.VITE_PLAYWRIGHT_TEST,
  );

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
  chat = new FeatureFlag("user", true);
  dashboardChat = new FeatureFlag("user", false);
  developerChat = new FeatureFlag("user", false);
  deploy = new FeatureFlag("user", true);
  stickyDashboardState = new FeatureFlag("user", false);
  cloudEditing = new FeatureFlag("user", false);
  customCharts = new FeatureFlag("user", false);

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

    const userFlagKeys = Object.keys(this).filter((key) => {
      const flag = this[key];
      return flag instanceof FeatureFlag && !flag.internalOnly;
    });

    for (const key of userFlagKeys) {
      const flag = this[key] as FeatureFlag;
      flag.resetToDefault();
    }

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
