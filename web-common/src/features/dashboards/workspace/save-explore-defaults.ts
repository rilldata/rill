import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { queryServiceConvertExpressionToMetricsSQL } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { parseDocument } from "yaml";

export type ExploreDefaults = {
  filter?: string;
  measures?: string[];
  dimensions?: string[];
  comparison_mode?: string;
  comparison_dimension?: string;
  time_range?: string;
  pinned?: string[];
};

export async function saveExploreDefaults(
  fileArtifact: FileArtifact,
  exploreState: ExploreState,
  instanceId: string,
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
    const comparisonName = exploreState.selectedComparisonTimeRange?.name;
    defaults.comparison_mode = comparisonName || "rill-PP";
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
        instanceId,
        { expression: merged },
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
