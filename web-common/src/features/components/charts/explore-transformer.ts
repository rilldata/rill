import type {
  ChartSpec,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
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

    if (!isFieldConfig(value)) {
      continue;
    }

    const fieldConfig = value;

    if (fieldConfig.type === "value") {
      continue;
    }

    // Handle multiple fields case
    if (fieldConfig.fields?.length) {
      columns.push(
        ...fieldConfig.fields.map((f) => ({
          id: f,
          title: f,
          type: PivotChipType.Measure,
        })),
      );
      continue;
    }

    // Determine chip type and id based on field config
    const { chipType, id } = getChipTypeAndId(fieldConfig, timeGrain);

    // Add to columns or rows based on key and chip type
    const chipData = {
      id,
      title: fieldConfig.field,
      type: chipType,
    };

    if (chipType === PivotChipType.Measure) {
      columns.push(chipData);
    } else {
      rows.push(chipData);
    }
  }

  const hasDimensionRows = rows.some((r) => r.type === PivotChipType.Dimension);
  if (hasDimensionRows) {
    const timeChips = rows.filter((r) => r.type === PivotChipType.Time);
    const nonTimeRows = rows.filter((r) => r.type !== PivotChipType.Time);
    columns.push(...timeChips);
    rows.length = 0;
    rows.push(...nonTimeRows);
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

function getChipTypeAndId(
  fieldConfig: FieldConfig,
  timeGrain: V1TimeGrain | undefined,
): { chipType: PivotChipType; id: string } {
  if (fieldConfig.type === "quantitative") {
    return {
      chipType: PivotChipType.Measure,
      id: fieldConfig.field,
    };
  }

  if (fieldConfig.type === "temporal") {
    return {
      chipType: PivotChipType.Time,
      id: timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
    };
  }

  return {
    chipType: PivotChipType.Dimension,
    id: fieldConfig.field,
  };
}
