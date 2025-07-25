import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { generateExploreLink } from "@rilldata/web-common/features/explore-mappers/generate-explore-link";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import {
  getRuntimeServiceGetExploreQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceGetExplore,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect } from "@sveltejs/kit";
import { get } from "svelte/store";
import { validateQuery, type Query } from "./query-types";

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
    let query: Query;
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
  query: Query,
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

  // Get time range summary for default "All Time" range
  let timeRangeSummary: any = undefined;
  if (metricsViewSpec?.timeDimension && !query.time_range?.start) {
    try {
      const {
        getQueryServiceMetricsViewTimeRangeQueryKey,
        queryServiceMetricsViewTimeRange,
      } = await import("@rilldata/web-common/runtime-client");
      const timeRangeResponse = await queryClient.fetchQuery({
        queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
          instanceId,
          query.metrics_view,
          {},
        ),
        queryFn: () =>
          queryServiceMetricsViewTimeRange(instanceId, query.metrics_view, {}),
      });
      timeRangeSummary = timeRangeResponse?.timeRangeSummary;
    } catch (e) {
      console.warn("Failed to fetch time range summary:", e);
    }
  }

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

  // Convert time range
  if (query.time_range?.start && query.time_range?.end) {
    partialExploreState.selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(query.time_range.start),
      end: new Date(query.time_range.end),
    };
  } else if (timeRangeSummary?.min && timeRangeSummary?.max) {
    // Default to "All Time" when no time range is specified
    partialExploreState.selectedTimeRange = {
      name: TimeRangePreset.ALL_TIME,
      start: new Date(timeRangeSummary.min),
      end: new Date(timeRangeSummary.max),
    };
  }

  // Convert where filter
  if (query.where) {
    partialExploreState.whereFilter = query.where;
  }

  // Convert sort
  if (query.sort && Array.isArray(query.sort) && query.sort.length > 0) {
    const sortField = query.sort[0];
    if (sortField.name) {
      // Validate the sort field is a valid measure
      const isValidMeasure = metricsViewSpec.measures?.some(
        (m) => m.name === sortField.name,
      );
      if (isValidMeasure) {
        partialExploreState.leaderboardSortByMeasureName = sortField.name;
        partialExploreState.sortDirection = sortField.desc
          ? SortDirection.DESCENDING
          : SortDirection.ASCENDING;
      }
    }
  }

  // Set default timezone if not specified
  if (query.time_zone) {
    partialExploreState.selectedTimezone = query.time_zone;
  }

  return partialExploreState;
}
