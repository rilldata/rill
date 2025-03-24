import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

export interface FilterSpecStore {
  measures: Readable<MetricsViewSpecMeasureV2[]>;
  dimensions: Readable<MetricsViewSpecDimensionV2[]>;
}

export class ExploreSpecStore implements FilterSpecStore {
  measures: Readable<MetricsViewSpecMeasureV2[]>;
  dimensions: Readable<MetricsViewSpecDimensionV2[]>;

  public constructor(instanceId: string, exploreName: string) {
    const spec = useExploreValidSpec(instanceId, exploreName);

    this.measures = derived(spec, (spec) => {
      return spec.data?.metricsView?.measures ?? [];
    });
    this.dimensions = derived(spec, (spec) => {
      return spec.data?.metricsView?.dimensions ?? [];
    });
  }
}
