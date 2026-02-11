import { goto } from "$app/navigation";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { buildGraphUrlNew, type KindToken } from "./seed-parser";

/**
 * Navigate to the resource graph view for a specific resource.
 *
 * Uses the new URL API:
 * - `/graph?resource=model:orders` for specific resources
 *
 * @param kind - Resource kind (source, model, metrics, etc.)
 * @param name - Resource name
 * @param additionalResources - Optional additional resources to include in the URL
 */
export function navigateToResourceGraph(
  kind: string,
  name: string,
  additionalResources?: string[],
): void {
  // Build resource identifier: "shortKind:name" format
  const shortKind = getShortKindName(kind);
  const resourceId = shortKind ? `${shortKind}:${name}` : name;

  const resources = [resourceId, ...(additionalResources || [])];
  const url = buildGraphUrlNew({ resources });
  goto(url);
}

/**
 * Navigate to the resource graph view filtered by kind.
 *
 * @param kindToken - Kind token (metrics, models, sources, dashboards)
 */
export function navigateToResourceGraphByKind(kindToken: KindToken): void {
  const url = buildGraphUrlNew({ kind: kindToken });
  goto(url);
}

/**
 * Convert a fully qualified kind to its short name.
 * @param kind - Fully qualified kind (e.g., "rill.runtime.v1.Model")
 * @returns Short kind name (e.g., "model") or null if unknown
 */
function getShortKindName(kind: string): string | null {
  const lower = kind.toLowerCase();
  if (lower.includes("source")) return "source";
  if (lower.includes("model")) return "model";
  if (lower.includes("metricsview")) return "metrics";
  if (lower.includes("explore")) return "dashboard";
  if (lower.includes("canvas")) return "canvas";
  return null;
}

/**
 * Create a reusable graph navigation handler with error handling.
 * This utility reduces code duplication in menu items and action buttons.
 *
 * @param componentName - Name of the component for logging (e.g., "ModelMenuItems")
 * @param kind - Resource kind to navigate to (e.g., "model", "metrics")
 * @param getResource - Function that returns the current resource (can be V1Resource or V1GetResourceResponse)
 * @returns A function that navigates to the resource graph with error handling
 *
 * @example
 * // With direct resource
 * const viewGraph = createGraphNavigationHandler(
 *   "ModelMenuItems",
 *   "model",
 *   () => $modelQuery.data
 * );
 *
 * @example
 * // With response containing resource property
 * const viewGraph = createGraphNavigationHandler(
 *   "MetricsViewMenuItems",
 *   "metrics",
 *   () => $resourceQuery.data?.resource
 * );
 */
export function createGraphNavigationHandler(
  componentName: string,
  kind: string,
  getResource: () => V1Resource | undefined,
): () => void {
  return () => {
    try {
      const resource = getResource();
      const name = resource?.meta?.name?.name;
      if (!name) {
        console.warn(
          `[${componentName}] Cannot navigate to graph: resource name is missing`,
        );
        return;
      }
      navigateToResourceGraph(kind, name);
    } catch (error) {
      console.error(`[${componentName}] Failed to navigate to graph:`, error);
      // TODO: Show toast notification to user when toast system is available
    }
  };
}

/**
 * Build a URL to the resource graph view with multiple resources.
 *
 * Uses the new URL API:
 * - `/graph?resource=model:orders&resource=source:raw_data`
 *
 * @param seeds - Array of seed objects with kind and name
 * @returns The constructed graph URL
 */
export function buildGraphUrl(
  seeds: Array<{ kind: string; name: string }>,
): string {
  const resources = seeds
    .map(({ kind, name }) => {
      const shortKind = getShortKindName(kind);
      return shortKind ? `${shortKind}:${name}` : name;
    })
    .filter((r) => r);

  return buildGraphUrlNew({ resources });
}
