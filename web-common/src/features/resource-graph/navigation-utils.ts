import { goto } from "$app/navigation";
import { resourceNameToId } from "@rilldata/web-common/features/entity-management/resource-utils";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

/**
 * Navigate to the resource graph view with a seed for a specific resource.
 * @param kind - Resource kind (source, model, metrics, etc.)
 * @param name - Resource name
 * @param additionalSeeds - Optional additional seeds to include in the URL
 */
export function navigateToResourceGraph(
  kind: string,
  name: string,
  additionalSeeds?: string[],
): void {
  const seedId = resourceNameToId({ kind, name });
  if (!seedId) return; // Early return if invalid kind/name
  const seeds = [seedId, ...(additionalSeeds || [])];
  const seedParams = seeds
    .map((s) => `seed=${encodeURIComponent(s)}`)
    .join("&");
  goto(`/graph?${seedParams}`);
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
 * Build a URL to the resource graph view with multiple seeds.
 * @param seeds - Array of seed objects with kind and name
 * @returns The constructed graph URL
 */
export function buildGraphUrl(
  seeds: Array<{ kind: string; name: string }>,
): string {
  const seedParams = seeds
    .map(({ kind, name }) => {
      const id = resourceNameToId({ kind, name });
      return id ? `seed=${encodeURIComponent(id)}` : "";
    })
    .filter((s) => s) // Remove empty strings from invalid seeds
    .join("&");
  return `/graph?${seedParams}`;
}
