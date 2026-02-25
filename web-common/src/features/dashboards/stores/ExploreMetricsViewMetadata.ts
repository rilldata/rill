import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { isSimpleMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

export class ExploreMetricsViewMetadata {
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly timeRangeSummary: ReturnType<typeof useMetricsViewTimeRange>;

  public readonly allDimensions: Readable<MetricsViewSpecDimension[]>;
  public readonly dimensionNameMap: Readable<
    Map<string, MetricsViewSpecDimension>
  >;
  public readonly allSimpleMeasures: Readable<MetricsViewSpecMeasure[]>;
  public readonly measureNameMap: Readable<Map<string, MetricsViewSpecMeasure>>;

  public constructor(
    client: RuntimeClient,
    public readonly metricsViewName: string,
    exploreName: string,
  ) {
    this.validSpecQuery = useExploreValidSpec(
      client.instanceId,
      exploreName,
      undefined,
      queryClient,
    );
    this.timeRangeSummary = useMetricsViewTimeRange(
      client,
      metricsViewName,
      undefined,
      queryClient,
    );

    this.allDimensions = derived(this.validSpecQuery, ($validSpecResp) => {
      const metricsViewSpec = $validSpecResp.data?.metricsView;
      return metricsViewSpec?.dimensions ?? [];
    });
    this.dimensionNameMap = derived(this.allDimensions, ($allDimensions) =>
      getMapFromArray($allDimensions, (d) => d.name!),
    );

    this.allSimpleMeasures = derived(this.validSpecQuery, ($validSpecResp) => {
      const metricsViewSpec = $validSpecResp.data?.metricsView;
      return metricsViewSpec?.measures?.filter(isSimpleMeasure) ?? [];
    });
    this.measureNameMap = derived(
      this.allSimpleMeasures,
      ($allSimpleMeasures) =>
        getMapFromArray($allSimpleMeasures, (m) => m.name!),
    );
  }
}
