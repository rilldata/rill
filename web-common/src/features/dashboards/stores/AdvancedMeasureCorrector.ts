import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  queryGroupsByRequiredDimensions,
  type QueryGroupBy,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

/**
 * Single use class to correct incorrect use of advanced measures.
 * Reason to use a class here is to avoid too many arguments in methods (especially measureIsValidForComponent).
 * NOTE: this doesnt have to deal with V1ExploreSpec since it is assumed measures/dimensions not present are already removed
 * TODO: this should not be necessary once we use V1ExplorePreset for everything
 */
export class AdvancedMeasureCorrector {
  private measuresMap: Map<string, MetricsViewSpecMeasure>;
  private measuresGrains: Map<string, V1TimeGrain>;
  private ephemeralMeasureNames: Set<string>;

  private constructor(
    private readonly exploreState: ExploreState,
    private readonly metricsViewSpec: V1MetricsViewSpec,
  ) {
    this.ephemeralMeasureNames = new Set(
      exploreState.ephemeralMeasures?.map((def) => def.name) ?? [],
    );
    this.measuresMap = getMapFromArray(
      metricsViewSpec.measures ?? [],
      (m) => m.name ?? "",
    );
    this.measuresGrains = getMapFromArray(
      metricsViewSpec.measures ?? [],
      (m) => m.name ?? "",
      (m) => {
        const d = m.requiredDimensions?.find(
          (d) =>
            d.timeGrain && d.timeGrain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
        );
        return d?.timeGrain ?? V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
      },
    );
  }

  public static correct(
    exploreState: ExploreState,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    new AdvancedMeasureCorrector(exploreState, metricsViewSpec).correct();
  }

  private correct() {
    // Filters are owned by ExpressionFilterManager, which corrects them itself.
    this.correctLeaderboards();
    this.correctTimeDimensionDetails();
    this.correctPivot();
  }

  private correctLeaderboards() {
    // Leaderboards group by each visible dimension in turn and also query totals,
    // so they cannot query measures with required dimensions.
    // The dimension table shares their measures and sort but groups only by its dimension,
    // so while it is open, measures that require only that dimension are valid too.
    const groupBy: QueryGroupBy =
      this.exploreState.activePage ===
        DashboardState_ActivePage.DIMENSION_TABLE &&
      this.exploreState.selectedDimensionName
        ? { dimensions: [this.exploreState.selectedDimensionName] }
        : {};
    const validLeaderboardMeasures = this.exploreState.visibleMeasures.filter(
      (m) => !this.measureIsValidForComponent(m, true, false, groupBy),
    );

    // Keep the sort measure when no visible measure is valid,
    // since ExploreState requires one.
    if (
      this.exploreState.leaderboardSortByMeasureName &&
      validLeaderboardMeasures.length > 0 &&
      !validLeaderboardMeasures.includes(
        this.exploreState.leaderboardSortByMeasureName,
      )
    ) {
      this.exploreState.leaderboardSortByMeasureName =
        validLeaderboardMeasures[0];
    }

    this.exploreState.leaderboardMeasureNames =
      this.exploreState.leaderboardMeasureNames.filter((lm) =>
        validLeaderboardMeasures.includes(lm),
      );
    if (
      this.exploreState.leaderboardMeasureNames.length === 0 &&
      this.exploreState.leaderboardSortByMeasureName
    ) {
      this.exploreState.leaderboardMeasureNames = [
        this.exploreState.leaderboardSortByMeasureName,
      ];
    }
  }

  private correctTimeDimensionDetails() {
    if (
      !this.measureIsValidForComponent(
        this.exploreState.tdd.expandedMeasureName ?? "",
        true,
        true,
        { time: true },
      )
    ) {
      return;
    }

    this.exploreState.tdd.expandedMeasureName = "";
    if (
      this.exploreState.activePage ===
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
    ) {
      this.exploreState.activePage = DashboardState_ActivePage.DEFAULT;
    }
  }

  private correctPivot() {
    this.exploreState.pivot.columns = this.exploreState.pivot.columns.filter(
      (m) =>
        m.type !== PivotChipType.Measure ||
        !this.measureIsValidForComponent(m.id, true, false, undefined),
    );
    this.exploreState.pivot.sorting = this.exploreState.pivot.sorting.filter(
      (s) =>
        !this.measuresMap.has(s.id) ||
        !this.measureIsValidForComponent(s.id, true, false, undefined),
    );
  }

  /**
   * Checks if a measure is valid for the component based on dashboard selections and component support.
   * Additional arguments indicate whether certain types of advanced measures are supported in the component or not.
   * groupBy is what the component's queries group by, used to check required dimensions;
   * undefined skips that check, for the pivot whose rows and columns the user picks.
   */
  private measureIsValidForComponent(
    measureName: string,
    supportsComparisonMeasure: boolean,
    supportsWindowedMeasure: boolean,
    groupBy: QueryGroupBy | undefined,
    supportsEphemeralMeasure = true,
  ) {
    // ephemeral measures are not in the metrics view spec;
    // they are validated separately when parsing the URL state.
    if (this.ephemeralMeasureNames.has(measureName)) {
      return !supportsEphemeralMeasure;
    }
    const measure = this.measuresMap.get(measureName);
    if (!measure) return true;
    const grain =
      this.measuresGrains.get(measureName) ??
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

    switch (true) {
      // selected grain and measure's grain mismatch
      case grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
        grain !== this.exploreState.selectedTimeRange?.interval:
        return true;

      // for comparison measures,
      // if the component supports it and no time comparison is enabled
      // or if the component does not support it
      case measure.type ===
        MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON &&
        ((supportsComparisonMeasure &&
          (!this.exploreState.showTimeComparison ||
            !this.exploreState.selectedComparisonTimeRange)) ||
          !supportsComparisonMeasure):
        return true;

      // for measures with window operations, if the component doesnt support it
      case !!measure.window && !supportsWindowedMeasure:
        return true;

      // for measures with required dimensions, if the component doesnt group by all of them
      case !!groupBy &&
        !queryGroupsByRequiredDimensions(
          this.exploreState,
          this.metricsViewSpec,
          measure,
          groupBy,
        ):
        return true;

      default:
        return false;
    }
  }
}
