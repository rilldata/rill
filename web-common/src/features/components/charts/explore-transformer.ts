import type { ChartSpec } from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import type {
  PivotChipData,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

/**
 * Transforms a chart spec into a pivot state for explore
 */
export function transformChartSpecToPivotState(
  spec: ChartSpec,
  timeGrain: V1TimeGrain | undefined,
): PivotState {
  const columns: PivotChipData[] = [];
  const rows: PivotChipData[] = [];

  // Iterate over all properties in the spec
  for (const [key, value] of Object.entries(spec)) {
    // Skip non-field properties
    if (key === "metrics_view" || key === "title" || key === "description") {
      continue;
    }

    // Check if this property is a field config object
    if (isFieldConfig(value)) {
      const fieldConfig = value;

      let chipType: PivotChipType;
      let id: string;
      if (fieldConfig.fields?.length) {
        columns.push(
          ...fieldConfig.fields.map((f) => ({
            id: f,
            title: f,
            type: PivotChipType.Measure,
          })),
        );
        continue;
      } else if (fieldConfig.type === "quantitative") {
        id = fieldConfig.field;
        chipType = PivotChipType.Measure;
      } else if (fieldConfig.type === "temporal") {
        id = timeGrain || V1TimeGrain.TIME_GRAIN_DAY;
        chipType = PivotChipType.Time;
      } else {
        id = fieldConfig.field;
        chipType = PivotChipType.Dimension;
      }

      if (key === "x" || chipType === PivotChipType.Measure) {
        columns.push({
          id,
          title: fieldConfig.field,
          type: chipType,
        });
      } else {
        rows.push({
          id,
          title: fieldConfig.field,
          type: chipType,
        });
      }
    }
  }

  return {
    columns,
    rows,
    expanded: {},
    sorting: [],
    columnPage: 0,
    rowPage: 0,
    enableComparison: false,
    tableMode: "nest",
    activeCell: null,
  };
}
