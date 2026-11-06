import {
  injectEphemeralMeasuresIntoMap,
  validateEphemeralDefsAgainstSpec,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-state";
import { stripMeasureSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export type ExploreFieldNames = {
  measures: string[];
  dimensions: string[];
};

/**
 * Names from the explore spec's measure and dimension lists that the metrics view spec also defines,
 * in the explore's order.
 * The two specs can disagree: when changes are staged, the runtime keeps serving an explore's last valid spec
 * after its metrics view dropped a field, and a query for that field fails.
 * So state built from the explore spec must only use the names returned here.
 */
export function getExploreFieldNames(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
): ExploreFieldNames {
  const measureNames = new Set(
    metricsViewSpec.measures?.map((m) => m.name) ?? [],
  );
  const dimensionNames = new Set(
    metricsViewSpec.dimensions?.map((d) => d.name) ?? [],
  );
  return {
    measures: exploreSpec.measures?.filter((m) => measureNames.has(m)) ?? [],
    dimensions:
      exploreSpec.dimensions?.filter((d) => dimensionNames.has(d)) ?? [],
  };
}

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
 * TODO: move all validations from convertUrlParamsToPreset here
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
  // Time dimensions cannot be filtered, compared, expanded or pivoted on.
  const nonTimeDimensions = new Map(
    [...dimensions].filter(([, d]) => d.type !== "DIMENSION_TYPE_TIME"),
  );

  // Validate restored ephemeral measures and treat the valid ones as known
  // measure names for the validations below.
  if (exploreState.ephemeralMeasures !== undefined) {
    const { valid, invalidEntries } = validateEphemeralDefsAgainstSpec(
      exploreState.ephemeralMeasures,
      measures,
      metricsViewSpec,
    );
    exploreState.ephemeralMeasures = valid.length ? valid : undefined;
    injectEphemeralMeasuresIntoMap(measures, valid);
    if (invalidEntries.length) {
      errors.push(getMultiFieldError("adhoc measure", invalidEntries));
    }
  }

  const errorsFromExploreView = validateAndCleanExploreViewState(
    measures,
    dimensions,
    exploreSpec,
    exploreState,
  );
  errors.push(...errorsFromExploreView);

  if (exploreState.whereFilter) {
    const { expr, errors: filterErrors } = dropUnknownFieldsFromFilter(
      exploreState.whereFilter,
      measures,
      dimensions,
    );
    exploreState.whereFilter = expr;
    errors.push(...filterErrors);
  }

  errors.push(
    ...validateAndCleanSelectedFields(
      measures,
      nonTimeDimensions,
      exploreState,
    ),
  );

  return errors;
}

/**
 * Drops the conditions of a filter that reference a field missing from the given maps,
 * and the measures a measure filter selects that are missing.
 * Time dimensions cannot be filtered on, so they count as missing.
 * IN and NIN conditions and measure filters (a subquery grouped by a dimension) are certainly on a dimension,
 * so their missing fields are reported apart from the rest.
 * Shared by the url parser, the legacy state converter and the explore state cleaner
 * so that a stale filter is handled the same way everywhere.
 */
export function dropUnknownFieldsFromFilter(
  expr: V1Expression,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
) {
  const isKnownMeasure = (name: string) =>
    measures.has(name) || measures.has(stripMeasureSuffix(name));
  const isKnownDimension = (name: string) =>
    dimensions.has(name) &&
    dimensions.get(name)?.type !== "DIMENSION_TYPE_TIME";

  // Sets, since a stale measure in a measure filter is found both in its HAVING clause and in its measure list.
  const missingDimensions = new Set<string>();
  const missingFields = new Set<string>();
  const cleanedExpr =
    filterIdentifiers(expr, (e, ident) => {
      if (
        e.cond?.op === V1Operation.OPERATION_IN ||
        e.cond?.op === V1Operation.OPERATION_NIN ||
        !!e.subquery
      ) {
        if (!isKnownDimension(ident)) {
          missingDimensions.add(ident);
          return false;
        }
        if (e.subquery?.measures) {
          // The subquery's HAVING clause is cleaned by now,
          // but the runtime resolves every measure the subquery selects, used or not.
          // `e` is a copy made by filterIdentifiers, so replacing its subquery leaves the input as is.
          const subqueryMeasures = e.subquery.measures.filter(isKnownMeasure);
          getMissingValues(subqueryMeasures, e.subquery.measures).forEach((m) =>
            missingFields.add(m),
          );
          e.subquery = { ...e.subquery, measures: subqueryMeasures };
        }
        return true;
      }

      if (isKnownMeasure(ident) || isKnownDimension(ident)) {
        return true;
      }
      missingFields.add(ident);
      return false;
    }) ?? createAndExpression([]);

  const errors: Error[] = [];
  if (missingDimensions.size) {
    errors.push(getMultiFieldError("filter dimension", [...missingDimensions]));
  }
  if (missingFields.size) {
    errors.push(getMultiFieldError("filter field", [...missingFields]));
  }
  return { expr: cleanedExpr, errors };
}

/**
 * Cleans the fields selected for comparison, the dimension table, time dimension details and the pivot.
 * A view that was opened on a field that no longer exists is reset to the default view.
 *
 * Note: Modifies exploreState in place, like the rest of the validations here.
 */
function validateAndCleanSelectedFields(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  exploreState: Partial<ExploreState>,
) {
  const errors: Error[] = [];

  if (
    exploreState.selectedComparisonDimension &&
    !dimensions.has(exploreState.selectedComparisonDimension)
  ) {
    errors.push(
      getSingleFieldError(
        "compare dimension",
        exploreState.selectedComparisonDimension,
      ),
    );
    delete exploreState.selectedComparisonDimension;
  }

  if (
    exploreState.selectedDimensionName &&
    !dimensions.has(exploreState.selectedDimensionName)
  ) {
    errors.push(
      getSingleFieldError(
        "expanded dimension",
        exploreState.selectedDimensionName,
      ),
    );
    delete exploreState.selectedDimensionName;
    if (exploreState.activePage === DashboardState_ActivePage.DIMENSION_TABLE) {
      exploreState.activePage = DashboardState_ActivePage.DEFAULT;
    }
  }

  if (
    exploreState.tdd?.expandedMeasureName &&
    !measures.has(exploreState.tdd.expandedMeasureName)
  ) {
    errors.push(
      getSingleFieldError(
        "expanded measure",
        exploreState.tdd.expandedMeasureName,
      ),
    );
    exploreState.tdd.expandedMeasureName = "";
    if (
      exploreState.activePage ===
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
    ) {
      exploreState.activePage = DashboardState_ActivePage.DEFAULT;
    }
  }

  if (exploreState.pivot) {
    const isKnownChip = (chip: PivotChipData) => {
      switch (chip.type) {
        case PivotChipType.Time:
          return true;
        case PivotChipType.Measure:
          return measures.has(chip.id);
        default:
          return dimensions.has(chip.id);
      }
    };
    const rows = exploreState.pivot.rows?.filter(isKnownChip) ?? [];
    const missingRows = getMissingValues(
      rows.map((r) => r.id),
      exploreState.pivot.rows?.map((r) => r.id) ?? [],
    );
    if (missingRows.length) {
      errors.push(getMultiFieldError("pivot row", missingRows));
      exploreState.pivot.rows = rows;
    }
    const columns = exploreState.pivot.columns?.filter(isKnownChip) ?? [];
    const missingColumns = getMissingValues(
      columns.map((c) => c.id),
      exploreState.pivot.columns?.map((c) => c.id) ?? [],
    );
    if (missingColumns.length) {
      errors.push(getMultiFieldError("pivot column", missingColumns));
      exploreState.pivot.columns = columns;
    }
  }

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
    // `measures` includes ephemeral measures, so a hidden spec measure can
    // never be masked by a visible ephemeral one.
    exploreState.allMeasuresVisible = selectedMeasures.length === measures.size;
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

  // Ensure leaderboardSortByMeasureName is in leaderboardMeasureNames (sync invariant)
  if (
    exploreState.leaderboardMeasureNames?.length &&
    exploreState.leaderboardSortByMeasureName &&
    !exploreState.leaderboardMeasureNames.includes(
      exploreState.leaderboardSortByMeasureName,
    )
  ) {
    exploreState.leaderboardSortByMeasureName =
      exploreState.leaderboardMeasureNames[0];
  }

  return errors;
}
