import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type {
  ComponentTransformerProperties,
  TransformerArgs,
} from "@rilldata/web-common/features/explore-mappers/types";

export function getDashboardFromComponentTransformer({
  queryClient: _queryClient,
  instanceId: _instanceId,
  req,
  dashboard,
  timeRangeSummary: _timeRangeSummary,
  executionTime: _executionTime,
  metricsView: _metricsView,
  explore: _explore,
  annotations: _annotations,
}: TransformerArgs<ComponentTransformerProperties>) {
  const combinedDashboardState: ExploreState = { ...dashboard, ...req };

  return Promise.resolve(combinedDashboardState);
}
