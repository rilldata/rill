import {
  createAndExpression,
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type FormFilterInput = {
  field: string;
  operation: V1Operation;
  value: number;
};

export function translateFilter(
  formInput: FormFilterInput[],
  joiner: V1Operation,
): V1Expression {
  const exprs = formInput
    .filter((fi) => fi.operation in V1Operation)
    .map((fi) => createBinaryExpression(fi.field, fi.operation, fi.value));
  if (joiner === V1Operation.OPERATION_AND) {
    return createAndExpression(exprs);
  } else {
    return createOrExpression(exprs);
  }
}
