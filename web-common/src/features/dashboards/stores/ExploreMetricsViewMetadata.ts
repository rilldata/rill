import {
  useMetricsViewTimeRange,
  useMetricsViewValidSpec,
} from "@rilldata/web-common/features/dashboards/selectors.ts";
import { isSimpleMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1ExploreTimeRange,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

type TimeDefaults = {
  timeRange: string | undefined;
  timeRanges: V1ExploreTimeRange[];
};

export class ExploreMetricsViewMetadata {
  public readonly metricsViewSpecQuery: ReturnType<
    typeof useMetricsViewValidSpec<V1MetricsViewSpec>
  >;
  public readonly timeDefaults: Readable<TimeDefaults>;
  public readonly timeRangeSummary: ReturnType<typeof useMetricsViewTimeRange>;

  public readonly allDimensions: Readable<MetricsViewSpecDimension[]>;
  public readonly dimensionNameMap: Readable<
    Map<string, MetricsViewSpecDimension>
  >;
  public readonly allSimpleMeasures: Readable<MetricsViewSpecMeasure[]>;
  public readonly measureNameMap: Readable<Map<string, MetricsViewSpecMeasure>>;

  public constructor(
    instanceId: string,
    public readonly metricsViewName: string,
    exploreName: string,
  ) {
    this.metricsViewSpecQuery = useMetricsViewValidSpec(
      instanceId,
      metricsViewName,
    );

    const validSpecQuery = useExploreValidSpec(
      instanceId,
      exploreName,
      undefined,
      queryClient,
    );
    this.timeDefaults = derived(validSpecQuery, ($validSpecQuery) => {
      const exploreSpec = $validSpecQuery.data?.explore;
      return {
        timeRange: exploreSpec?.defaultPreset?.timeRange,
        timeRanges: exploreSpec?.timeRanges ?? [],
      };
    });

    this.timeRangeSummary = useMetricsViewTimeRange(
      instanceId,
      metricsViewName,
      undefined,
      queryClient,
    );

    this.allDimensions = derived(
      [this.metricsViewSpecQuery, validSpecQuery],
      ([$metricsViewSpecQuery, $validSpecQuery]) => {
        const metricsViewSpec = $metricsViewSpecQuery.data;
        const dimensions = metricsViewSpec?.dimensions ?? [];
        if (!exploreName) return dimensions;

        const exploreSpec = $validSpecQuery.data?.explore ?? {};
        return dimensions.filter((d) =>
          exploreSpec.dimensions?.includes(d.name!),
        );
      },
    );
    this.dimensionNameMap = derived(this.allDimensions, ($allDimensions) =>
      getMapFromArray($allDimensions, (d) => d.name!),
    );

    this.allSimpleMeasures = derived(
      [this.metricsViewSpecQuery, validSpecQuery],
      ([$metricsViewSpecQuery, $validSpecQuery]) => {
        const metricsViewSpec = $metricsViewSpecQuery.data;
        const measures =
          metricsViewSpec?.measures?.filter(isSimpleMeasure) ?? [];
        if (!exploreName) return measures;

        const exploreSpec = $validSpecQuery.data?.explore ?? {};
        return measures.filter((m) => exploreSpec.measures?.includes(m.name!));
      },
    );
    this.measureNameMap = derived(
      this.allSimpleMeasures,
      ($allSimpleMeasures) =>
        getMapFromArray($allSimpleMeasures, (m) => m.name!),
    );
  }
}
