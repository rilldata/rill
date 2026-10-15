import type { PartialMessage, Timestamp } from "@bufbuild/protobuf";
import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import {
  generateExploreLink,
  getExploreSpecs,
} from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
  runtimeServiceQueryResolver,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryServiceConvertMetricsSQLToQuery } from "@rilldata/web-common/runtime-client/v2/gen/query-service";

export type OpenMetricsSQLArgs = {
  client: RuntimeClient;
  // The metrics SQL of the alert or report.
  sql: string;
  // The explore the alert or report names, if any. Without one, the only explore on the
  // query's metrics view: the runtime grants an emailed link's token that same explore.
  exploreName?: string;
  // The run the link was sent for: now() and rilltime expressions resolve against it.
  executionTime?: string | null;
  // The magic token of an emailed link. Its row filter already applies the query's WHERE.
  token?: string | null;
  organization: string;
  project: string;
};

/**
 * Builds the explore URL that shows the rows an alert or report written in metrics SQL saw.
 *
 * It throws when the explore can't show the query faithfully, rather than opening it
 * with fewer filters than the alert or report applied: callers fall back to the
 * alert or report page.
 */
export async function getExploreUrlForMetricsSQL({
  client,
  sql,
  exploreName,
  executionTime,
  token,
  organization,
  project,
}: OpenMetricsSQLArgs): Promise<string> {
  const { query: compiled } = await queryServiceConvertMetricsSQLToQuery(
    client,
    {
      sql,
      // The request is decoded from JSON, where a Timestamp is an RFC 3339 string.
      executionTime: (executionTime ?? undefined) as unknown as
        | PartialMessage<Timestamp>
        | undefined,
    },
  );
  const query = compiled as unknown as MetricsResolverQuery | undefined;
  if (!query?.metrics_view) {
    throw new Error("The metrics SQL did not compile to a metrics view query");
  }

  exploreName ||= await findOnlyExploreForMetricsView(
    client,
    query.metrics_view,
  );
  const { metricsViewSpec, exploreSpec } = await getExploreSpecs(
    client,
    exploreName,
  );
  if (exploreSpec.metricsView !== query.metrics_view) {
    throw new Error(
      `Explore "${exploreName}" does not show metrics view "${query.metrics_view}"`,
    );
  }
  const exploreState = mapMetricsResolverQueryToDashboard(
    metricsViewSpec,
    exploreSpec,
    { query, ignoreWhere: !!token },
  );
  if (query.having) {
    exploreState.whereFilter =
      mergeFilters(
        exploreState.whereFilter ?? createAndExpression([]),
        createAndExpression([
          await havingAsInFilter(client, {
            sql,
            query,
            executionTime,
            exploreDimensions: exploreSpec.dimensions ?? [],
          }),
        ]),
      ) ?? createAndExpression([]);
  }

  const url = new URL(
    await generateExploreLink(
      client,
      exploreState,
      metricsViewSpec,
      exploreSpec,
      exploreName,
      organization,
      project,
    ),
  );
  if (token) {
    url.pathname = `/${organization}/${project}/-/share/${token}/explore/${encodeURIComponent(exploreName)}`;
  }
  return url.toString();
}

async function findOnlyExploreForMetricsView(
  client: RuntimeClient,
  metricsViewName: string,
): Promise<string> {
  const request = { kind: ResourceKind.Explore };
  const { resources } = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(
      client.instanceId,
      request,
    ),
    queryFn: ({ signal }) =>
      runtimeServiceListResources(client, request, { signal }),
  });
  const explores = (resources ?? []).filter(
    (r) => r.explore?.state?.validSpec?.metricsView === metricsViewName,
  );
  if (explores.length !== 1 || !explores[0].meta?.name?.name) {
    throw new Error(
      `Expected exactly one explore on metrics view "${metricsViewName}", found ${explores.length}`,
    );
  }
  return explores[0].meta.name.name;
}

/**
 * A top-level HAVING as the dimension values that pass it, found by running the query.
 *
 * The explore's measure filter looks like a HAVING but isn't one: it computes the measure
 * over the whole time range, ignoring the other filters, so "category with a return rate
 * over 15% in the last month" would become "over 15% overall". Running the query keeps
 * the WHERE in the computation, the way the alert or report saw it.
 */
async function havingAsInFilter(
  client: RuntimeClient,
  {
    sql,
    query,
    executionTime,
    exploreDimensions,
  }: {
    sql: string;
    query: MetricsResolverQuery;
    executionTime: string | null | undefined;
    exploreDimensions: string[];
  },
): Promise<V1Expression> {
  // One plain dimension the explore shows: a filter on its values is then the same rows.
  const dimension =
    query.dimensions?.length === 1 ? query.dimensions[0] : undefined;
  if (
    !dimension?.name ||
    dimension.compute ||
    !exploreDimensions.includes(dimension.name)
  ) {
    throw new Error(
      "The explore can only show a HAVING on one dimension it shows",
    );
  }
  const { data } = await runtimeServiceQueryResolver(client, {
    resolver: "metrics_sql",
    resolverProperties: { sql },
    resolverArgs: executionTime ? { execution_time: executionTime } : undefined,
  });
  const name = dimension.name;
  const values = (data ?? []).map((row) => {
    if (!(name in row)) {
      throw new Error(`The query's rows have no "${name}" column`);
    }
    return row[name];
  });
  return createInExpression(name, values);
}
