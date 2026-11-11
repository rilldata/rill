import type {
  PivotTableMode,
  PivotTotalsRowPosition,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { Operation } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState_PivotTableMode,
  DashboardState_PivotTotalsRowPosition,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1Operation, V1TimeGrain } from "@rilldata/web-common/runtime-client";

// This file should contain all the map from proto and API values.
// TODO: we should try and find a way to merge these enums

export const ToProtoOperationMap: Record<V1Operation, Operation> = {
  [V1Operation.OPERATION_UNSPECIFIED]: Operation.UNSPECIFIED,
  [V1Operation.OPERATION_EQ]: Operation.EQ,
  [V1Operation.OPERATION_NEQ]: Operation.NEQ,
  [V1Operation.OPERATION_LT]: Operation.LT,
  [V1Operation.OPERATION_LTE]: Operation.LTE,
  [V1Operation.OPERATION_GT]: Operation.GT,
  [V1Operation.OPERATION_GTE]: Operation.GTE,
  [V1Operation.OPERATION_OR]: Operation.OR,
  [V1Operation.OPERATION_AND]: Operation.AND,
  [V1Operation.OPERATION_IN]: Operation.IN,
  [V1Operation.OPERATION_NIN]: Operation.NIN,
  [V1Operation.OPERATION_LIKE]: Operation.LIKE,
  [V1Operation.OPERATION_NLIKE]: Operation.NLIKE,
  [V1Operation.OPERATION_CAST]: Operation.CAST,
};

export const FromProtoOperationMap = {} as Record<Operation, V1Operation>;
for (const op in ToProtoOperationMap) {
  FromProtoOperationMap[ToProtoOperationMap[op]] = op;
}

export const ToProtoTimeGrainMap: Record<V1TimeGrain, TimeGrain> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: TimeGrain.TIME_GRAIN_UNSPECIFIED,
  [V1TimeGrain.TIME_GRAIN_DAY]: TimeGrain.TIME_GRAIN_DAY,
  [V1TimeGrain.TIME_GRAIN_HOUR]: TimeGrain.TIME_GRAIN_HOUR,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: TimeGrain.TIME_GRAIN_MILLISECOND,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: TimeGrain.TIME_GRAIN_MINUTE,
  [V1TimeGrain.TIME_GRAIN_MONTH]: TimeGrain.TIME_GRAIN_MONTH,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: TimeGrain.TIME_GRAIN_QUARTER,
  [V1TimeGrain.TIME_GRAIN_SECOND]: TimeGrain.TIME_GRAIN_SECOND,
  [V1TimeGrain.TIME_GRAIN_WEEK]: TimeGrain.TIME_GRAIN_WEEK,
  [V1TimeGrain.TIME_GRAIN_YEAR]: TimeGrain.TIME_GRAIN_YEAR,
  [V1TimeGrain.TIME_GRAIN_5_MINUTE]: TimeGrain.TIME_GRAIN_5_MINUTE,
  [V1TimeGrain.TIME_GRAIN_10_MINUTE]: TimeGrain.TIME_GRAIN_10_MINUTE,
  [V1TimeGrain.TIME_GRAIN_15_MINUTE]: TimeGrain.TIME_GRAIN_15_MINUTE,
  [V1TimeGrain.TIME_GRAIN_30_MINUTE]: TimeGrain.TIME_GRAIN_30_MINUTE,
  [V1TimeGrain.TIME_GRAIN_2_HOUR]: TimeGrain.TIME_GRAIN_2_HOUR,
  [V1TimeGrain.TIME_GRAIN_3_HOUR]: TimeGrain.TIME_GRAIN_3_HOUR,
  [V1TimeGrain.TIME_GRAIN_4_HOUR]: TimeGrain.TIME_GRAIN_4_HOUR,
  [V1TimeGrain.TIME_GRAIN_6_HOUR]: TimeGrain.TIME_GRAIN_6_HOUR,
  [V1TimeGrain.TIME_GRAIN_8_HOUR]: TimeGrain.TIME_GRAIN_8_HOUR,
  [V1TimeGrain.TIME_GRAIN_12_HOUR]: TimeGrain.TIME_GRAIN_12_HOUR,
};

export const FromProtoTimeGrainMap = {} as Record<TimeGrain, V1TimeGrain>;
for (const grain in ToProtoTimeGrainMap) {
  FromProtoTimeGrainMap[ToProtoTimeGrainMap[grain]] = grain;
}

export const ToProtoPivotTableModeMap: Record<
  PivotTableMode,
  DashboardState_PivotTableMode
> = {
  flat: DashboardState_PivotTableMode.FLAT,
  nest: DashboardState_PivotTableMode.NEST,
};

export const FromProtoPivotTableModeMap = {} as Record<
  DashboardState_PivotTableMode,
  PivotTableMode
>;
for (const op in ToProtoPivotTableModeMap) {
  FromProtoPivotTableModeMap[ToProtoPivotTableModeMap[op]] = op;
}

export const ToProtoPivotTotalsRowPositionMap: Record<
  PivotTotalsRowPosition,
  DashboardState_PivotTotalsRowPosition
> = {
  top: DashboardState_PivotTotalsRowPosition.TOP,
  bottom: DashboardState_PivotTotalsRowPosition.BOTTOM,
};

export const FromProtoPivotTotalsRowPositionMap = {} as Record<
  DashboardState_PivotTotalsRowPosition,
  PivotTotalsRowPosition
>;
for (const op in ToProtoPivotTotalsRowPositionMap) {
  FromProtoPivotTotalsRowPositionMap[ToProtoPivotTotalsRowPositionMap[op]] = op;
}
