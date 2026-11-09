import {
  AIAlertResolver,
  AIConditionType,
  type AlertConditionType,
  CriteriaConditionType,
  getAIAlertResolverProperties,
  isAIAlertSpec,
} from "@rilldata/web-common/features/alerts/ai-alert-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
  ComparisonPercentOfTotal,
  mapMeasureFilterToExpr,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterType } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { mapEphemeralMeasuresForRequest } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { getInitialScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import type {
  V1AlertSpec,
  V1ExploreSpec,
  V1ExportFormat,
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { ValidationErrors } from "sveltekit-superforms";
import { yup, type ValidationAdapter } from "sveltekit-superforms/adapters";
import { getExportLimitSchema } from "@rilldata/web-common/features/exports/export-file-options";
import { object, array, string } from "yup";

// An alert fires either on criteria over a measure, or when the analyst judges that an AI alert's Condition holds.
export type { AlertConditionType };

export type AlertFormValues = {
  name: string;
  conditionType: AlertConditionType;
  // The AI alert's Condition, used when conditionType is AIConditionType.
  aiCondition: string;
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
  // Options for attaching the rows that triggered the alert to emails
  exportAttach: boolean;
  exportFormat: V1ExportFormat;
  exportLimit: string;
  exportIncludeHeader: boolean;
  refreshWhenDataRefreshes: boolean;
  // The following fields are not editable in the form, but they're state that's used throughout the form, so
  // it's helpful to have them here. Also, in the future they may be editable in the form.
  metricsViewName: string;
  exploreName: string;
  // Definitions for the dashboard's ephemeral measures, so an alert can be
  // built on one: the saved query embeds the expression it needs.
  ephemeralMeasures?: EphemeralMeasureDef[];
} & ReturnType<typeof getInitialScheduleFormValues>;

export function getAlertQueryArgsFromFormValues(
  formValues: AlertFormValues,
  expr: V1Expression | undefined,
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
    measures: mapEphemeralMeasuresForRequest(
      [
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
      formValues.ephemeralMeasures,
    ),
    dimensions: formValues.splitByDimension
      ? [{ name: formValues.splitByDimension }]
      : [],
    where: sanitiseExpression(expr, undefined),
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

/**
 * The resolver options the alert dialog saves: an AI alert's ai resolver properties, or a criteria alert's metrics query.
 * `previousSpec` is the alert being edited. When it already was an AI alert, the ai resolver properties the dialog does not edit survive;
 * a criteria alert's properties (query_name, query_args_json) belong to another resolver and are dropped.
 */
export function getAlertResolverOptions(
  values: AlertFormValues,
  where: V1Expression | undefined,
  timeControlState: TimeControlState,
  exploreSpec: V1ExploreSpec,
  previousSpec?: Pick<V1AlertSpec, "resolver" | "resolverProperties">,
):
  | { resolver: string; resolverProperties: Record<string, unknown> }
  | { queryName: string; queryArgsJson: string } {
  if (values.conditionType === AIConditionType) {
    return {
      resolver: AIAlertResolver,
      resolverProperties: getAIAlertResolverProperties({
        condition: values.aiCondition,
        exploreName: values.exploreName,
        measure: values.measure,
        splitByDimension: values.splitByDimension,
        where,
        selectedTimeRange: timeControlState.selectedTimeRange,
        defaultTimeRange: exploreSpec.defaultPreset?.timeRange,
        timeZone: timeControlState.selectedTimezone,
        previous: isAIAlertSpec(previousSpec)
          ? previousSpec?.resolverProperties
          : undefined,
      }),
    };
  }
  return {
    queryName: "MetricsViewAggregation",
    queryArgsJson: JSON.stringify(
      getAlertQueryArgsFromFormValues(
        values,
        where,
        timeControlState,
        exploreSpec,
      ),
    ),
  };
}

export const alertFormValidationSchema = yup(
  object({
    name: string().required("Required"),
    conditionType: string().oneOf([CriteriaConditionType, AIConditionType]),
    aiCondition: string().when("conditionType", {
      is: AIConditionType,
      then: (s) => s.trim().required("Required"),
    }),
    // An AI alert's measure only narrows what the AI looks at, so it is optional; its criteria are not used.
    measure: string().when("conditionType", {
      is: AIConditionType,
      otherwise: (s) => s.required("Required"),
    }),
    criteria: array().when("conditionType", {
      is: AIConditionType,
      otherwise: (s) =>
        s.of(
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
    }),
    criteriaOperation: string().required("Required"),
    snooze: string().required("Required"),
    slackUsers: array(string().email("Invalid email")),
    emailRecipients: array(string().email("Invalid email")),
    // The row limit input is hidden unless attachments are enabled for email delivery
    exportLimit: string().when(["enableEmailNotification", "exportAttach"], {
      is: (enableEmailNotification: boolean, exportAttach: boolean) =>
        enableEmailNotification && exportAttach,
      then: () => getExportLimitSchema(),
    }),
  }),
) as unknown as ValidationAdapter<AlertFormValues>;
export const FieldsByTab: (keyof AlertFormValues)[][] = [
  ["measure"],
  ["criteria", "criteriaOperation", "aiCondition"],
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
    hasRequiredFields =
      formValues.conditionType === AIConditionType || formValues.measure !== "";
    hasErrors = !!errors.measure;
  } else if (tabIndex === 1 && formValues.conditionType === AIConditionType) {
    hasRequiredFields = formValues.aiCondition.trim() !== "";
    hasErrors = !!errors.aiCondition;
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
