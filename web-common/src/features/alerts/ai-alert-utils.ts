import {
  convertMapstructureToV1Expression,
  convertV1ExpressionToMapstructure,
} from "@rilldata/web-common/features/components/charts/custom/expression-utils";
import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { validateRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1AlertSpec,
  V1Expression,
  V1MetricsViewAggregationRequest,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

// An AI alert runs the ai resolver: the analyst judges the alert's Condition (the resolver's prompt) instead of a formula.
export const AIAlertResolver = "ai";

// The two kinds of condition the alert dialog writes: criteria over a measure, or a Condition the AI judges (an AI alert).
export const CriteriaConditionType = "criteria";
export const AIConditionType = "ai";
export type AlertConditionType =
  | typeof CriteriaConditionType
  | typeof AIConditionType;

export function isAIAlertSpec(
  alertSpec: Pick<V1AlertSpec, "resolver"> | undefined,
): boolean {
  return alertSpec?.resolver === AIAlertResolver;
}

// The ai resolver properties the dialog edits. Any other property (e.g. comparison_time_range, set through the API) is kept as it was.
const EditedProperties = [
  "prompt",
  "explore",
  "measures",
  "dimensions",
  "where",
  "time_range",
  "time_zone",
];

export type AIAlertDialogValues = {
  condition: string;
  exploreName: string;
  measure: string;
  splitByDimension: string;
  where: V1Expression | undefined;
  selectedTimeRange: Partial<DashboardTimeControls> | undefined;
  // The explore's default time range (its default preset), written when the user keeps it.
  defaultTimeRange?: string;
  timeZone: string;
  // The resolver properties of the alert being edited, if any.
  previous?: Record<string, unknown>;
};

/**
 * Maps the alert dialog's values to the ai resolver's properties.
 * The data tab's measure, dimension, filters and time range narrow what the AI looks at; empty ones are left out.
 */
export function getAIAlertResolverProperties(
  values: AIAlertDialogValues,
): Record<string, unknown> {
  const previous = values.previous ?? {};
  const props: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(previous)) {
    if (!EditedProperties.includes(key)) props[key] = value;
  }

  props.prompt = values.condition.trim();
  props.explore = values.exploreName;

  const measures = keepListWhileFirstUnchanged(
    values.measure,
    previous.measures,
  );
  if (measures) props.measures = measures;
  const dimensions = keepListWhileFirstUnchanged(
    values.splitByDimension,
    previous.dimensions,
  );
  if (dimensions) props.dimensions = dimensions;

  if (values.where?.cond?.exprs?.length) {
    props.where = convertV1ExpressionToMapstructure(values.where);
  }

  const timeRange = getAIAlertTimeRange(
    values.selectedTimeRange,
    values.defaultTimeRange,
  );
  if (timeRange) {
    props.time_range = timeRange;
    props.time_zone = values.timeZone;
  } else if (previous.time_zone) {
    // Without a time range the dialog has no time zone to write, but one written elsewhere still applies (e.g. to a comparison_time_range).
    props.time_zone = previous.time_zone;
  }

  return props;
}

/**
 * Maps an AI alert's resolver properties to the aggregation request the dialog and the alert page read their data context from.
 */
export function getAggregationRequestFromAIAlertProperties(
  props: Record<string, unknown>,
  metricsViewName: string,
): V1MetricsViewAggregationRequest {
  const req: V1MetricsViewAggregationRequest = {
    metricsView: metricsViewName,
    measures: ((props.measures as string[] | undefined) ?? []).map((name) => ({
      name,
    })),
    dimensions: ((props.dimensions as string[] | undefined) ?? []).map(
      (name) => ({ name }),
    ),
  };
  if (props.where) {
    req.where = convertMapstructureToV1Expression(
      props.where as Record<string, unknown>,
    );
  }
  const timeRange = getV1TimeRange(
    props.time_range as Record<string, string> | undefined,
  );
  if (timeRange) {
    req.timeRange = {
      ...timeRange,
      ...(props.time_zone ? { timeZone: props.time_zone as string } : {}),
    };
  }
  return req;
}

function keepListWhileFirstUnchanged(
  first: string,
  previous: unknown,
): string[] | undefined {
  if (!first) return undefined;
  const list = previous as string[] | undefined;
  if (list?.[0] === first) return list;
  return [first];
}

// The ai resolver takes a time range as a rill time expression or as start and end.
// Rill time also parses the dashboards' preset names (e.g. P4W, rill-PDC), so a preset is written as its name.
function getAIAlertTimeRange(
  selectedTimeRange: Partial<DashboardTimeControls> | undefined,
  defaultTimeRange: string | undefined,
): Record<string, string> | undefined {
  const name = selectedTimeRange?.name;
  if (!name || name === TimeRangePreset.ALL_TIME) return undefined;
  if (name === TimeRangePreset.CUSTOM) {
    if (!selectedTimeRange.start || !selectedTimeRange.end) return undefined;
    return {
      start: selectedTimeRange.start.toISOString(),
      end: selectedTimeRange.end.toISOString(),
    };
  }
  // The explore's default preset is written as the range it stands for, so what the user sees is what is judged,
  // even if the explore's default changes later. An explore without a default range shows all time.
  if (name === TimeRangePreset.DEFAULT) {
    if (!defaultTimeRange || defaultTimeRange === TimeRangePreset.ALL_TIME)
      return undefined;
    return { expression: defaultTimeRange };
  }
  return { expression: name };
}

// The inverse of getAIAlertTimeRange, as the dashboards' time controls represent each kind of range.
function getV1TimeRange(
  timeRange: Record<string, string> | undefined,
): V1TimeRange | undefined {
  if (!timeRange) return undefined;
  if (timeRange.start && timeRange.end) {
    return { start: timeRange.start, end: timeRange.end };
  }
  const expression = timeRange.expression;
  if (!expression) return undefined;
  const previousComplete =
    PreviousCompleteRangeMap[expression as TimeRangePreset];
  if (previousComplete) return { ...previousComplete };
  const isPreset = (Object.values(TimeRangePreset) as string[]).includes(
    expression,
  );
  if (!isPreset && !validateRillTime(expression)) return { expression };
  return { isoDuration: expression };
}

// An AI alert's default name is its Condition, shortened.
export function generateAIAlertName(condition: string): string {
  const trimmed = condition.trim().replace(/\s+/g, " ");
  if (trimmed.length <= AIAlertNameMaxLength) return trimmed;
  return trimmed.slice(0, AIAlertNameMaxLength - 1).trimEnd() + "…";
}

const AIAlertNameMaxLength = 60;
