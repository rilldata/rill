import type {
  ComparisonTimeRangeState,
  TimeRangeState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type {
  V1Expression,
  V1TimeGrain,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

export interface TimeAndFilterStore {
  timeRange: V1TimeRange;
  comparisonTimeRange: V1TimeRange | undefined;
  where: V1Expression | undefined;
  timeGrain: V1TimeGrain | undefined;
  showTimeComparison: boolean;
  timeRangeState: TimeRangeState | undefined;
  comparisonTimeRangeState: ComparisonTimeRangeState | undefined;
}
