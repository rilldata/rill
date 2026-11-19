import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getQueryServiceConvertExpressionToMetricsSQLQueryOptions,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { type Document, YAMLMap, YAMLSeq } from "yaml";

export function getMissingRequiredFilters(
  expressionFilterManager: ExpressionFilterManager,
  metricsViewsProvider: MetricsViewsProvider,
  yamlConfigProvider: YAMLConfigProvider,
) {
  return yamlConfigProvider.requiredFilters.filter((filterName) => {
    const isValidDimensionOrMeasure =
      filterName in metricsViewsProvider.dimensionSpecs ||
      filterName in metricsViewsProvider.measureSpecs;
    const hasNoFilter =
      !expressionFilterManager.filterManagersMap[filterName]?.expr;
    return isValidDimensionOrMeasure && hasNoFilter;
  });
}

/**
 * Converts a filter expression to Metrics SQL, the form in which a dashboard's YAML holds its default filters.
 * Rejects when the request fails.
 */
export async function convertExpressionToMetricsSQL(
  client: RuntimeClient,
  expr: V1Expression,
): Promise<string> {
  const resp = await queryClient.fetchQuery(
    getQueryServiceConvertExpressionToMetricsSQLQueryOptions(client, {
      // TODO: the API type is the proto message, but there is a fromJson underneath, which takes this JSON form.
      //       Once that is fixed, this needs to call toExpressionProto.
      expression: expr as any,
    }),
  );
  return resp.sql ?? "";
}

/**
 * Writes `filters.pinned` and `filters.required` to a dashboard's YAML.
 * An empty list removes its key, and the `filters` block is removed once nothing is left in it.
 * Other keys in the block are left as they are.
 *
 * `keyPath` locates the dashboard's properties in the document:
 * empty for a canvas or explore file, `["explore"]` for an explore defined inside a metrics view.
 */
export function setPinnedAndRequiredFilters(
  yaml: Document,
  pinned: string[],
  required: string[],
  keyPath: string[] = [],
) {
  const filtersPath = [...keyPath, "filters"];
  Object.entries({ pinned, required }).forEach(([key, names]) => {
    if (names.length > 0) {
      yaml.setIn([...filtersPath, key], names);
    } else if (yaml.hasIn([...filtersPath, key])) {
      yaml.deleteIn([...filtersPath, key]);
    }
  });

  const filters = yaml.getIn(filtersPath);
  if (filters instanceof YAMLMap && filters.items.length === 0) {
    yaml.deleteIn(filtersPath);
  }
}

/**
 * Takes fields out of the filter config in a dashboard's YAML.
 * For when the fields are removed from an explore, whose filter config may only name its own fields.
 *
 * `filters.pinned` and `filters.required` lose the fields.
 * `defaults.filters` is Metrics SQL, which can't be edited here, so it is removed as a whole when it uses one of them.
 * `defaultFilter` is that filter as the runtime parsed it.
 */
export function removeFieldsFromFilterConfig(
  yaml: Document,
  fields: Set<string>,
  defaultFilter: V1Expression | undefined,
  keyPath: string[] = [],
) {
  const [pinned, required] = ["pinned", "required"].map((key) => {
    const names = yaml.getIn([...keyPath, "filters", key]);
    return names instanceof YAMLSeq ? (names.toJSON() as string[]) : [];
  });
  // The lists are only rewritten when they change, which keeps their formatting otherwise.
  if ([...pinned, ...required].some((name) => fields.has(name))) {
    setPinnedAndRequiredFilters(
      yaml,
      pinned.filter((name) => !fields.has(name)),
      required.filter((name) => !fields.has(name)),
      keyPath,
    );
  }

  const defaultFilterPath = [...keyPath, "defaults", "filters"];
  if (
    yaml.hasIn(defaultFilterPath) &&
    [...getFilterFieldNames(defaultFilter)].some((name) => fields.has(name))
  ) {
    yaml.deleteIn(defaultFilterPath);
  }
}

/**
 * Names of the fields that a filter references, including the ones in the subquery of a measure filter.
 */
function getFilterFieldNames(
  expr: V1Expression | undefined,
  names = new Set<string>(),
) {
  if (!expr) return names;

  if (expr.ident) names.add(expr.ident);
  expr.cond?.exprs?.forEach((operand) => getFilterFieldNames(operand, names));
  if (expr.subquery) {
    if (expr.subquery.dimension) names.add(expr.subquery.dimension);
    expr.subquery.measures?.forEach((measure) => names.add(measure));
    getFilterFieldNames(expr.subquery.where, names);
    getFilterFieldNames(expr.subquery.having, names);
  }
  return names;
}
