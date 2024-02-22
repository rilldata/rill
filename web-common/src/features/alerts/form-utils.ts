import {
  createAndExpression,
  createBinaryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  V1Expression,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
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

// TODO: revisit if Partial<AlertFormValues> could work instead
export type AlertFormValuesSubset = {
  metricsViewName: string;
  whereFilter: V1Expression;
  timeRange: V1TimeRange;
  measure: string;
  splitByDimension: string;
  splitByTimeGrain: string;
  criteria: {
    field: string;
    operation: string;
    value: string;
  }[];
  criteriaOperation: V1Operation;
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
    where: formValues.whereFilter,
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

export function getFormValuesFromAlertQueryArgs(
  queryArgs: V1MetricsViewAggregationRequest,
): AlertFormValuesSubset {
  if (!queryArgs) return {} as AlertFormValuesSubset;

  console.log("queryArgs", queryArgs);

  const measures = queryArgs.measures as V1MetricsViewAggregationMeasure[];
  const dimensions =
    queryArgs.dimensions as V1MetricsViewAggregationDimension[];

  return {
    // TODO: get measure label, if available
    measure: measures[0].name as string,
    // TODO: ensure I don't pick up a time dimension
    splitByDimension:
      dimensions.length > 0 ? (dimensions[0].name as string) : "",
    // TODO: filter the dimensions list for a time dimension
    splitByTimeGrain: "",
    // TODO: get criteria from queryArgs
    criteria: [
      {
        field: "",
        operation: "",
        value: "0",
      },
    ],
    criteriaOperation: V1Operation.OPERATION_AND,
    // These are not part of the form, but are used to track the state of the form
    metricsViewName: queryArgs.metricsView as string,
    whereFilter: queryArgs.where as V1Expression,
    timeRange: (queryArgs.timeRange as V1TimeRange) ?? { isoOffset: "P7D" },
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
