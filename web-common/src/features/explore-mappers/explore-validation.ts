import { useGetExploresForMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";
import type {
  DashboardSelectionCriteria,
  ExploreAvailabilityResult,
  ExploreLinkError,
} from "./types";

/**
 * Validates if explore dashboards are available for the given metrics view
 */
export function useExploreAvailability(
  instanceId: string,
  metricsViewName: string,
): Readable<ExploreAvailabilityResult> {
  const exploresQuery = useGetExploresForMetricsView(
    instanceId,
    metricsViewName,
  );

  return derived(exploresQuery, (data) => {
    if (data.error) {
      return {
        isAvailable: false,
        error: `Failed to fetch explores: ${data.error.message}`,
      };
    }
    if (!data.data || data.data.length === 0) {
      return {
        isAvailable: false,
        error: "No explore dashboards found for this metrics view",
      };
    }

    // Use the best available dashboard
    const selectedDashboard = selectBestDashboard(data.data);

    return {
      isAvailable: true,
      exploreName: selectedDashboard?.meta?.name?.name,
      displayName:
        selectedDashboard?.explore?.spec?.displayName ||
        selectedDashboard?.explore?.state?.validSpec?.displayName,
    };
  });
}

/**
 * Selects the best dashboard from available options based on criteria
 */
export function selectBestDashboard(
  dashboards: V1Resource[],
  criteria: DashboardSelectionCriteria = { preferredType: "first_available" },
): V1Resource | null {
  if (!dashboards || dashboards.length === 0) {
    return null;
  }

  // Filter for valid explores only
  const validDashboards = dashboards.filter(
    (dashboard) => !!dashboard.explore?.state?.validSpec,
  );

  if (validDashboards.length === 0) {
    return null;
  }

  switch (criteria.preferredType) {
    case "recent":
      // Sort by most recent update time if available
      return validDashboards.sort((a, b) => {
        const aTime = new Date(
          a.explore?.state?.dataRefreshedOn || 0,
        ).getTime();
        const bTime = new Date(
          b.explore?.state?.dataRefreshedOn || 0,
        ).getTime();
        return bTime - aTime;
      })[0];

    case "most_used":
      return validDashboards[0];

    case "first_available":
    default:
      return validDashboards[0];
  }
}

/**
 * Creates a standardized error for the linking system
 */
export function createLinkError(
  type: ExploreLinkError["type"],
  message: string,
  details?: any,
): ExploreLinkError {
  return {
    type,
    message,
    details,
  };
}

/**
 * Validates user permissions for accessing the explore dashboard
 */
export function validateUserPermissions(): boolean {
  // TODO: Implement permission checking logic
  return true;
}
