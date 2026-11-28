import {
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { forEachIdentifier } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  BOOLEANS,
  INTEGERS,
  isFloat,
} from "@rilldata/web-common/lib/duckdb-data-types";
import { dedupePivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-chip-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector.ts";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { getRangePrecision } from "@rilldata/web-common/lib/time/rill-time-grains.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import { applyAdaptiveTimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/adaptive-time-grain.ts";

/**
 * Corrects the final merged explore state.
 * Fixes invalid values (advanced measures, leaderboard sync) rather than removing them.
 * Called after cascading merge to fix mismatches from combining different sources.
 */
export function correctExploreState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreState: ExploreState,
) {
  // The grain is resolved first since measure validation below compares against it.
  if (exploreState.adaptiveTimeGrain) {
    applyAdaptiveTimeGrain(exploreState, metricsViewSpec);
  } else if (exploreState.selectedTimeRange) {
    deriveIntervalFromRillTimeName(exploreState.selectedTimeRange);
  }

  // Resuse code for now. We might want to consolidate more in the future.
  AdvancedMeasureCorrector.correct(exploreState, metricsViewSpec);

  correctLeaderboardMeasures(exploreState);
  correctPivotFields(exploreState);
  correctTimeDimensionDetails(metricsViewSpec, exploreState);
  correctFilterValueTypes(metricsViewSpec, exploreState);
}

const ValueComparisonOperations = new Set<string>([
  V1Operation.OPERATION_IN,
  V1Operation.OPERATION_NIN,
  V1Operation.OPERATION_EQ,
  V1Operation.OPERATION_NEQ,
]);

/**
 * Legacy dashboard states stored every filter value as a string, whatever the dimension's type.
 * The url filter writes numbers and booleans unquoted, so a string value on such a dimension only comes from legacy state.
 * Convert them using the dimension's type, so that engines without implicit casts can run the filter.
 */
function correctFilterValueTypes(
  metricsViewSpec: V1MetricsViewSpec,
  exploreState: ExploreState,
) {
  if (!exploreState.whereFilter) return;

  forEachIdentifier(exploreState.whereFilter, (e, ident) => {
    if (!e.cond?.op || !ValueComparisonOperations.has(e.cond.op)) return;
    const code = metricsViewSpec.dimensions?.find((d) => d.name === ident)
      ?.dataType?.code;
    if (!code) return;

    e.cond.exprs = e.cond.exprs?.map((valueExpr, i) => {
      // The 1st expr is the identifier.
      if (i === 0 || typeof valueExpr.val !== "string") return valueExpr;
      return { val: convertStringFilterValue(valueExpr.val, code) };
    });
  });
}

function convertStringFilterValue(val: string, code: string) {
  let converted: number | boolean | undefined;
  if (INTEGERS.has(code)) {
    converted = Number.parseInt(val);
  } else if (isFloat(code)) {
    converted = Number.parseFloat(val);
  } else if (BOOLEANS.has(code) && (val === "true" || val === "false")) {
    converted = val === "true";
  }
  if (converted === undefined || Number.isNaN(converted)) return val;
  return converted;
}

/**
 * Time dimension details need a time dimension to chart against.
 * A link can still open them on an explore without one,
 * e.g. a report's request without dimensions is mapped without the specs.
 */
function correctTimeDimensionDetails(
  metricsViewSpec: V1MetricsViewSpec,
  exploreState: ExploreState,
) {
  if (
    exploreState.activePage !==
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL ||
    metricsViewSpec.timeDimension
  ) {
    return;
  }
  exploreState.activePage = DashboardState_ActivePage.DEFAULT;
  exploreState.tdd = { ...exploreState.tdd, expandedMeasureName: "" };
}

/**
 * Drops a pivot field listed twice in a zone, or in both zones, and fields
 * with an empty name. Every loaded state ends here, whatever its source
 * (url, bookmark, legacy proto state, session or yaml defaults), and the
 * pivot header keys its chips by field, so a repeat would break the view.
 */
function correctPivotFields(exploreState: ExploreState) {
  exploreState.pivot.rows = dedupePivotChips(exploreState.pivot.rows);
  exploreState.pivot.columns = dedupePivotChips(
    exploreState.pivot.columns,
    exploreState.pivot.rows.map((chip) => chip.id),
  );
}

function correctLeaderboardMeasures(exploreState: ExploreState) {
  const sortMeasureIsInvalid = Boolean(
    exploreState.leaderboardMeasureNames?.length &&
      !exploreState.leaderboardMeasureNames.includes(
        exploreState.leaderboardSortByMeasureName,
      ),
  );
  const leaderboardMeasuresAreInvalid = Boolean(
    !exploreState.leaderboardMeasureNames?.length &&
      exploreState.leaderboardSortByMeasureName,
  );

  if (sortMeasureIsInvalid) {
    exploreState.leaderboardSortByMeasureName =
      exploreState.leaderboardMeasureNames[0];
  } else if (leaderboardMeasuresAreInvalid) {
    exploreState.leaderboardMeasureNames = [
      exploreState.leaderboardSortByMeasureName,
    ];
  }
}

/**
 * Derives and sets the interval (time grain) on a time range from its RillTime name.
 * This is needed when the URL doesn't explicitly specify a grain.
 */
function deriveIntervalFromRillTimeName(
  selectedRange: DashboardTimeControls | undefined,
): void {
  if (!selectedRange?.name || selectedRange.interval) return;

  try {
    const parsed = parseRillTime(selectedRange.name);
    selectedRange.interval = getRangePrecision(parsed);
  } catch {
    // Parsing fails for non-rill-time names like "CUSTOM" - use undefined
  }
}
