import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  mapMeasureFilterToExpr,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterType } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { getInitialScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import type {
  V1ExploreSpec,
  V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { ValidationErrors } from "sveltekit-superforms";
import { yup, type ValidationAdapter } from "sveltekit-superforms/adapters";
import { object, array, string } from "yup";

export type AlertFormValues = {
  name: string;
  measure: string;
  splitByDimension: string;
  criteria: MeasureFilterEntry[];
  criteriaOperation: V1Operation;
  evaluationInterval: string;
  snooze: string;
  enableSlackNotification: boolean;
  slackChannels: string[];
  slackUsers: string[];
  enableEmailNotification: boolean;
  emailRecipients: string[];
  refreshWhenDataRefreshes: boolean;
  // The following fields are not editable in the form, but they're state that's used throughout the form, so
  // it's helpful to have them here. Also, in the future they may be editable in the form.
  metricsViewName: string;
  exploreName: string;
} & ReturnType<typeof getInitialScheduleFormValues>;

export function getAlertQueryArgsFromFormValues(
  formValues: AlertFormValues,
  filtersArgs: FiltersState,
  timeControlArgs: TimeControlState,
  exploreSpec: V1ExploreSpec,
): V1MetricsViewAggregationRequest {
  const timeRange = mapSelectedTimeRangeToV1TimeRange(
    timeControlArgs.selectedTimeRange,
    timeControlArgs.selectedTimezone,
    exploreSpec,
  );
  const comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
    timeControlArgs.selectedComparisonTimeRange,
    timeControlArgs.showTimeComparison,
    timeRange,
  );

  return {
    metricsView: formValues.metricsViewName,
    measures: [
      {
        name: formValues.measure,
      },
      ...(comparisonTimeRange
        ? [
            {
              name: formValues.measure + ComparisonDeltaAbsoluteSuffix,
              comparisonDelta: { measure: formValues.measure },
            },
            {
              name: formValues.measure + ComparisonDeltaRelativeSuffix,
              comparisonRatio: { measure: formValues.measure },
            },
          ]
        : []),
      ...(formValues.criteria.some(
        (c) => c.type === MeasureFilterType.PercentOfTotal,
      )
        ? [
            {
              name: formValues.measure + ComparisonPercentOfTotal,
              percentOfTotal: { measure: formValues.measure },
            },
          ]
        : []),
    ],
    dimensions: formValues.splitByDimension
      ? [{ name: formValues.splitByDimension }]
      : [],
    where: sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        filtersArgs.whereFilter,
        filtersArgs.dimensionThresholdFilters,
      ),
      undefined,
    ),
    having: sanitiseExpression(undefined, {
      cond: {
        op: formValues.criteriaOperation,
        exprs: formValues.criteria
          .map(mapMeasureFilterToExpr)
          .filter((e) => !!e),
      },
    }),
    timeRange,
    sort: [
      {
        name: formValues.measure,
        desc: true,
      },
    ],
    comparisonTimeRange,
  };
}

export const alertFormValidationSchema = yup(
  object({
    name: string().required("Required"),
    measure: string().required("Required"),
    criteria: array().of(
      object().shape({
        measure: string().required("Required"),
        operation: string().required("Required"),
        type: string().required("Required"),
        value1: string()
          .required("Required")
          .test((value, context) => {
            // `number` doest allow for string representation of number with the superforms yup adapter.
            // So we use `string` and check for NaN
            // TODO: do a greater refactor changing the type of value1 in all the places to a number
            const numValue = Number(value);
            if (Number.isNaN(numValue)) {
              return context.createError({
                message: `${context.path} must be a valid number.`,
              });
            }

            const criteria = context.parent as MeasureFilterEntry;
            if (
              criteria.type === MeasureFilterType.PercentOfTotal &&
              (numValue < 0 || numValue > 100)
            ) {
              return context.createError({
                message: `${context.path} must be a value between 0 and 100.`,
              });
            }
            return true;
          }),
      }),
    ),
    criteriaOperation: string().required("Required"),
    snooze: string().required("Required"),
    slackUsers: array(string().email("Invalid email")),
    emailRecipients: array(string().email("Invalid email")),
  }),
) as unknown as ValidationAdapter<AlertFormValues>;
export const FieldsByTab: (keyof AlertFormValues)[][] = [
  ["measure"],
  ["criteria", "criteriaOperation"],
  ["name", "snooze", "slackUsers", "emailRecipients"],
];

export function checkIsTabValid(
  tabIndex: number,
  formValues: AlertFormValues,
  errors: ValidationErrors<AlertFormValues> | undefined,
): boolean {
  if (!errors) return true;

  let hasRequiredFields: boolean;
  let hasErrors: boolean;

  if (tabIndex === 0) {
    hasRequiredFields = formValues.measure !== "";
    hasErrors = !!errors.measure;
  } else if (tabIndex === 1) {
    hasRequiredFields = true;
    formValues.criteria.forEach((criteria) => {
      if (
        criteria.measure === "" ||
        (criteria.operation as string) === "" ||
        criteria.value1 === ""
      ) {
        hasRequiredFields = false;
      }
    });
    hasErrors = someCriteriaHasErrors(errors.criteria);
  } else if (tabIndex === 2) {
    // TODO: do better for >1 recipients
    hasRequiredFields =
      !formValues.name && !formValues.snooze && !!formValues.emailRecipients[0];

    hasErrors =
      !!errors.name || !!errors.snooze || !!errors.emailRecipients?.[0]?.length;
  } else {
    throw new Error(`Unexpected tabIndex: ${tabIndex}`);
  }

  return hasRequiredFields && !hasErrors;
}

function someCriteriaHasErrors(
  criteriaErrors: ValidationErrors<AlertFormValues>["criteria"],
) {
  if (!criteriaErrors) return false;
  return Object.values(criteriaErrors).every((criteriaError) => {
    if (!criteriaError) return false;
    return Object.values(criteriaError).every((c: string[]) => !!c?.[0]);
  });
}
