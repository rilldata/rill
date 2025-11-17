import { goto } from "$app/navigation";
import { resourceNameToId } from "@rilldata/web-common/features/entity-management/resource-utils";

/**
 * Navigate to the resource graph view with a seed for a specific resource.
 * @param kind - Resource kind (source, model, metrics, etc.)
 * @param name - Resource name
 * @param additionalSeeds - Optional additional seeds to include in the URL
 */
export function navigateToResourceGraph(
  kind: string,
  name: string,
  additionalSeeds?: string[]
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
 * Build a URL to the resource graph view with multiple seeds.
 * @param seeds - Array of seed objects with kind and name
 * @returns The constructed graph URL
 */
export function buildGraphUrl(
  seeds: Array<{ kind: string; name: string }>
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
