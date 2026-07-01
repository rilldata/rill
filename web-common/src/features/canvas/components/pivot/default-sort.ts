import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  COMPARISON_VALUE,
  type PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  extractNumbers,
  getTimeGrainFromDimension,
  isTimeDimension,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { DefaultSort } from "./index";

/** Separator between column-dimension values in a nested-sort label. */
const LABEL_SEPARATOR = " › ";

/** Matches a nested pivot leaf accessor, e.g. "c0v2_c1v0m0" or "c0v2m0". */
const NESTED_ACCESSOR_RE = /^(c\d+v\d+_?)+m\d+$/;

/**
 * Builds the human-readable label for a sort id. For stable ids (measure,
 * dimension, time grain) this resolves the display name from the metrics view.
 * For nested pivot leaf accessors it decodes the column-dimension value path
 * and joins it with the measure display name, e.g. "US › 2024 › Sales".
 *
 * Returns the raw id as a fallback when nothing better can be resolved.
 */
export function getSortLabel(
  id: string,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): string {
  if (NESTED_ACCESSOR_RE.test(id)) {
    return getNestedAccessorLabel(id, config, columnDimensionAxes);
  }
  return getFieldLabel(id, config);
}

export type SortFieldType = "measure" | "dimension" | "time";

export interface SortChip {
  label: string;
  type: SortFieldType;
}

/**
 * Breaks a sort id into the chips shown in the inspector. A stable field is a
 * single chip; a nested pivot leaf accessor becomes one chip per column-
 * dimension value followed by the measure chip, e.g. [US, 2024, Sales].
 */
export function getSortChips(
  id: string,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): SortChip[] {
  if (NESTED_ACCESSOR_RE.test(id)) {
    const [colPart, measureIndexPart] = id.split("m");
    const parts = colPart.split("_").filter(Boolean);

    const chips: SortChip[] = parts.map((part) => {
      const { c, v } = extractNumbers(part);
      const dimensionName = config.colDimensionNames[c];
      return {
        label: columnDimensionAxes[dimensionName]?.[v] ?? "",
        type: isTimeDimension(dimensionName, config.time.timeDimension)
          ? "time"
          : "dimension",
      };
    });

    const measureName = config.measureNames[parseInt(measureIndexPart)];
    if (measureName) {
      chips.push({
        label: getFieldLabel(measureName, config),
        type: "measure",
      });
    }
    return chips;
  }

  return [
    { label: getFieldLabel(id, config), type: getSortFieldType(id, config) },
  ];
}

/**
 * Resolves the field type of a sort id, used to color the inspector chip.
 * A nested pivot leaf accessor sorts by its measure, so it reads as "measure".
 */
export function getSortFieldType(
  id: string,
  config: PivotDataStoreConfig,
): SortFieldType {
  if (NESTED_ACCESSOR_RE.test(id)) return "measure";
  if (isTimeDimension(id, config.time.timeDimension)) return "time";
  if (config.allMeasures.some((m) => m.name === id)) return "measure";
  return "dimension";
}

/** Human-readable suffix appended to a base measure for comparison columns. */
const COMPARISON_MODIFIER: Record<string, string> = {
  [COMPARISON_VALUE]: " (previous)",
  [COMPARISON_DELTA]: " Δ",
  [COMPARISON_PERCENT]: " Δ %",
};

function getFieldLabel(field: string, config: PivotDataStoreConfig): string {
  if (isTimeDimension(field, config.time.timeDimension)) {
    const grain = getTimeGrainFromDimension(field);
    return `Time ${TIME_GRAIN[grain]?.label ?? grain}`;
  }

  // Comparison columns are the base measure plus a suffix (e.g. "sales__delta_rel").
  for (const [suffix, modifier] of Object.entries(COMPARISON_MODIFIER)) {
    if (field.endsWith(suffix)) {
      const baseName = field.slice(0, -suffix.length);
      return getFieldLabel(baseName, config) + modifier;
    }
  }

  const measure = config.allMeasures.find((m) => m.name === field);
  if (measure) return measure.displayName || (measure.name as string);

  const dimension = config.allDimensions.find(
    (d) => (d.name || d.column) === field,
  );
  if (dimension) {
    return (
      dimension.displayName || dimension.name || (dimension.column as string)
    );
  }

  return field;
}

function getNestedAccessorLabel(
  accessor: string,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]>,
): string {
  const [colPart, measureIndexPart] = accessor.split("m");
  const parts = colPart.split("_").filter(Boolean);

  const valueLabels = parts.map((part) => {
    const { c, v } = extractNumbers(part);
    const dimensionName = config.colDimensionNames[c];
    return columnDimensionAxes[dimensionName]?.[v] ?? "";
  });

  const measureName = config.measureNames[parseInt(measureIndexPart)];
  const measureLabel = measureName ? getFieldLabel(measureName, config) : "";

  return [...valueLabels, measureLabel].filter(Boolean).join(LABEL_SEPARATOR);
}

/**
 * Whether two sort configs target the same column and direction. Used to
 * disable the "Make default" button when the active sort already is the default.
 */
export function sortEquals(
  a: DefaultSort | undefined,
  b: { id: string; desc: boolean } | undefined,
): boolean {
  if (!a || !b) return false;
  return a.id === b.id && a.desc === b.desc;
}
