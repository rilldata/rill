import type { PivotRowJoinType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { Operation } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import { DashboardState_PivotRowJoinType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
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
  [V1Operation.OPERATION_NOT]: Operation.NOT,
  [V1Operation.OPERATION_IN]: Operation.IN,
  [V1Operation.OPERATION_NIN]: Operation.NIN,
  [V1Operation.OPERATION_LIKE]: Operation.LIKE,
  [V1Operation.OPERATION_NLIKE]: Operation.NLIKE,
};

export const FromProtoOperationMap = {} as Record<Operation, V1Operation>;
for (const op in ToProtoOperationMap) {
  FromProtoOperationMap[ToProtoOperationMap[op]] = op;
}

export const ToProtoTimeGrainMap: Record<V1TimeGrain, TimeGrain> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: TimeGrain.UNSPECIFIED,
  [V1TimeGrain.TIME_GRAIN_DAY]: TimeGrain.DAY,
  [V1TimeGrain.TIME_GRAIN_HOUR]: TimeGrain.HOUR,
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: TimeGrain.MILLISECOND,
  [V1TimeGrain.TIME_GRAIN_MINUTE]: TimeGrain.MINUTE,
  [V1TimeGrain.TIME_GRAIN_MONTH]: TimeGrain.MONTH,
  [V1TimeGrain.TIME_GRAIN_QUARTER]: TimeGrain.QUARTER,
  [V1TimeGrain.TIME_GRAIN_SECOND]: TimeGrain.SECOND,
  [V1TimeGrain.TIME_GRAIN_WEEK]: TimeGrain.WEEK,
  [V1TimeGrain.TIME_GRAIN_YEAR]: TimeGrain.YEAR,
};

export const FromProtoTimeGrainMap = {} as Record<TimeGrain, V1TimeGrain>;
for (const grain in ToProtoTimeGrainMap) {
  FromProtoTimeGrainMap[ToProtoTimeGrainMap[grain]] = grain;
}

export const ToProtoPivotRowJoinTypeMap: Record<
  PivotRowJoinType,
  DashboardState_PivotRowJoinType
> = {
  flat: DashboardState_PivotRowJoinType.FLAT,
  nest: DashboardState_PivotRowJoinType.NEST,
};

export const FromProtoPivotRowJoinTypeMap = {} as Record<
  DashboardState_PivotRowJoinType,
  PivotRowJoinType
>;
for (const op in ToProtoPivotRowJoinTypeMap) {
  FromProtoPivotRowJoinTypeMap[ToProtoPivotRowJoinTypeMap[op]] = op;
}
