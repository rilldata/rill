import {
  CompareWith,
  CriteriaOperations,
} from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import type { AlertCriteria } from "@rilldata/web-common/features/alerts/form-utils";
import {
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

const DeltaAbsoluteSuffix = "__delta_abs";
const DeltaRelativeSuffix = "__delta_rel";
const HasSuffixRegex = /__delta_(rel|abs)/;

const CriteriaOperationsToProtoOperation = {
  [CriteriaOperations.GreaterThan]: V1Operation.OPERATION_GT,
  [CriteriaOperations.GreaterThanOrEquals]: V1Operation.OPERATION_GTE,
  [CriteriaOperations.LessThan]: V1Operation.OPERATION_LT,
  [CriteriaOperations.LessThanOrEquals]: V1Operation.OPERATION_LTE,
};
const ProtoOperationToCriteriaOperations: Partial<
  Record<V1Operation, CriteriaOperations>
> = {};
for (const criteriaOperation in CriteriaOperationsToProtoOperation) {
  ProtoOperationToCriteriaOperations[
    CriteriaOperationsToProtoOperation[criteriaOperation]
  ] = criteriaOperation;
}
const ProtoOperationToCompareCriteriaOperations = {
  [V1Operation.OPERATION_GT]: CriteriaOperations.IncreasesBy,
  [V1Operation.OPERATION_GTE]: CriteriaOperations.IncreasesBy,
  [V1Operation.OPERATION_LT]: CriteriaOperations.DecreasesBy,
  [V1Operation.OPERATION_LTE]: CriteriaOperations.DecreasesBy,
};

export function mapAlertCriteriaToExpression(
  criteria: AlertCriteria,
): V1Expression | undefined {
  const value =
    Number(criteria.value) /
    (criteria.compareWith === CompareWith.Percent ? 100 : 1);
  const comparisonSuffix =
    criteria.compareWith === CompareWith.Percent
      ? DeltaRelativeSuffix
      : DeltaAbsoluteSuffix;

  switch (criteria.operation) {
    // <field> > <value>
    // <field> >= <value>
    // <field> < <value>
    // <field> <= <value>
    case CriteriaOperations.GreaterThan:
    case CriteriaOperations.GreaterThanOrEquals:
    case CriteriaOperations.LessThan:
    case CriteriaOperations.LessThanOrEquals:
      if (criteria.compareWith === CompareWith.Percent) return undefined;
      return createBinaryExpression(
        criteria.field,
        CriteriaOperationsToProtoOperation[criteria.operation],
        value,
      );

    case CriteriaOperations.IncreasesBy:
      // Δ<field> > <value>
      // or
      // Δ%<field> > <value>
      return createBinaryExpression(
        criteria.field + comparisonSuffix,
        V1Operation.OPERATION_GT,
        value,
      );

    case CriteriaOperations.DecreasesBy:
      // Δ<field> < -<value>
      // or
      // Δ%<field> < -<value>
      return createBinaryExpression(
        criteria.field + comparisonSuffix,
        V1Operation.OPERATION_LT,
        -Math.abs(value),
      );

    case CriteriaOperations.ChangesBy:
      // Δ<field> < -<value> && Δ<field> > <value>
      // or
      // Δ%<field> < -<value> && Δ%<field> > <value>
      return createOrExpression([
        createBinaryExpression(
          criteria.field + comparisonSuffix,
          V1Operation.OPERATION_LT,
          -Math.abs(value),
        ),
        createBinaryExpression(
          criteria.field + comparisonSuffix,
          V1Operation.OPERATION_GT,
          value,
        ),
      ]);
  }

  return undefined;
}

export function mapExpressionToAlertCriteria(
  expr: V1Expression,
): AlertCriteria {
  let value = 0;
  let field = "";
  let operation = CriteriaOperations.GreaterThan;
  switch (expr.cond?.op) {
    case V1Operation.OPERATION_OR:
      field = expr.cond.exprs?.[0].cond?.exprs?.[0].ident ?? "";
      value = (expr.cond.exprs?.[1].cond?.exprs?.[1].val as number) * 100 ?? 0;
      operation = CriteriaOperations.ChangesBy;
      break;

    case V1Operation.OPERATION_GT:
    case V1Operation.OPERATION_GTE:
    case V1Operation.OPERATION_LT:
    case V1Operation.OPERATION_LTE:
      field = expr.cond.exprs?.[0].ident ?? "";
      value = Math.abs((expr.cond.exprs?.[1].val as number) ?? 0);
      if (field.endsWith(DeltaRelativeSuffix)) {
        // convert decimal to percent
        value *= 100;
      }
      operation =
        (HasSuffixRegex.test(field)
          ? ProtoOperationToCompareCriteriaOperations[expr.cond?.op]
          : ProtoOperationToCriteriaOperations[expr.cond?.op]) ??
        CriteriaOperations.GreaterThan;
      break;
  }

  return {
    field: field.replace(HasSuffixRegex, ""),
    value: value.toString(),
    operation,
    compareWith: field.endsWith(DeltaRelativeSuffix)
      ? CompareWith.Percent
      : CompareWith.Value,
  };
}
