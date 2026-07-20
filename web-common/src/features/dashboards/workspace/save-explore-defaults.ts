import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import { queryServiceConvertExpressionToMetricsSQL } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { parseDocument } from "yaml";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export type ComparisonModeValue =
  | "none"
  | "time"
  | "dimension"
  | "rill-PP"
  | "rill-PD"
  | "rill-PW"
  | "rill-PM"
  | "rill-PQ"
  | "rill-PY";

export type ExploreDefaults = {
  filter?: string;
  measures?: string[];
  dimensions?: string[];
  comparison_mode?: ComparisonModeValue;
  comparison_dimension?: string;
  time_range?: string;
  pinned?: string[];
};

export async function saveExploreDefaults(
  runtimeClient: RuntimeClient,
  fileArtifact: FileArtifact,
  exploreState: ExploreState,
  autoSave: boolean,
) {
  const doc = parseDocument(get(fileArtifact.editorContent) ?? "");
  const defaults: ExploreDefaults = {};

  // Time range (skip CUSTOM and ALL_TIME)
  const timeRangeName = exploreState.selectedTimeRange?.name;
  if (
    timeRangeName &&
    timeRangeName !== TimeRangePreset.CUSTOM &&
    timeRangeName !== TimeRangePreset.ALL_TIME
  ) {
    defaults.time_range = timeRangeName;
  }

  // Comparison
  if (exploreState.showTimeComparison) {
    defaults.comparison_mode = "time"; // TODO: set comparison time
  } else if (exploreState.selectedComparisonDimension) {
    defaults.comparison_mode = "dimension";
    defaults.comparison_dimension = exploreState.selectedComparisonDimension;
  }

  // Visible measures/dimensions
  if (exploreState.visibleMeasures?.length) {
    defaults.measures = [...exploreState.visibleMeasures];
  }
  if (exploreState.visibleDimensions?.length) {
    defaults.dimensions = [...exploreState.visibleDimensions];
  }

  // Filters → SQL string via queryServiceConvertExpressionToMetricsSQL
  const merged = mergeDimensionAndMeasureFilters(
    exploreState.whereFilter,
    exploreState.dimensionThresholdFilters ?? [],
  );
  if (merged?.cond?.exprs?.length) {
    try {
      const { sql } = await queryServiceConvertExpressionToMetricsSQL(
        runtimeClient,
        { expression: merged as any },
      );
      if (sql) defaults.filter = sql;
    } catch {
      // If conversion fails, skip filter
    }
  }

  // Pinned filters
  if (exploreState.pinnedFilters?.size) {
    defaults.pinned = Array.from(exploreState.pinnedFilters);
  }

  doc.set("defaults", defaults);
  fileArtifact.updateEditorContent(doc.toString(), false, autoSave);
}
