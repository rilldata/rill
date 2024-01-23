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
  isAnd: boolean,
): V1Expression {
  const exprs = formInput
    .filter((fi) => fi.operation in V1Operation)
    .map((fi) => createBinaryExpression(fi.field, fi.operation, fi.value));
  if (isAnd) {
    return createAndExpression(exprs);
  } else {
    return createOrExpression(exprs);
  }
}
