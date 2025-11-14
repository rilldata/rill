import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource, V1ResourceName } from "@rilldata/web-common/runtime-client";

/**
 * Mapping of common kind aliases to their full ResourceKind values.
 * Used for parsing seed strings with short-form kind names.
 */
export const KIND_ALIASES: Record<string, ResourceKind> = {
  metrics: ResourceKind.MetricsView,
  metric: ResourceKind.MetricsView,
  metricsview: ResourceKind.MetricsView,
  dashboard: ResourceKind.Explore,
  explore: ResourceKind.Explore,
  model: ResourceKind.Model,
  source: ResourceKind.Source,
  canvas: ResourceKind.Canvas,
};

/**
 * Resource kinds that are allowed in the graph visualization.
 */
export const ALLOWED_FOR_GRAPH = new Set<ResourceKind>([
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
]);

/**
 * Normalize a seed string into a standard format.
 * Handles various input formats:
 * - "name" -> defaults to MetricsView kind
 * - "kind:name" -> parses kind and name
 * - "rill.runtime.v1.Kind:name" -> fully qualified kind
 *
 * @param s - Seed string to normalize
 * @returns Normalized seed as either a string or V1ResourceName object
 *
 * @example
 * normalizeSeed("orders") // { kind: "rill.runtime.v1.MetricsView", name: "orders" }
 * normalizeSeed("model:clean_orders") // { kind: "rill.runtime.v1.Model", name: "clean_orders" }
 */
export function normalizeSeed(s: string): string | V1ResourceName {
  const idx = s.indexOf(":");
  if (idx === -1) {
    // No colon: treat as a metrics view name
    return { kind: ResourceKind.MetricsView, name: s };
  }
  const kindPart = s.slice(0, idx);
  const namePart = s.slice(idx + 1);

  // Check if it's a fully qualified kind (contains dots)
  if (kindPart.includes(".")) {
    return { kind: kindPart, name: namePart };
  }

  // Map alias to full kind
  const mapped = KIND_ALIASES[kindPart.trim().toLowerCase()];
  if (mapped) return { kind: mapped, name: namePart };

  // Return as-is if no mapping found
  return s;
}

/**
 * Check if a string is a kind token (plural or singular kind name).
 * Kind tokens are used to expand to all resources of that kind.
 *
 * @param s - String to check
 * @returns The ResourceKind if it's a kind token, undefined otherwise
 *
 * @example
 * isKindToken("metrics") // ResourceKind.MetricsView
 * isKindToken("sources") // ResourceKind.Source
 * isKindToken("orders") // undefined (not a kind token)
 */
export function isKindToken(s: string): ResourceKind | undefined {
  const key = s.trim().toLowerCase();
  switch (key) {
    case "metrics":
    case "metric":
    case "metricsview":
      return ResourceKind.MetricsView;
    case "dashboards":
    case "dashboard":
    case "explore":
    case "explores":
      return ResourceKind.Explore;
    case "models":
    case "model":
      return ResourceKind.Model;
    case "sources":
    case "source":
      return ResourceKind.Source;
    default:
      return undefined;
  }
}

/**
 * Expand seed strings by kind tokens into individual resource seeds.
 * Handles three input formats:
 * 1. Explicit seeds ("kind:name") - kept as-is
 * 2. Name-only seeds ("name") - defaults to MetricsView
 * 3. Kind tokens ("metrics", "sources") - expands to all visible resources of that kind
 *
 * @param seedStrings - Array of seed strings to expand
 * @param resources - All resources to consider for expansion
 * @param coerceKindFn - Function to coerce resource kinds (e.g., Model -> Source for defined-as-source)
 * @returns Array of normalized seeds ready for graph partitioning
 *
 * @example
 * expandSeedsByKind(["metrics"], resources, coerceKind)
 * // Returns one seed per MetricsView resource
 *
 * expandSeedsByKind(["model:orders", "sources"], resources, coerceKind)
 * // Returns the orders model plus one seed per Source resource
 */
export function expandSeedsByKind(
  seedStrings: string[] | undefined,
  resources: V1Resource[],
  coerceKindFn: (res: V1Resource) => ResourceKind | undefined,
): (string | V1ResourceName)[] {
  const input = seedStrings ?? [];
  const expanded: (string | V1ResourceName)[] = [];
  const seen = new Set<string>(); // de-dupe by id "kind:name"

  // Helper to push a normalized seed and avoid duplicates
  const pushSeed = (s: string | V1ResourceName) => {
    const id = typeof s === "string" ? s : `${s.kind}:${s.name}`;
    if (seen.has(id)) return;
    seen.add(id);
    expanded.push(s);
  };

  // Filter to visible resources only (to align with graph rendering)
  const visible = resources.filter(
    (r) => ALLOWED_FOR_GRAPH.has(coerceKindFn(r) as ResourceKind) && !r.meta?.hidden,
  );

  for (const raw of input) {
    if (!raw) continue;

    // Explicit seed with colon
    if (raw.includes(":")) {
      pushSeed(normalizeSeed(raw));
      continue;
    }

    // Check if it's a kind token
    const kindToken = isKindToken(raw);
    if (!kindToken) {
      // Name-only, defaults to metrics view name
      pushSeed(normalizeSeed(raw));
      continue;
    }

    // Expand: one seed per visible resource of this kind
    for (const r of visible) {
      if (coerceKindFn(r) !== kindToken) continue;
      const name = r.meta?.name?.name;
      const kind = r.meta?.name?.kind; // use actual runtime kind for matching ids
      if (!name || !kind) continue;
      pushSeed({ kind, name });
    }
  }

  return expanded;
}
