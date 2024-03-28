import { mapAlertCriteriaToExpression } from "@rilldata/web-common/features/alerts/criteria-tab/map-alert-criteria";
import type {
  CompareWith,
  CriteriaOperations,
} from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import {
  createAndExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1Expression,
  V1MetricsViewComparisonRequest,
  V1Operation,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import * as yup from "yup";

export type AlertCriteria = {
  field: string;
  operation: CriteriaOperations;
  compareWith: CompareWith;
  value: string;
};
export type AlertFormValues = {
  name: string;
  measure: string;
  splitByDimension: string;
  criteria: AlertCriteria[];
  criteriaOperation: V1Operation;
  evaluationInterval: string;
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
): V1MetricsViewComparisonRequest {
  return {
    metricsViewName: formValues.metricsViewName,
    measures: [{ name: formValues.measure }],
    dimension: formValues.splitByDimension
      ? { name: formValues.splitByDimension }
      : undefined,
    where: sanitiseExpression(formValues.whereFilter, undefined),
    having: sanitiseExpression(
      undefined,
      createAndExpression(
        formValues.criteria
          .map(mapAlertCriteriaToExpression)
          .filter((e) => !!e) as V1Expression[],
      ),
    ),
    timeRange: {
      isoDuration: formValues.timeRange.isoDuration,
    },
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
        (criteria.operation as string) === "" ||
        criteria.value === ""
      ) {
        hasRequiredFields = false;
      }
    });
    hasErrors = !!errors.criteria;
  } else if (tabIndex === 2) {
    // TODO: do better for >1 recipients
    hasRequiredFields =
      formValues.snooze !== "" && formValues.recipients[0].email !== "";

    // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
    // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
    const recipientErrors = errors.recipients as unknown as {
      email: string;
    }[];

    hasErrors = !!errors.snooze || !!recipientErrors[0].email;
  } else {
    throw new Error(`Unexpected tabIndex: ${tabIndex}`);
  }

  return hasRequiredFields && !hasErrors;
}
