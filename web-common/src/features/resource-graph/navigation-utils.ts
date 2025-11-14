import { goto } from "$app/navigation";

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
  const seeds = [`${kind}:${name}`, ...(additionalSeeds || [])];
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
    .map(({ kind, name }) => `seed=${encodeURIComponent(`${kind}:${name}`)}`)
    .join("&");
  return `/graph?${seedParams}`;
}
