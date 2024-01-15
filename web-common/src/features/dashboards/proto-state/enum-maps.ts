import { Operation } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import { V1Operation } from "@rilldata/web-common/runtime-client";

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
};

export const FromProtoOperationMap = {} as Record<Operation, V1Operation>;
for (const op in ToProtoOperationMap) {
  FromProtoOperationMap[ToProtoOperationMap[op]] = op;
}
