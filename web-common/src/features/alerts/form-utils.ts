import {
  createAndExpression,
  createBinaryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type AlertFormValue = {
  name: string;
  measure: string;
  splitByDimension: string;
  splitByTimeGrain: string;
  criteria: {
    field: string;
    operation: string;
    value: string;
  }[];
  criteriaOperation: V1Operation;
  snooze: string;
  recipients: { email: string }[];
};

export function getAlertQueryArgs(
  metricsViewName: string,
  formValues: AlertFormValue,
  dashboard: MetricsExplorerEntity,
): V1MetricsViewAggregationRequest {
  return {
    metricsView: metricsViewName,
    measures: [{ name: formValues.measure }],
    dimensions: formValues.splitByDimension
      ? [{ name: formValues.splitByDimension }]
      : [],
    where: dashboard.whereFilter,
    having: createAndExpression(
      formValues.criteria.map((c) =>
        createBinaryExpression(
          c.field,
          c.operation as V1Operation,
          Number(c.value),
        ),
      ),
    ),
  };
}

export function checkIsTabValid(
  tabIndex: number,
  formValues: AlertFormValue,
  errors: Record<string, string>,
): boolean {
  let hasRequiredFields: boolean;
  let hasErrors: boolean;

  if (tabIndex === 0) {
    hasRequiredFields = formValues.name !== "" && formValues.measure !== "";
    hasErrors = !!errors.name && !!errors.measure;
  } else if (tabIndex === 1) {
    hasRequiredFields = true;
    formValues.criteria.forEach((criteria) => {
      if (
        criteria.field === "" ||
        criteria.operation === "" ||
        criteria.value === ""
      ) {
        hasRequiredFields = false;
      }
    });
    hasErrors = false;
    (errors.criteria as unknown as any[]).forEach((criteriaError) => {
      if (
        criteriaError.field ||
        criteriaError.operation ||
        criteriaError.value
      ) {
        hasErrors = true;
      }
    });
  } else if (tabIndex === 2) {
    // TODO: do better for >1 recipients
    hasRequiredFields =
      formValues.snooze !== "" && formValues.recipients[0].email !== "";

    // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
    // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
    const receipientErrors = errors.recipients as unknown as {
      email: string;
    }[];

    hasErrors = !!errors.snooze || !!receipientErrors[0].email;
  } else {
    throw new Error(`Unexpected tabIndex: ${tabIndex}`);
  }

  return hasRequiredFields && !hasErrors;
}
