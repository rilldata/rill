import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { getResourceStatus } from "@rilldata/web-common/features/resource-graph/shared/resource-status";

export type { ResourceStatusFilterValue as ResourceStatus } from "@rilldata/web-common/features/resource-graph/shared/types";
export { getResourceStatus };

/**
 * Filters resources by kind, search text, and status.
 * All filters are AND-ed together. Empty filter arrays match all.
 */
export function filterResources(
  resources: V1Resource[] | undefined,
  types: string[],
  search: string,
  statuses: string[],
): V1Resource[] {
  if (!resources) return [];

  return resources.filter((r) => {
    const kind = r.meta?.name?.kind;
    const name = r.meta?.name?.name ?? "";

    const matchesType = types.length === 0 || types.includes(kind ?? "");
    const matchesSearch =
      !search || name.toLowerCase().includes(search.toLowerCase());
    const matchesStatus =
      statuses.length === 0 || statuses.includes(getResourceStatus(r));

    return matchesType && matchesSearch && matchesStatus;
  });
}
