import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { resourceNameToId } from "@rilldata/web-common/features/entity-management/resource-utils";
import { ResourceShortNameToResourceKind } from "@rilldata/web-common/features/entity-management/entity-mappers";
import type {
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";

/**
 * Normalize plural forms to singular forms for graph seed parsing.
 * The graph feature accepts both singular and plural forms for user convenience,
 * but internally normalizes to the singular forms used by the rest of the app.
 */
function normalizePluralToSingular(kind: string): string {
  const normalized = kind.toLowerCase().trim();

  // Map plural forms to singular forms
  const pluralToSingular: Record<string, string> = {
    metrics: "metricsview",
    models: "model",
    sources: "source",
    dashboards: "explore",
    // Also handle "dashboard" -> "explore" for consistency
    dashboard: "explore",
    // Handle "metric" -> "metricsview"
    metric: "metricsview",
  };

  return pluralToSingular[normalized] ?? normalized;
}

/**
 * Resolve a kind alias string to its ResourceKind.
 * Handles both singular and plural forms by normalizing them first,
 * then looking up in the canonical ResourceShortNameToResourceKind mapping.
 *
 * @param kindAlias - Short name or alias for a resource kind (e.g., "models", "sources")
 * @returns The ResourceKind if found, undefined otherwise
 */
function resolveKindAlias(kindAlias: string): ResourceKind | undefined {
  const normalized = normalizePluralToSingular(kindAlias);
  return ResourceShortNameToResourceKind[normalized];
}

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
 * - "kind:name" -> parses kind and name (accepts plural forms)
 * - "rill.runtime.v1.Kind:name" -> fully qualified kind
 *
 * The graph feature accepts plural forms (e.g., "models", "sources") for user convenience,
 * but normalizes them to the singular forms used by the rest of the app.
 *
 * @param s - Seed string to normalize
 * @returns Normalized seed as either a string or V1ResourceName object
 *
 * @example
 * normalizeSeed("orders") // { kind: "rill.runtime.v1.MetricsView", name: "orders" }
 * normalizeSeed("model:clean_orders") // { kind: "rill.runtime.v1.Model", name: "clean_orders" }
 * normalizeSeed("models:clean_orders") // { kind: "rill.runtime.v1.Model", name: "clean_orders" }
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

  // Resolve alias (handles both singular and plural forms)
  const mapped = resolveKindAlias(kindPart);
  if (mapped) return { kind: mapped, name: namePart };

  // Return as-is if no mapping found
  return s;
}

/**
 * Check if a string is a kind token (plural or singular kind name).
 * Kind tokens are used to expand to all resources of that kind.
 *
 * Uses the normalization function to handle plural forms, then looks up
 * in the canonical ResourceShortNameToResourceKind mapping.
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
  // Use the same normalization and resolution logic
  return resolveKindAlias(s);
}

/**
 * Convert a ResourceKind to its display token (plural form).
 * Used for highlighting overview nodes in the summary graph.
 *
 * @param kind - ResourceKind or string kind
 * @returns Token string ("sources", "models", "metrics", "dashboards") or null
 *
 * @example
 * tokenForKind(ResourceKind.Model) // "models"
 * tokenForKind("rill.runtime.v1.Source") // "sources"
 */
export function tokenForKind(
  kind?: ResourceKind | string | null,
): "metrics" | "sources" | "models" | "dashboards" | null {
  if (!kind) return null;
  const key = `${kind}`.toLowerCase();
  if (key.includes("source")) return "sources";
  if (key.includes("model")) return "models";
  if (key.includes("metricsview") || key.includes("metric")) return "metrics";
  if (key.includes("explore") || key.includes("dashboard")) return "dashboards";
  return null;
}

/**
 * Convert a seed string to its display token.
 * Parses the kind from the seed and converts it to a token.
 *
 * @param seed - Seed string (e.g., "model:orders", "metrics", "orders")
 * @returns Token string or null
 *
 * @example
 * tokenForSeedString("model:orders") // "models"
 * tokenForSeedString("metrics") // "metrics"
 * tokenForSeedString("orders") // "metrics" (defaults to metrics)
 */
export function tokenForSeedString(
  seed?: string | null,
): "metrics" | "sources" | "models" | "dashboards" | null {
  if (!seed) return null;
  const normalized = seed.trim().toLowerCase();
  if (!normalized) return null;

  // Check if it's a kind token first
  const kindToken = isKindToken(normalized);
  if (kindToken) return tokenForKind(kindToken);

  // Parse "kind:name" format
  const idx = normalized.indexOf(":");
  if (idx !== -1) {
    const kindPart = normalized.slice(0, idx);
    const mapped = resolveKindAlias(kindPart);
    if (mapped) return tokenForKind(mapped);
    return tokenForKind(kindPart);
  }

  // Name-only defaults to metrics
  return "metrics";
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
    const id = typeof s === "string" ? s : resourceNameToId(s);
    if (!id) return; // Skip if resourceNameToId returns undefined
    if (seen.has(id)) return;
    seen.add(id);
    expanded.push(s);
  };

  // Filter to visible resources only (to align with graph rendering)
  const visible = resources.filter(
    (r) =>
      ALLOWED_FOR_GRAPH.has(coerceKindFn(r) as ResourceKind) && !r.meta?.hidden,
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
