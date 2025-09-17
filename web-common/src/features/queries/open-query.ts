import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { generateExploreLink } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import {
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceGetExplore,
  runtimeServiceListResources,
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type {
  Expression,
  Measure,
  Schema as MetricsResolverQuery,
  Sort,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect } from "@sveltejs/kit";
import { get } from "svelte/store";
import { validateQuery } from "./query-types";

export async function openQuery({
  url,
  organization,
  project,
}: {
  url: URL;
  organization?: string;
  project?: string;
}) {
  let exploreURL: string;

  try {
    const queryParams = url.searchParams;

    // Get the JSON-encoded query parameters
    const queryJSON = queryParams.get("query");
    if (!queryJSON) {
      throw new Error("query parameter is required");
    }

    // Parse and validate the query with proper type safety
    let query: MetricsResolverQuery;
    try {
      const rawQuery = JSON.parse(queryJSON);
      query = validateQuery(rawQuery);
    } catch (e) {
      throw new Error(`Invalid query: ${e.message}`);
    }

    // Extract metrics view name (now type-safe)
    const metricsViewName = query.metrics_view;
    if (!metricsViewName) {
      throw new Error("metrics_view is required in query");
    }

    // Find an explore dashboard that uses this metrics view
    const exploreName = await findExploreForMetricsView(metricsViewName);

    // Convert query to ExploreState
    const exploreState = await convertQueryToExploreState(query, exploreName);

    // Generate the final explore URL
    exploreURL = await generateExploreLink(
      exploreState,
      exploreName,
      organization,
      project,
    );
    console.log("exploreURL", exploreURL);
  } catch (e) {
    console.error("Failed to process open-query:", e);

    // Use SvelteKit's error handling instead of manual redirect
    throw error(
      400,
      e.message || "Unable to open a dashboard that represents this query",
    );
  }

  // Redirect outside the try/catch since redirect() throws internally
  redirect(302, exploreURL);
}

/**
 * Find an explore dashboard that uses the given metrics view.
 * This mirrors the backend's findExploreForMetricsView logic.
 */
async function findExploreForMetricsView(
  metricsViewName: string,
): Promise<string> {
  const instanceId = get(runtime).instanceId;
  console.log("instanceId", instanceId);

  // List all explore resources
  const exploreResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: ResourceKind.Explore,
    }),
    queryFn: ({ signal }) =>
      runtimeServiceListResources(
        instanceId,
        { kind: ResourceKind.Explore },
        signal,
      ),
  });

  // Look for an explore that references this metrics view
  if (exploreResources.resources) {
    for (const resource of exploreResources.resources) {
      if (resource.explore?.state?.validSpec?.metricsView === metricsViewName) {
        return resource.meta?.name?.name || "";
      }
    }
  }

  // If no explore found, throw an error
  throw new Error(
    `No explore dashboard found for metrics view: ${metricsViewName}`,
  );
}

/**
 * Convert Query directly to ExploreState without going through URL parameters.
 */
async function convertQueryToExploreState(
  query: MetricsResolverQuery,
  exploreName: string,
): Promise<Partial<ExploreState>> {
  const instanceId = get(runtime).instanceId;

  // Get explore and metrics view specs
  const getExploreResponse = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetExploreQueryKey(instanceId, {
      name: exploreName,
    }),
    queryFn: ({ signal }) =>
      runtimeServiceGetExplore(
        instanceId,
        {
          name: exploreName,
        },
        signal,
      ),
  });
  const exploreResource = getExploreResponse.explore;
  const metricsViewResource = getExploreResponse.metricsView;

  if (!exploreResource?.explore?.state?.validSpec) {
    throw new Error("Could not load explore specification");
  }

  if (!metricsViewResource?.metricsView?.state?.validSpec) {
    throw new Error("Could not load metrics view specification");
  }

  const exploreSpec = exploreResource.explore.state.validSpec;
  const metricsViewSpec = metricsViewResource?.metricsView?.state?.validSpec;

  // Build partial ExploreState directly from Query
  const partialExploreState: Partial<ExploreState> = {};

  // Convert dimensions
  if (query.dimensions && Array.isArray(query.dimensions)) {
    const dimensionNames = query.dimensions.map((d) => d.name).filter(Boolean);

    // Validate dimensions exist in the metrics view
    const validDimensions = dimensionNames.filter(
      (name) =>
        metricsViewSpec.dimensions?.some((d) => d.name === name) &&
        exploreSpec.dimensions?.includes(name),
    );

    if (validDimensions.length > 0) {
      partialExploreState.visibleDimensions = validDimensions;
      partialExploreState.allDimensionsVisible =
        validDimensions.length === exploreSpec.dimensions?.length;
    }
  }

  // Convert measures
  if (query.measures && Array.isArray(query.measures)) {
    const measureNames = query.measures.map((m) => m.name).filter(Boolean);

    // Validate measures exist in the metrics view
    const validMeasures = measureNames.filter(
      (name) =>
        metricsViewSpec.measures?.some((m) => m.name === name) &&
        exploreSpec.measures?.includes(name),
    );

    if (validMeasures.length > 0) {
      partialExploreState.visibleMeasures = validMeasures;
      partialExploreState.allMeasuresVisible =
        validMeasures.length === exploreSpec.measures?.length;
    }
  }

  partialExploreState.selectedTimeRange =
    mapResolverTimeRangeToDashboardControls(query.time_range);
  if (query.comparison_time_range) {
    partialExploreState.selectedComparisonTimeRange =
      mapResolverTimeRangeToDashboardControls(query.comparison_time_range);
    partialExploreState.showTimeComparison = true;
  }

  // Convert where filter
  partialExploreState.whereFilter = mapResolverExpressionToV1Expression(
    query.where,
  );

  // Convert sort
  if (query.sort) {
    mapSort(query.measures ?? [], query.sort, partialExploreState);
  }

  // Set default timezone if not specified
  if (query.time_zone) {
    partialExploreState.selectedTimezone = query.time_zone;
  }

  return partialExploreState;
}

function mapResolverTimeRangeToDashboardControls(
  timeRange: TimeRange | undefined,
): DashboardTimeControls {
  // Default to "All Time" when no time range is specified
  if (!timeRange)
    return { name: TimeRangePreset.ALL_TIME } as DashboardTimeControls;

  if (timeRange.start && timeRange.end) {
    return {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (timeRange.expression) {
    return {
      name: timeRange.expression,
    } as DashboardTimeControls;
  } else if (timeRange.iso_duration) {
    return {
      name: timeRange.iso_duration,
    } as DashboardTimeControls;
  }

  // Fallback to all-time
  return { name: TimeRangePreset.ALL_TIME } as DashboardTimeControls;
}

const OperationMap: Record<string, V1Operation> = {
  "": V1Operation.OPERATION_UNSPECIFIED,
  eq: V1Operation.OPERATION_UNSPECIFIED,
  neq: V1Operation.OPERATION_UNSPECIFIED,
  lt: V1Operation.OPERATION_UNSPECIFIED,
  lte: V1Operation.OPERATION_UNSPECIFIED,
  gt: V1Operation.OPERATION_UNSPECIFIED,
  gte: V1Operation.OPERATION_UNSPECIFIED,
  in: V1Operation.OPERATION_UNSPECIFIED,
  nin: V1Operation.OPERATION_UNSPECIFIED,
  ilike: V1Operation.OPERATION_UNSPECIFIED,
  nilike: V1Operation.OPERATION_UNSPECIFIED,
  or: V1Operation.OPERATION_UNSPECIFIED,
  and: V1Operation.OPERATION_UNSPECIFIED,
};
function mapResolverExpressionToV1Expression(
  expr: Expression | undefined,
): V1Expression | undefined {
  if (!expr) return undefined;

  if (expr.name) {
    return { ident: expr.name };
  }

  if (expr.value) {
    return { val: expr.value };
  }

  if (expr.cond) {
    return {
      cond: {
        op: OperationMap[expr.cond.op] || V1Operation.OPERATION_UNSPECIFIED,
        exprs: expr.cond.exprs?.map(mapResolverExpressionToV1Expression),
      },
    };
  }

  if (expr.subquery) {
    return {
      subquery: {
        dimension: expr.subquery.dimension.name,
        measures: expr.subquery.measures.map((m) => m.name),
        where: mapResolverExpressionToV1Expression(expr.subquery.where),
        having: mapResolverExpressionToV1Expression(expr.subquery.having),
      },
    };
  }

  return {};
}

function mapSort(
  measures: Measure[],
  sort: Sort[] | undefined,
  partialExploreState: Partial<ExploreState>,
) {
  if (!sort?.length) return;
  const sortField = sort[0];
  const measure = measures.find((m) => m.name === sortField.name);
  if (!measure) return;
  const { name, type } = getMeasureNameAndType(measure);
  partialExploreState.leaderboardSortByMeasureName = name;
  partialExploreState.sortDirection = sortField.desc
    ? SortDirection.DESCENDING
    : SortDirection.ASCENDING;
  partialExploreState.dashboardSortType = type;
}
function getMeasureNameAndType(measure: Measure) {
  if (measure.compute?.comparison_delta?.measure) {
    return {
      name: measure.compute.comparison_delta.measure,
      type: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
    };
  }

  if (measure.compute?.comparison_ratio?.measure) {
    return {
      name: measure.compute.comparison_ratio.measure,
      type: DashboardState_LeaderboardSortType.DELTA_PERCENT,
    };
  }

  if (measure.compute?.percent_of_total?.measure) {
    return {
      name: measure.compute.percent_of_total.measure,
      type: DashboardState_LeaderboardSortType.PERCENT,
    };
  }

  return {
    name: measure.name,
    type: DashboardState_LeaderboardSortType.VALUE,
  };
}
