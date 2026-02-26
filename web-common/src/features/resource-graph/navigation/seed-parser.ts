import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { resourceNameToId } from "@rilldata/web-common/features/entity-management/resource-utils";
import { ResourceShortNameToResourceKind } from "@rilldata/web-common/features/entity-management/entity-mappers";
import type {
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";

/**
 * URL parameter names for graph navigation.
 * - `kind`: Filter by resource kind (e.g., "metrics", "models", "sources", "dashboards")
 * - `resource`: Show graph for a specific resource by name (e.g., "orders", "revenue")
 *
 * These parameters are mutually exclusive:
 * - `/graph?kind=metrics` - Shows all MetricsView graphs
 * - `/graph?resource=orders` - Shows graph for resource named "orders"
 *
 * The `resource` parameter defaults to MetricsView kind if no kind is specified.
 * To specify a different kind, use the format: `/graph?resource=model:orders`
 */
export const URL_PARAMS = {
  KIND: "kind",
  RESOURCE: "resource",
  EXPANDED: "expanded",
} as const;

/**
 * Valid kind tokens that can be used in the `kind` URL parameter.
 */
export type KindToken =
  | "connector"
  | "metrics"
  | "sources"
  | "models"
  | "dashboards";

/**
 * Normalize plural forms to singular forms for graph seed parsing.
 * The graph feature accepts both singular and plural forms for user convenience,
 * but internally normalizes to the singular forms used by the rest of the app.
 */
function normalizePluralToSingular(kind: string): string {
  const normalized = kind.toLowerCase().trim();

  // Map plural forms to singular forms
  const pluralToSingular: Record<string, string> = {
    connector: "connector",
    connectors: "connector",
    metrics: "metricsview",
    models: "model",
    sources: "source",
    dashboards: "explore", // Maps to explore for token resolution, but expandSeedsByKind handles both Explore and Canvas
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
  ResourceKind.Connector,
  ResourceKind.Source,
  ResourceKind.Model,
  ResourceKind.MetricsView,
  ResourceKind.Explore,
  ResourceKind.Canvas,
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
 * tokenForKind(ResourceKind.Source) // "sources"
 * tokenForKind("rill.runtime.v1.Source") // "sources"
 */
export function tokenForKind(
  kind?: ResourceKind | string | null,
): KindToken | null {
  if (!kind) return null;
  const key = `${kind}`.toLowerCase();
  if (key.includes("connector")) return "connector";
  if (key.includes("source")) return "sources";
  if (key.includes("model")) return "models";
  if (key.includes("metricsview") || key.includes("metric")) return "metrics";
  if (
    key.includes("explore") ||
    key.includes("dashboard") ||
    key.includes("canvas")
  )
    return "dashboards";
  return null;
}

/**
 * Convert a seed string to its display token.
 * Parses the kind from the seed and converts it to a token.
 *
 * @param seed - Seed string (e.g., "model:orders", "source:raw_data", "metrics", "orders")
 * @returns Token string or null
 *
 * @example
 * tokenForSeedString("model:orders") // "models"
 * tokenForSeedString("source:raw_data") // "sources"
 * tokenForSeedString("sources") // "sources"
 * tokenForSeedString("metrics") // "metrics"
 * tokenForSeedString("orders") // "metrics" (defaults to metrics)
 */
export function tokenForSeedString(seed?: string | null): KindToken | null {
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
    // Handle "dashboard:" or "canvas:" prefix - treat as "dashboards"
    if (
      kindPart === "dashboard" ||
      kindPart === "dashboards" ||
      kindPart === "canvas"
    ) {
      return "dashboards";
    }
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
 * 3. Kind tokens ("metrics", "models", "sources", "dashboards") - expands to all visible resources of that kind
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
 * expandSeedsByKind(["models"], resources, coerceKind)
 * // Returns one seed per Model resource
 *
 * expandSeedsByKind(["sources"], resources, coerceKind)
 * // Returns one seed per Source resource
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

  // Filter to visible resources only (to align with graph rendering).
  // Allow connectors even if hidden; GraphContainer pre-filters to OLAP only.
  const visible = resources.filter((r) => {
    const kind = coerceKindFn(r);
    if (!kind || !ALLOWED_FOR_GRAPH.has(kind)) return false;
    if (r.meta?.hidden && kind !== ResourceKind.Connector) return false;
    return true;
  });

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
    // Special case: "dashboards" includes both Explore and Canvas
    const isDashboardsToken =
      raw.toLowerCase() === "dashboards" || raw.toLowerCase() === "dashboard";
    for (const r of visible) {
      const resourceKind = coerceKindFn(r);
      if (isDashboardsToken) {
        // Include both Explore and Canvas for dashboards token
        if (
          resourceKind !== ResourceKind.Explore &&
          resourceKind !== ResourceKind.Canvas
        )
          continue;
      } else {
        // Normal kind matching
        if (resourceKind !== kindToken) continue;
      }
      const name = r.meta?.name?.name;
      const kind = r.meta?.name?.kind; // use actual runtime kind for matching ids
      if (!name || !kind) continue;
      pushSeed({ kind, name });
    }
  }

  return expanded;
}

/**
 * Parsed graph URL parameters.
 */
export interface GraphUrlParams {
  /**
   * Kind filter (e.g., "metrics", "sources", "models", "dashboards").
   * When set, shows all graphs of this resource kind.
   */
  kind: KindToken | null;

  /**
   * Specific resources to show graphs for.
   * Format: "name" (defaults to MetricsView) or "kind:name".
   */
  resources: string[];

  /**
   * Currently expanded graph ID.
   */
  expanded: string | null;
}

/**
 * Parse graph URL parameters from a URL.
 *
 * Handles the new URL API:
 * - `/graph?kind=metrics` - All MetricsView graphs
 * - `/graph?resource=orders` - Specific resource (defaults to MetricsView)
 * - `/graph?resource=model:orders` - Specific resource with explicit kind
 * - `/graph?resource=orders&resource=revenue` - Multiple resources
 *
 * @param url - URL or URLSearchParams to parse
 * @returns Parsed parameters
 *
 * @example
 * parseGraphUrlParams(new URL('/graph?kind=metrics'))
 * // { kind: 'metrics', resources: [], expanded: null }
 *
 * parseGraphUrlParams(new URL('/graph?resource=orders'))
 * // { kind: null, resources: ['orders'], expanded: null }
 */
export function parseGraphUrlParams(
  url: URL | URLSearchParams,
): GraphUrlParams {
  const params = url instanceof URL ? url.searchParams : url;

  // Parse kind parameter
  const kindParam = params.get(URL_PARAMS.KIND)?.trim().toLowerCase() || null;
  // Normalize singular "source" to plural "sources"
  const normalizedKindParam = kindParam === "source" ? "sources" : kindParam;

  const validKind = normalizedKindParam as KindToken | null;
  const kind =
    validKind &&
    ["connector", "metrics", "sources", "models", "dashboards"].includes(
      validKind,
    )
      ? validKind
      : null;

  // Parse resource parameters (supports multiple)
  const resources = params
    .getAll(URL_PARAMS.RESOURCE)
    .map((r) => r.trim())
    .filter((r) => r.length > 0);

  // Parse expanded parameter
  const expanded = params.get(URL_PARAMS.EXPANDED)?.trim() || null;

  return { kind, resources, expanded };
}

/**
 * Convert parsed URL params to internal seed format.
 *
 * This bridges the new URL API to the existing internal processing.
 * The `kind` parameter becomes a kind token (e.g., "metrics"),
 * and `resource` parameters become explicit seeds (e.g., "model:orders").
 *
 * @param params - Parsed URL parameters
 * @returns Array of seed strings for internal processing
 *
 * @example
 * urlParamsToSeeds({ kind: 'metrics', resources: [], expanded: null })
 * // ['metrics']
 *
 * urlParamsToSeeds({ kind: null, resources: ['orders'], expanded: null })
 * // ['orders']
 *
 * urlParamsToSeeds({ kind: null, resources: ['model:orders'], expanded: null })
 * // ['model:orders']
 */
export function urlParamsToSeeds(params: GraphUrlParams): string[] {
  // If kind is specified, use it as the seed (expands to all resources of that kind)
  if (params.kind) {
    return [params.kind];
  }

  // Otherwise, use the resource parameters as seeds
  return params.resources;
}

/**
 * Build a graph URL with the new API.
 *
 * @param options - URL building options
 * @returns URL path with query string
 *
 * @example
 * buildGraphUrlNew({ kind: 'metrics' })
 * // '/graph?kind=metrics'
 *
 * buildGraphUrlNew({ resources: ['orders', 'revenue'] })
 * // '/graph?resource=orders&resource=revenue'
 *
 * buildGraphUrlNew({ resources: ['model:orders'], expanded: 'rill.runtime.v1.Model:orders' })
 * // '/graph?resource=model:orders&expanded=rill.runtime.v1.Model:orders'
 */
export function buildGraphUrlNew(options: {
  kind?: KindToken | null;
  resources?: string[];
  expanded?: string | null;
  basePath?: string;
}): string {
  const { kind, resources = [], expanded, basePath = "/graph" } = options;

  const params = new URLSearchParams();

  if (kind) {
    params.set(URL_PARAMS.KIND, kind);
  } else {
    for (const resource of resources) {
      if (resource && resource.trim()) {
        params.append(URL_PARAMS.RESOURCE, resource.trim());
      }
    }
  }

  if (expanded) {
    params.set(URL_PARAMS.EXPANDED, expanded);
  }

  const queryString = params.toString();
  return queryString ? `${basePath}?${queryString}` : basePath;
}
