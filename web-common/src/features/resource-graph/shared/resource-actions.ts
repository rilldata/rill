import type { V1Resource } from "@rilldata/web-common/runtime-client";

/**
 * Extract the kind-specific spec from a resource for display.
 * Falls back to all non-meta fields if no known kind key matches.
 */
export function getResourceSpec(res: V1Resource | undefined): string {
  if (!res) return "";
  const kindKeys = [
    "source",
    "model",
    "metricsView",
    "explore",
    "theme",
    "component",
    "canvas",
    "api",
    "connector",
    "report",
    "alert",
  ] as const;
  for (const key of kindKeys) {
    if (res[key]) {
      return JSON.stringify(res[key], null, 2);
    }
  }
  const rest = Object.fromEntries(
    Object.entries(res).filter(([k]) => k !== "meta"),
  );
  return JSON.stringify(rest, null, 2);
}
