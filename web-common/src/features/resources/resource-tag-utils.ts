// Virtual tag value used for dashboards that have no tags. Used as the
// ?tags= URL param value, the map key for grouping, and the comparison
// constant for "show untagged" logic.
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  buildTagIndex,
  type DimensionTag,
} from "@rilldata/web-common/components/menu/tag-utils.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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

export function getTagFilterLabel(tags: string[]) {
  return tags.length === 0
    ? m.status_all_tags()
    : tags.length === 1
      ? tags[0]
      : m.status_levels_selected({ first: tags[0], count: tags.length - 1 });
}
