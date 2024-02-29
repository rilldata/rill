import {
  createAndExpression,
  createBinaryExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1Operation,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import * as yup from "yup";

export type AlertFormValues = {
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
  // The following fields are not editable in the form, but they're state that's used throughout the form, so
  // it's helpful to have them here. Also, in the future they may be editable in the form.
  metricsViewName: string;
  whereFilter: V1Expression;
  timeRange: V1TimeRange;
};

export function getAlertQueryArgsFromFormValues(
  formValues: AlertFormValues,
): V1MetricsViewAggregationRequest {
  return {
    metricsView: formValues.metricsViewName,
    measures: [{ name: formValues.measure }],
    dimensions: formValues.splitByDimension
      ? [{ name: formValues.splitByDimension }]
      : [],
    where: sanitiseExpression(formValues.whereFilter, undefined),
    having: sanitiseExpression(
      undefined,
      createAndExpression(
        formValues.criteria.map((c) =>
          createBinaryExpression(
            c.field,
            c.operation as V1Operation,
            Number(c.value),
          ),
        ),
      ),
    ),
    ...(formValues.splitByTimeGrain
      ? {
          timeRange: {
            isoDuration: formValues.splitByTimeGrain,
          },
        }
      : {}),
  };
}

export const alertFormValidationSchema = yup.object({
  name: yup.string().required("Required"),
  measure: yup.string().required("Required"),
  criteria: yup.array().of(
    yup.object().shape({
      field: yup.string().required("Required"),
      operation: yup.string().required("Required"),
      value: yup.number().required("Required"),
    }),
  ),
  criteriaOperation: yup.string().required("Required"),
  snooze: yup.string().required("Required"),
  recipients: yup.array().of(
    yup.object().shape({
      email: yup.string().email("Invalid email"),
    }),
  ),
});
export const FieldsByTab: (keyof AlertFormValues)[][] = [
  ["name", "measure"],
  ["criteria", "criteriaOperation"],
  ["snooze", "recipients"],
];

export function checkIsTabValid(
  tabIndex: number,
  formValues: AlertFormValues,
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
    (errors.criteria as unknown as any[])?.forEach?.((criteriaError) => {
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
