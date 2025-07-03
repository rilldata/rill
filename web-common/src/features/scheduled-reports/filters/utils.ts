import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import type {
  V1Expression,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export type FiltersFormValues = {
  whereFilter: V1Expression;
  dimensionsWithInlistFilter: string[];
  dimensionThresholdFilters: Array<DimensionThresholdFilter>;
  timeRange: V1TimeRange;
  comparisonTimeRange: V1TimeRange | undefined;
};
