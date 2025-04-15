import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Validates various fields in explore state. Correct any invalid state.
 * Currently, it acts on only a small section of the state.
 *
 * TODO: move all validations from convertUrlParamsToPreset and AdvancedMeasureCorrector here
 */
export function correctExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const correctedExploreState = { ...exploreState };
  const errors: Error[] = [];

  const measures = getMapFromArray(
    metricsViewSpec.measures?.filter((m) =>
      exploreSpec.measures?.includes(m.name!),
    ) ?? [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsViewSpec.dimensions?.filter((d) =>
      exploreSpec.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  const { correctedExploreViewState, errors: errorsFromExploreView } =
    correctExploreViewState(measures, dimensions, exploreSpec, exploreState);
  Object.assign(correctedExploreState, correctedExploreViewState);
  errors.push(...errorsFromExploreView);

  return { correctedExploreState, errors };
}

function correctExploreViewState(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const errors: Error[] = [];
  const correctedExploreViewState: Partial<MetricsExplorerEntity> = {};

  let visibleMeasures = new Set(measures.keys());

  if (exploreState.visibleMeasures) {
    const selectedMeasures = exploreState.visibleMeasures.filter((m) =>
      measures.has(m),
    );
    const missingMeasures = getMissingValues(
      selectedMeasures,
      exploreState.visibleMeasures,
    );
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }

    // add the 1st measure if there is no valid measure
    if (selectedMeasures.length === 0 && exploreSpec.measures?.length) {
      selectedMeasures.push(exploreSpec.measures[0]);
    }

    correctedExploreViewState.allMeasuresVisible =
      selectedMeasures.length === exploreSpec.measures?.length;
    correctedExploreViewState.visibleMeasures = [...selectedMeasures];
    visibleMeasures = new Set(selectedMeasures);
  }

  if (exploreState.visibleDimensions) {
    const selectedDimensions = exploreState.visibleDimensions.filter((d) =>
      dimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      exploreState.visibleDimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    // add the 1st dimension if there is no valid dimension
    if (selectedDimensions.length === 0 && exploreSpec.dimensions?.length) {
      selectedDimensions.push(exploreSpec.dimensions[0]);
    }

    correctedExploreViewState.allDimensionsVisible =
      selectedDimensions.length === exploreSpec.dimensions?.length;
    correctedExploreViewState.visibleDimensions = [...selectedDimensions];
  }

  if (exploreState.leaderboardSortByMeasureName) {
    if (!visibleMeasures.has(exploreState.leaderboardSortByMeasureName)) {
      errors.push(
        getSingleFieldError(
          "sort by measure",
          exploreState.leaderboardSortByMeasureName,
        ),
      );
      correctedExploreViewState.leaderboardSortByMeasureName =
        correctedExploreViewState.visibleMeasures?.[0];
    } else {
      correctedExploreViewState.leaderboardSortByMeasureName =
        exploreState.leaderboardSortByMeasureName;
    }
  }

  if (exploreState.leaderboardMeasureNames?.length) {
    const selectedLeaderboardMeasures =
      exploreState.leaderboardMeasureNames.filter((m) =>
        visibleMeasures.has(m),
      );
    const missingLeaderboardMeasures = getMissingValues(
      selectedLeaderboardMeasures,
      exploreState.leaderboardMeasureNames,
    );
    if (missingLeaderboardMeasures.length) {
      errors.push(
        getMultiFieldError("leaderboard measure", missingLeaderboardMeasures),
      );
    }

    // If there are no valid leaderboard measure then set to leaderboard sort measure
    if (
      selectedLeaderboardMeasures.length === 0 &&
      correctedExploreViewState.leaderboardSortByMeasureName
    ) {
      selectedLeaderboardMeasures.push(
        correctedExploreViewState.leaderboardSortByMeasureName,
      );
    }

    correctedExploreViewState.leaderboardMeasureNames = [
      ...selectedLeaderboardMeasures,
    ];
  }

  // TODO: more validation once we need the full suite of validation
  return { correctedExploreViewState, errors };
}
