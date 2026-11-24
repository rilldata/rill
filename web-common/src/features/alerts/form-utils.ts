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
  mapTimeManagerComparisonRangeToV1TimeRange,
  mapTimeManagerRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import { getInitialScheduleFormValues } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
import type {
  V1ExportFormat,
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { ValidationErrors } from "sveltekit-superforms";
import { yup, type ValidationAdapter } from "sveltekit-superforms/adapters";
import { getExportLimitSchema } from "@rilldata/web-common/features/exports/export-file-options";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { object, array, string } from "yup";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

// The number of matching rows listed in notifications when an alert doesn't set it.
// Keep in sync with alertNotifyRowsLimitDefault in runtime/parser/parse_alert.go.
export const AlertNotifyRowsLimitDefault = 5;
// Keep in sync with alertNotifyRowsLimitMax in runtime/parser/parse_alert.go.
export const AlertNotifyRowsLimitMax = 50;

export type AlertFormValues = {
  name: string;
  measure: string;
  splitByDimension: string;
  // The maximum number of matching rows listed in notifications, which only matters when the alert is split by a dimension
  notifyRowsLimit: string;
  criteria: MeasureFilterEntry[];
  criteriaOperation: V1Operation;
  evaluationInterval: string;
  snooze: string;
  enableSlackNotification: boolean;
  slackChannels: string[];
  slackUsers: string[];
  enableEmailNotification: boolean;
  emailRecipients: string[];
  // Email of the embed user who creates the alert; only shown in an embed if rill.yaml configures `embeds.open_url`.
  embedOwnerEmail: string;
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
  timeFilterManager: TimeFilterManager | undefined,
): V1MetricsViewAggregationRequest {
  const timeRange = mapTimeManagerRangeToV1TimeRange(timeFilterManager);
  const comparisonTimeRange = mapTimeManagerComparisonRangeToV1TimeRange(
    timeFilterManager,
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

// Returns the limit on the number of matching rows listed in notifications to save for an alert.
// It returns 0, which means the runtime's default, when the alert isn't split by a dimension (so there's a single row)
// or when the limit is the default. Leaving the limit unset keeps it out of the alert's YAML,
// which runtimes that predate the limit can't parse.
export function getNotifyRowsLimitFromFormValues(
  formValues: AlertFormValues,
): number {
  if (!formValues.splitByDimension) return 0;
  const limit = Number(formValues.notifyRowsLimit);
  return limit === AlertNotifyRowsLimitDefault ? 0 : limit;
}

export const alertFormValidationSchema = yup(
  object({
    name: string().required("Required"),
    measure: string().required("Required"),
    // The rows limit input is hidden unless the alert is split by a dimension
    notifyRowsLimit: string().when("splitByDimension", {
      is: (splitByDimension: string) => !!splitByDimension,
      then: (schema) =>
        schema.test(
          "notify-rows-limit",
          () =>
            m.alert_form_data_rows_limit_invalid({
              max: AlertNotifyRowsLimitMax,
            }),
          (value) =>
            !!value &&
            /^\d+$/.test(value) &&
            Number(value) >= 1 &&
            Number(value) <= AlertNotifyRowsLimitMax,
        ),
    }),
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
                message: m.alert_form_criteria_value_invalid({
                  field: context.path,
                }),
              });
            }

            const criteria = context.parent as MeasureFilterEntry;
            // Percent of total compares each dimension value's share of the measure, so it needs a split.
            // Removing the split keeps such criteria but blocks saving until a split is chosen again,
            // instead of saving an alert that has lost its condition.
            const formValues = context.from?.at(-1)?.value as
              | AlertFormValues
              | undefined;
            if (
              criteria.type === MeasureFilterType.PercentOfTotal &&
              !formValues?.splitByDimension
            ) {
              return context.createError({
                message: m.alert_form_criteria_percent_of_total_needs_split({
                  field: context.path,
                }),
              });
            }
            if (
              criteria.type === MeasureFilterType.PercentOfTotal &&
              (numValue < 0 || numValue > 100)
            ) {
              return context.createError({
                message: m.alert_form_criteria_value_percent_range({
                  field: context.path,
                }),
              });
            }
            return true;
          }),
      }),
    ),
    criteriaOperation: string().required("Required"),
    snooze: string().required("Required"),
    slackUsers: array(string().email("Invalid email")),
    // Checked on the email recipients rather than the whole form, so the error shows up on the delivery tab.
    emailRecipients: array(string().email("Invalid email")).test(
      "has-notification",
      () => m.alert_form_no_notification(),
      (_, context) => hasNotification(context.parent as AlertFormValues),
    ),
    embedOwnerEmail: string().email("Invalid email"),
    // The row limit input is hidden unless attachments are enabled for email delivery
    exportLimit: string().when(["enableEmailNotification", "exportAttach"], {
      is: (enableEmailNotification: boolean, exportAttach: boolean) =>
        enableEmailNotification && exportAttach,
      then: () => getExportLimitSchema(),
    }),
  }),
) as unknown as ValidationAdapter<AlertFormValues>;
export const FieldsByTab: (keyof AlertFormValues)[][] = [
  ["measure", "notifyRowsLimit"],
  ["criteria", "criteriaOperation"],
  ["name", "snooze", "slackUsers", "emailRecipients", "embedOwnerEmail"],
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
    hasErrors = !!errors.measure || !!errors.notifyRowsLimit;
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

// hasNotification returns true if an alert created from an embed is sent to someone:
// the embed creator's email (always a recipient if set) or an email recipient.
// Alerts created from an embed can't be sent to Slack.
// Other alerts aren't checked, since they have always been allowed without recipients.
function hasNotification(values: AlertFormValues) {
  if (!EmbedStore.isEmbedded()) return true;
  return (
    !!values.embedOwnerEmail?.trim() ||
    (values.enableEmailNotification &&
      values.emailRecipients.some((r) => r.trim() !== ""))
  );
}

function someCriteriaHasErrors(
  criteriaErrors: ValidationErrors<AlertFormValues>["criteria"],
) {
  if (!criteriaErrors) return false;
  // `validate(path)` clears a passing field's error to undefined,
  // so look for any criterion field that still has an error.
  return Object.values(criteriaErrors).some((criteriaError) => {
    if (!criteriaError) return false;
    return Object.values(criteriaError).some((c: string[]) => !!c?.[0]);
  });
}
