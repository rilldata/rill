import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";

export const { singleSelector: selectMetricsExplorerById } =
  generateEntitySelectors<MetricsExplorerEntity, "metricsExplorer">(
    "metricsExplorer"
  );
