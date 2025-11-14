import type {
  V1ResourceMeta,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";

/**
 * Create a string ID from resource metadata in the format "kind:name".
 * This is the standard format for identifying resources across the application.
 *
 * @param meta - Resource metadata containing name information
 * @returns String ID in format "kind:name", or undefined if metadata is incomplete
 *
 * @example
 * createResourceId({ name: { kind: 'rill.runtime.v1.Model', name: 'orders' } })
 * // Returns: "rill.runtime.v1.Model:orders"
 */
export function createResourceId(meta?: V1ResourceMeta): string | undefined {
  if (!meta?.name?.name || !meta?.name?.kind) return undefined;
  return `${meta.name.kind}:${meta.name.name}`;
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
  return {
    kind: id.slice(0, idx),
    name: id.slice(idx + 1),
  };
}

/**
 * Create a resource ID string from a V1ResourceName object.
 *
 * @param resourceName - Resource name object
 * @returns String ID in format "kind:name"
 *
 * @example
 * resourceNameToId({ kind: 'rill.runtime.v1.Model', name: 'orders' })
 * // Returns: "rill.runtime.v1.Model:orders"
 */
export function resourceNameToId(resourceName: V1ResourceName): string {
  return `${resourceName.kind}:${resourceName.name}`;
}
