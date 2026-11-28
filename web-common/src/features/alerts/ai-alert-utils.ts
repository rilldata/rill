import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { convertV1ExpressionToMapstructure } from "@rilldata/web-common/features/components/charts/custom/expression-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import type { AIResolverProps } from "@rilldata/web-common/features/explore-mappers/map-ai-resolver-props-to-metrics-resolver-query";
import { mapResolverExpressionToV1Expression } from "@rilldata/web-common/features/explore-mappers/map-resolver-expression-to-v1-expression";
import {
  convertRequestKeysToCamelCase,
  getSpecQuery,
} from "@rilldata/web-common/features/explore-mappers/utils";
import type {
  V1AlertSpec,
  V1Expression,
  V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

// An alert fires on criteria over a measure, or when the AI analyst judges that a condition written in words holds (an AI alert).
export const CriteriaConditionType = "criteria";
export const AIConditionType = "ai";
export type AlertConditionType =
  | typeof CriteriaConditionType
  | typeof AIConditionType;

// The properties of the `ai` resolver that an AI alert runs; the prompt is its condition.
export type AIAlertResolverProps = AIResolverProps & { prompt: string };

// AI alerts run the `ai` resolver, which judges the condition instead of evaluating criteria.
export function isAIAlertSpec(alertSpec: V1AlertSpec): boolean {
  return alertSpec.resolver === "ai";
}

/**
 * Maps the alert dialog's values to the properties of the `ai` resolver.
 * The data tab's measure, dimension, filters and time range narrow what the AI looks at; empty ones are left out.
 */
export function getAIAlertResolverProps(
  values: AlertFormValues,
  where: V1Expression | undefined,
  timeFilterManager: TimeFilterManager,
): AIAlertResolverProps {
  const props: AIAlertResolverProps = {
    prompt: values.aiCondition.trim(),
    explore: values.exploreName,
  };
  if (values.measure) props.measures = [values.measure];
  if (values.splitByDimension) props.dimensions = [values.splitByDimension];

  const sanitisedWhere = sanitiseExpression(where);
  if (sanitisedWhere) {
    props.where = convertV1ExpressionToMapstructure(
      sanitisedWhere,
    ) as Expression;
  }

  // The resolver evaluates the time filter's range with rill time,
  // which also reads the legacy presets the filter can hold, such as P7D or rill-PDC.
  const { requestedTimeRange, timeDimension, timeZone } = timeFilterManager;
  if (requestedTimeRange) {
    props.time_range = {
      expression: requestedTimeRange,
      ...(timeDimension ? { time_dimension: timeDimension } : {}),
    };
    props.time_zone = timeZone;
  }

  return props;
}

/**
 * Returns the aggregation request that the alert dialog and the alert page read an alert's data from:
 * a criteria alert's query, or the request that an AI alert's resolver properties narrow the data to.
 */
export function getAlertAggregationRequest(
  alertSpec: V1AlertSpec,
  metricsViewName: string,
): V1MetricsViewAggregationRequest {
  if (!isAIAlertSpec(alertSpec)) {
    return JSON.parse(getSpecQuery(alertSpec).queryArgsJson || "{}");
  }

  const props = (alertSpec.resolverProperties ?? {}) as AIAlertResolverProps;
  return {
    metricsView: metricsViewName,
    measures: props.measures?.map((name) => ({ name })) ?? [],
    dimensions: props.dimensions?.map((name) => ({ name })) ?? [],
    where: mapResolverExpressionToV1Expression(props.where),
    timeRange: props.time_range && {
      ...convertRequestKeysToCamelCase(props.time_range),
      timeZone: props.time_zone,
    },
  };
}

const AIAlertNameMaxLength = 60;

// An AI alert's default name is its condition, shortened.
export function generateAIAlertName(condition: string): string {
  const name = condition.trim().replace(/\s+/g, " ");
  if (name.length <= AIAlertNameMaxLength) return name;
  return name.slice(0, AIAlertNameMaxLength - 1).trimEnd() + "…";
}
