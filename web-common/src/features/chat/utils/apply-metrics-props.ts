import { applyMetricsPropsToDashboard, findExploreForMetricsView } from "@rilldata/web-common/features/explore-mappers/apply-metrics-props";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * Handle applying metricsProps from a chat tool result to the dashboard.
 * This function extracts the metricsProps from the tool result, finds the appropriate
 * explore dashboard, and applies the state.
 */
export async function handleApplyMetricsProps(
  metricsProps: MetricsResolverQuery,
  toolCall: any,
): Promise<void> {
  try {
    console.log("Starting applyMetricsProps with:", { metricsProps, toolCall });

    // Extract metrics view name from the metricsProps
    const metricsViewName = metricsProps.metrics_view;
    if (!metricsViewName) {
      throw new Error("No metrics_view found in metricsProps");
    }

    console.log("Found metrics view name:", metricsViewName);

    // Find the explore dashboard that uses this metrics view
    const exploreName = await findExploreForMetricsView(metricsViewName);
    console.log("Found explore name:", exploreName);

    // Convert metricsProps to ExploreState
    const { partialExploreState, metricsViewSpec } = await applyMetricsPropsToDashboard(
      metricsProps,
      exploreName,
    );
    console.log("Generated partial explore state:", partialExploreState);
    console.log("Retrieved metrics view spec:", metricsViewSpec);

    // Apply the state to the dashboard store
    metricsExplorerStore.mergePartialExplorerEntity(
      exploreName,
      partialExploreState,
      metricsViewSpec,
    );

    console.log("Successfully applied metricsProps to dashboard:", {
      exploreName,
      partialExploreState,
    });
  } catch (error) {
    console.error("Failed to apply metricsProps to dashboard:", error);
    // Re-throw with more context
    throw new Error(`Failed to apply metrics props: ${error instanceof Error ? error.message : String(error)}`);
  }
}
