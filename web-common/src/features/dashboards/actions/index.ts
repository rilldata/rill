import { get } from "svelte/store";
import type { StateManagers } from "../state-managers/state-managers";
import { cancelDashboardQueries } from "../dashboard-queries";
import { removeIfExists } from "@rilldata/web-common/lib/arrayUtils";

export function clearFilterForDimension(
  ctx: StateManagers,
  dimensionId: string,
  include: boolean,
) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  ctx.updateDashboard((dashboard) => {
    if (include) {
      removeIfExists(
        dashboard.filters.include,
        (dimensionValues) => dimensionValues.name === dimensionId,
      );
    } else {
      removeIfExists(
        dashboard.filters.exclude,
        (dimensionValues) => dimensionValues.name === dimensionId,
      );
    }
    dashboard.pinIndex = -1;
  });
}

export function clearAllFilters(ctx: StateManagers) {
  const filters = get(ctx.dashboardStore).filters;
  const hasFilters =
    (filters && filters.include.length > 0) || filters.exclude.length > 0;
  const metricViewName = get(ctx.metricsViewName);
  if (hasFilters) {
    cancelDashboardQueries(ctx.queryClient, metricViewName);
    ctx.updateDashboard((dashboard) => {
      dashboard.filters.include = [];
      dashboard.filters.exclude = [];
      dashboard.dimensionFilterExcludeMode.clear();
      dashboard.pinIndex = -1;
    });
  }
}

export function toggleDimensionValue(
  ctx: StateManagers,
  dimensionName: string,
  dimensionValue: string,
) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);

  ctx.updateDashboard((dashboard) => {
    const relevantFilterKey = dashboard.dimensionFilterExcludeMode.get(
      dimensionName,
    )
      ? "exclude"
      : "include";

    const dimensionEntryIndex = dashboard.filters[relevantFilterKey].findIndex(
      (filter) => filter.name === dimensionName,
    );

    if (dimensionEntryIndex >= 0) {
      const index = dashboard.filters[relevantFilterKey][
        dimensionEntryIndex
      ].in?.findIndex((value) => value === dimensionValue) as number;

      if (index >= 0) {
        dashboard.filters[relevantFilterKey][dimensionEntryIndex].in?.splice(
          index,
          1,
        );

        // Only decrement pinIndex if the removed value was before the pinned value
        if (dashboard.pinIndex >= index) {
          dashboard.pinIndex--;
        }

        return;
      }

      dashboard.filters[relevantFilterKey][dimensionEntryIndex].in.push(
        dimensionValue,
      );
    } else {
      dashboard.filters[relevantFilterKey].push({
        name: dimensionName,
        in: [dimensionValue],
      });
    }
  });
}

export function toggleFilterMode(ctx: StateManagers, dimensionName: string) {
  const metricViewName = get(ctx.metricsViewName);
  cancelDashboardQueries(ctx.queryClient, metricViewName);
  ctx.updateDashboard((dashboard) => {
    const exclude = dashboard.dimensionFilterExcludeMode.get(dimensionName);
    dashboard.dimensionFilterExcludeMode.set(dimensionName, !exclude);

    const relevantFilterKey = exclude ? "exclude" : "include";
    const otherFilterKey = exclude ? "include" : "exclude";

    const otherFilterEntryIndex = dashboard.filters[
      relevantFilterKey
    ].findIndex((filter) => filter.name === dimensionName);
    // if relevant filter is not present then return
    if (otherFilterEntryIndex === -1) return;

    // push relevant filters to other filter
    dashboard.filters[otherFilterKey].push(
      dashboard.filters[relevantFilterKey][otherFilterEntryIndex],
    );
    // remove entry from relevant filter
    dashboard.filters[relevantFilterKey].splice(otherFilterEntryIndex, 1);
  });
}
