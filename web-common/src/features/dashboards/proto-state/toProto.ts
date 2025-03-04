import {
  NullValue,
  type PartialMessage,
  protoBase64,
  Timestamp,
  Value,
} from "@bufbuild/protobuf";
import { mapMeasureFilterToExpr } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  ToProtoOperationMap,
  ToProtoPivotTableModeMap,
  ToProtoTimeGrainMap,
} from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  Condition,
  Expression,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import {
  DashboardDimensionFilter,
  DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardContextColumn,
  DashboardTimeRange,
  PivotElement,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { V1Operation, V1TimeGrain } from "@rilldata/web-common/runtime-client";

// TODO: make a follow up PR to use the one from the proto directly
const LeaderboardContextColumnMap: Record<
  LeaderboardContextColumn,
  DashboardState_LeaderboardContextColumn
> = {
  [LeaderboardContextColumn.PERCENT]:
    DashboardState_LeaderboardContextColumn.PERCENT,
  [LeaderboardContextColumn.DELTA_PERCENT]:
    DashboardState_LeaderboardContextColumn.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]:
    DashboardState_LeaderboardContextColumn.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.HIDDEN]:
    DashboardState_LeaderboardContextColumn.HIDDEN,
};

const TDDChartTypeMap: Record<TDDChart, string> = {
  [TDDChart.DEFAULT]: "default",
  [TDDChart.STACKED_BAR]: "stacked_bar",
  [TDDChart.GROUPED_BAR]: "grouped_bar",
  [TDDChart.STACKED_AREA]: "stacked_area",
};

export function getProtoFromDashboardState(
  metrics: MetricsExplorerEntity,
): string {
  if (!metrics) return "";

  const state: PartialMessage<DashboardState> = {};
  if (metrics.whereFilter) {
    state.where = toExpressionProto(metrics.whereFilter);
  }
  if (metrics.dimensionThresholdFilters?.length) {
    state.having = metrics.dimensionThresholdFilters.map(
      ({ name, filters }) =>
        new DashboardDimensionFilter({
          name,
          filter: toExpressionProto(
            createAndExpression(
              filters
                .map(mapMeasureFilterToExpr)
                .filter(Boolean) as V1Expression[],
            ),
          ),
        }),
    );
  }
  if (metrics.selectedTimeRange) {
    state.timeRange = toTimeRangeProto(metrics.selectedTimeRange);
    if (metrics.selectedTimeRange.interval) {
      state.timeGrain =
        ToProtoTimeGrainMap[metrics.selectedTimeRange.interval] ??
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    }
  }
  if (metrics.selectedComparisonTimeRange) {
    state.compareTimeRange = toTimeRangeProto(
      metrics.selectedComparisonTimeRange,
    );
  }
  if (metrics.lastDefinedScrubRange) {
    state.scrubRange = toScrubProto(metrics.lastDefinedScrubRange);
  }
  state.showTimeComparison = Boolean(metrics.showTimeComparison);
  if (metrics.selectedComparisonDimension) {
    state.comparisonDimension = metrics.selectedComparisonDimension;
  }

  state.selectedTimezone = metrics.selectedTimezone;

  if (metrics.leaderboardMeasureName) {
    state.leaderboardMeasure = metrics.leaderboardMeasureName;
  }

  if (metrics.tdd?.pinIndex !== undefined) {
    state.pinIndex = metrics.tdd.pinIndex;
  }
  if (metrics.tdd?.chartType !== undefined) {
    state.chartType = TDDChartTypeMap[metrics.tdd.chartType];
  }

  if (metrics.allMeasuresVisible) {
    state.allMeasuresVisible = true;
  } else if (metrics.visibleMeasureKeys) {
    state.visibleMeasures = [...metrics.visibleMeasureKeys];
  }

  if (metrics.allDimensionsVisible) {
    state.allDimensionsVisible = true;
  } else if (metrics.visibleDimensionKeys) {
    state.visibleDimensions = [...metrics.visibleDimensionKeys];
  }

  if (metrics.leaderboardContextColumn) {
    state.leaderboardContextColumn =
      LeaderboardContextColumnMap[metrics.leaderboardContextColumn];
  }

  if (metrics.sortDirection) {
    state.leaderboardSortDirection = metrics.sortDirection;
  }
  if (metrics.dashboardSortType) {
    state.leaderboardSortType = metrics.dashboardSortType;
  }

  if (metrics.pivot) {
    Object.assign(state, toPivotProto(metrics.pivot));
  }

  Object.assign(state, toActivePageProto(metrics));

  const message = new DashboardState(state);
  return protoToBase64(message.toBinary());
}

function protoToBase64(proto: Uint8Array) {
  return protoBase64.enc(proto);
}

function toTimeRangeProto(range: DashboardTimeControls) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: range.name,
  };
  if (
    range.name === TimeRangePreset.CUSTOM ||
    range.name === TimeComparisonOption.CUSTOM
  ) {
    if (range.start) timeRangeArgs.timeStart = toTimeProto(range.start);
    if (range.end) timeRangeArgs.timeEnd = toTimeProto(range.end);
  }
  return new DashboardTimeRange(timeRangeArgs);
}

function toScrubProto(range: ScrubRange) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: TimeRangePreset.CUSTOM,
  };
  timeRangeArgs.timeStart = toTimeProto(range.start);
  timeRangeArgs.timeEnd = toTimeProto(range.end);

  return new DashboardTimeRange(timeRangeArgs);
}

function toTimeProto(date: Date) {
  return new Timestamp({
    seconds: BigInt(date.getTime()),
  });
}

function toExpressionProto(expression: V1Expression): Expression {
  if ("ident" in expression) {
    return new Expression({
      expression: {
        case: "ident",
        value: expression.ident as string,
      },
    });
  }
  if ("val" in expression) {
    return new Expression({
      expression: {
        case: "val",
        value: toPbValue(expression.val),
      },
    });
  }
  if (expression.cond) {
    return new Expression({
      expression: {
        case: "cond",
        value: new Condition({
          op: ToProtoOperationMap[
            expression.cond.op ?? V1Operation.OPERATION_UNSPECIFIED
          ],
          exprs: expression.cond.exprs?.map((e) => toExpressionProto(e)) ?? [],
        }),
      },
    });
  }
  return new Expression();
}

function toPbValue(val: unknown) {
  if (val === null) {
    return new Value({
      kind: {
        case: "nullValue",
        value: NullValue.NULL_VALUE,
      },
    });
  }
  switch (typeof val) {
    case "string":
      return new Value({
        kind: {
          case: "stringValue",
          value: val,
        },
      });
    case "number":
      return new Value({
        kind: {
          case: "numberValue",
          value: val,
        },
      });
    case "boolean":
      return new Value({
        kind: {
          case: "boolValue",
          value: val,
        },
      });
    // TODO: other options are not currently in a filter. but we might need them in future
    default:
      // force as string for unknown types. this is the older behaviour
      return new Value({
        kind: {
          case: "stringValue",
          value: JSON.stringify(val),
        },
      });
  }
}

const mapPivotDimensions: (
  dimension: PivotChipData,
) => PartialMessage<PivotElement> = (dimension: PivotChipData) => {
  if (dimension.type === PivotChipType.Dimension) {
    return {
      element: {
        case: "pivotDimension",
        value: dimension.id,
      },
    };
  } else if (dimension.type === PivotChipType.Time) {
    return {
      element: {
        case: "pivotTimeDimension",
        value: ToProtoTimeGrainMap[dimension.id as V1TimeGrain],
      },
    };
  } else {
    throw new Error("Unsupported pivot dimension type");
  }
};

function toPivotProto(pivotState: PivotState): PartialMessage<DashboardState> {
  return {
    pivotIsActive: pivotState.active,
    pivotRowAllDimensions: pivotState.rows.dimension.map(mapPivotDimensions),
    pivotColumnAllDimensions:
      pivotState.columns.dimension.map(mapPivotDimensions),

    pivotColumnMeasures: pivotState.columns.measure.map((m) => m.id),

    // pivotExpanded: pivotState.expanded,
    pivotSort: pivotState.sorting,
    pivotColumnPage: pivotState.columnPage,
    pivotEnableComparison: pivotState.enableComparison,
    pivotTableMode: ToProtoPivotTableModeMap[pivotState.tableMode],
  };
}

function toActivePageProto(
  metrics: MetricsExplorerEntity,
): PartialMessage<DashboardState> {
  switch (metrics.activePage) {
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.PIVOT:
      return {
        activePage: metrics.activePage,
      };

    case DashboardState_ActivePage.DIMENSION_TABLE:
      return {
        activePage: metrics.activePage,
        selectedDimension: metrics.selectedDimensionName,
      };

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      return {
        activePage: metrics.activePage,
        expandedMeasure: metrics.tdd.expandedMeasureName,
      };
  }

  return {};
}
