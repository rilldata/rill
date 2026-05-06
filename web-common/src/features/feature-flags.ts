import { useQueryClient } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
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
