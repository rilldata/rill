import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
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
 * Validates various fields in explore state.
 * Removes any invalid state. Cascading merge should fill in remaining state.
 *
 * Note: Since this method will keep the state as is most of the time,
 * it is more performant to do an in place modification of exploreState, rather than creating a copy.
 * TODO: Look into doing this and cascading merge in a single place for performance when we have isolated states like FilterState etc
 *
 * Currently, it only acts on only a small section of the state.
 *
 * TODO: move all validations from convertUrlParamsToPreset and AdvancedMeasureCorrector here
 */
export function validateAndCleanExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<ExploreState>,
) {
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

  const errorsFromExploreView = validateAndCleanExploreViewState(
    measures,
    dimensions,
    exploreSpec,
    exploreState,
  );
  errors.push(...errorsFromExploreView);

  return errors;
}

/**
 * Looks at any invalid fields in explore view.
 * Cleans up any invalid fields.
 *
 * Note: Since this method will keep the state as is most of the time,
 * it is more performant to do an in place modification of exploreState, rather than creating a copy.
 */
function validateAndCleanExploreViewState(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<ExploreState>,
) {
  const errors: Error[] = [];

  if (exploreState.visibleDimensions) {
    const selectedDimensions = exploreState.visibleDimensions.filter(
      (d) =>
        dimensions.has(d) && dimensions.get(d)?.type !== "DIMENSION_TYPE_TIME",
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      exploreState.visibleDimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    if (selectedDimensions.length > 0) {
      // If there are any remaining valid dimensions then set it.
      exploreState.allDimensionsVisible =
        selectedDimensions.length === exploreSpec.dimensions?.length;
      exploreState.visibleDimensions = selectedDimensions;
    } else {
      // Else remove the relevant fields so that cascading merge can set fields from other sources.
      delete exploreState.allDimensionsVisible;
      delete exploreState.visibleDimensions;
    }
  }

  // TODO: more validation once we need the full suite of validation
  return [
    ...errors,
    ...validateAndCleanMeasureRelatedExploreState(
      measures,
      exploreSpec,
      exploreState,
    ),
  ];
}

/**
 * Filters out invalid visible measures.
 * If all measures are invalid then it deletes the key for visible measures and any other settings based on visible measures.
 *
 * Note: Since this method will keep the state as is most of the time,
 * it is more performant to do an in place modification of exploreState, rather than creating a copy.
 */
function validateAndCleanMeasureRelatedExploreState(
  measures: Map<string, MetricsViewSpecMeasure>,
  exploreSpec: V1ExploreSpec,
  exploreState: Partial<ExploreState>,
) {
  if (!exploreState.visibleMeasures) {
    // Each source is meant to have relevant fields.
    // So if there are no visible measures in this source then remove fields dependant on it.
    // Note: This is not exhaustive right now
    delete exploreState.leaderboardSortByMeasureName;
    delete exploreState.leaderboardMeasureNames;
    return [];
  }

  const errors: Error[] = [];

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

  if (selectedMeasures.length > 0) {
    // If there are any remaining valid measures then set it.
    exploreState.allMeasuresVisible =
      selectedMeasures.length === exploreSpec.measures?.length;
    exploreState.visibleMeasures = selectedMeasures;
  } else {
    // Else remove the relevant fields so that cascading merge can set fields from other sources.
    delete exploreState.allMeasuresVisible;
    delete exploreState.visibleMeasures;
    // Remove other fields dependent on measures as well.
    delete exploreState.leaderboardSortByMeasureName;
    delete exploreState.leaderboardMeasureNames;
    // Return early since the rest of the validations assume visible measures are present
    return errors;
  }

  const visibleMeasures = new Set(selectedMeasures);

  if (
    exploreState.leaderboardSortByMeasureName &&
    !visibleMeasures.has(exploreState.leaderboardSortByMeasureName)
  ) {
    const measureIsPresentInMetricsView = measures.has(
      exploreState.leaderboardSortByMeasureName,
    );
    errors.push(
      getSingleFieldError(
        "sort by measure",
        exploreState.leaderboardSortByMeasureName,
        measureIsPresentInMetricsView ? "It is hidden." : "",
      ),
    );

    // Set the 1st valid sort measure if the selected measure is not valid
    exploreState.leaderboardSortByMeasureName = exploreState.visibleMeasures[0];
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

    if (selectedLeaderboardMeasures.length > 0) {
      // If some leaderboard measures are valid then set it.
      exploreState.leaderboardMeasureNames = selectedLeaderboardMeasures;
    } else {
      // Else set the 1st visible measure as leaderboard measure.
      exploreState.leaderboardMeasureNames = [exploreState.visibleMeasures[0]];
    }
  }

  return errors;
}
