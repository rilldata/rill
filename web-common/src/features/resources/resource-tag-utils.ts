// Virtual tag value used for dashboards that have no tags. Used as the
// ?tags= URL param value, the map key for grouping, and the comparison
// constant for "show untagged" logic.
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  buildTagIndex,
  type DimensionTag,
} from "@rilldata/web-common/components/menu/tag-utils.ts";

export const UNTAGGED_KEY = "__rill_untagged";
// Display label rendered to the user wherever `UNTAGGED_KEY` would otherwise
// surface (folder header, breadcrumb dropdown). The URL value stays
// kebab-case for stable links.
export const UNTAGGED_LABEL = "Not Tagged";

export function getResourceTags(resource: V1Resource): string[] {
  return resource.meta?.tags ?? [];
}

export function getPrimaryTag(resource: V1Resource): string {
  return getResourceTags(resource)[0] ?? UNTAGGED_KEY;
}

export function getAllTagsForResources(
  resources: V1Resource[],
): DimensionTag[] {
  return buildTagIndex(
    resources.map((r) => ({
      name: r.meta?.name?.name,
      tags: r.meta?.tags,
    })),
  ).tags.sort((a, b) => a.name.localeCompare(b.name));
}
