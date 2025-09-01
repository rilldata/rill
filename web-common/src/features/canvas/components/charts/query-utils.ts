import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export function getFilterWithNullHandling(
  where: V1Expression | undefined,
  fieldConfig: FieldConfig | undefined,
): V1Expression | undefined {
  if (!fieldConfig || fieldConfig.showNull || fieldConfig.type !== "nominal") {
    return where;
  }

  const excludeNullFilter = createInExpression(fieldConfig.field, [null], true);
  return mergeFilters(where, excludeNullFilter);
}
