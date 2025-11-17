import type {
  V1ResourceMeta,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";

/**
 * Create a string ID from resource metadata in the format "kind:name".
 * This is a convenience wrapper around resourceNameToId for V1ResourceMeta objects.
 *
 * @param meta - Resource metadata containing name information
 * @returns String ID in format "kind:name", or undefined if metadata is incomplete
 *
 * @example
 * createResourceId({ name: { kind: 'rill.runtime.v1.Model', name: 'orders' } })
 * // Returns: "rill.runtime.v1.Model:orders"
 */
export function createResourceId(meta?: V1ResourceMeta): string | undefined {
  return resourceNameToId(meta?.name);
}

/**
 * Parse a resource ID string into its kind and name components.
 *
 * @param id - Resource ID string in format "kind:name"
 * @returns Object with kind and name, or null if parsing fails
 *
 * @example
 * parseResourceId("rill.runtime.v1.Model:orders")
 * // Returns: { kind: "rill.runtime.v1.Model", name: "orders" }
 */
export function parseResourceId(id: string): V1ResourceName | null {
  const idx = id.indexOf(":");
  if (idx <= 0) return null;

  const kind = id.slice(0, idx);
  const name = id.slice(idx + 1);

  // Validate both parts are non-empty
  if (!kind || !name) return null;

  return {
    kind,
    name,
  };
}

/**
 * Create a resource ID string from a V1ResourceName object.
 * This is the core utility for generating resource identifiers.
 *
 * @param resourceName - Resource name object (can be null/undefined)
 * @returns String ID in format "kind:name", or undefined if input is invalid
 *
 * @example
 * resourceNameToId({ kind: 'rill.runtime.v1.Model', name: 'orders' })
 * // Returns: "rill.runtime.v1.Model:orders"
 *
 * resourceNameToId(null)
 * // Returns: undefined
 */
export function resourceNameToId(
  resourceName?: V1ResourceName | null,
): string | undefined {
  if (!resourceName?.kind || !resourceName?.name) return undefined;
  return `${resourceName.kind}:${resourceName.name}`;
}
