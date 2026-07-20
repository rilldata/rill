import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import { queryServiceConvertExpressionToMetricsSQL } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { parseDocument } from "yaml";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { YAMLOnlyExploreState } from "@rilldata/web-common/features/dashboards/stores/yaml-only-explore-state.svelte.ts";

export type ComparisonModeValue = "none" | "time" | "dimension";

export type ExploreDefaults = {
  filter?: string;
  pinned_filters?: string[];
  required_filters?: string[];
  measures?: string[];
  dimensions?: string[];
  comparison_mode?: ComparisonModeValue;
  comparison_dimension?: string;
  time_range?: string;
};

export async function saveExploreDefaults(
  runtimeClient: RuntimeClient,
  fileArtifact: FileArtifact,
  exploreState: ExploreState,
  yamlOnlyConfig: YAMLOnlyExploreState,
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
  if (yamlOnlyConfig.pinnedFilters.value.length) {
    defaults.pinned_filters = Array.from(yamlOnlyConfig.pinnedFilters.value);
  }

  // Required filters
  if (yamlOnlyConfig.requiredFilters.value.length) {
    defaults.required_filters = Array.from(
      yamlOnlyConfig.requiredFilters.value,
    );
  }

  doc.set("defaults", defaults);
  fileArtifact.updateEditorContent(doc.toString(), false, true);
}
