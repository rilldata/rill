import type { SortingState } from "tanstack-table-8-svelte-5";
import type { PivotDataStoreConfig, PivotSortTarget } from "./types";

const NESTED_ACCESSOR_RE = /^(c\d+v\d+)(?:_c\d+v\d+)*m\d+$/;

type NestedAccessor = {
  columnIndexes: Array<{ dimensionIndex: number; valueIndex: number }>;
  measureIndex: number;
};

function parseNestedAccessor(accessor: string): NestedAccessor | undefined {
  if (!NESTED_ACCESSOR_RE.test(accessor)) return undefined;

  const measureSeparator = accessor.lastIndexOf("m");
  const columnPart = accessor.slice(0, measureSeparator);
  const measureIndex = Number(accessor.slice(measureSeparator + 1));
  const columnIndexes = columnPart.split("_").map((part) => {
    const match = /^c(\d+)v(\d+)$/.exec(part);
    return match
      ? { dimensionIndex: Number(match[1]), valueIndex: Number(match[2]) }
      : undefined;
  });

  if (
    !Number.isInteger(measureIndex) ||
    columnIndexes.some((part) => part === undefined)
  ) {
    return undefined;
  }

  return {
    columnIndexes: columnIndexes as NestedAccessor["columnIndexes"],
    measureIndex,
  };
}

/**
 * Converts a current TanStack sort ID into a stable sort target. This is also
 * the compatibility decoder for existing Explore URLs and protobuf state.
 */
export function decodePivotSort(
  sorting: SortingState | undefined,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): PivotSortTarget | undefined {
  const sort = sorting?.[0];
  if (!sort) return undefined;

  const stableFields = config.isFlat
    ? [...config.rowDimensionNames, ...config.measureNames]
    : [config.rowDimensionNames[0], ...config.measureNames].filter(Boolean);

  if (stableFields.includes(sort.id)) {
    return { field: sort.id, desc: sort.desc };
  }

  if (config.isFlat) return undefined;

  const parsed = parseNestedAccessor(sort.id);
  if (
    !parsed ||
    parsed.columnIndexes.length !== config.colDimensionNames.length
  ) {
    return undefined;
  }

  const measure = config.measureNames[parsed.measureIndex];
  if (!measure) return undefined;

  const columnValues: Extract<
    PivotSortTarget,
    { measure: string }
  >["column_values"] = [];

  for (const [index, part] of parsed.columnIndexes.entries()) {
    if (part.dimensionIndex !== index) return undefined;

    const dimension = config.colDimensionNames[part.dimensionIndex];
    const value = columnDimensionAxes[dimension]?.[part.valueIndex];
    if (dimension === undefined || value === undefined) return undefined;

    columnValues.push({ dimension, value });
  }

  return {
    measure,
    column_values: columnValues,
    desc: sort.desc,
  };
}

/** Returns the configured default only when it still matches the current table. */
export function validatePivotSortTarget(
  target: PivotSortTarget | undefined,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): PivotSortTarget | undefined {
  if (!target) return undefined;

  if ("field" in target) {
    const stableFields = config.isFlat
      ? [...config.rowDimensionNames, ...config.measureNames]
      : [config.rowDimensionNames[0], ...config.measureNames].filter(Boolean);
    return stableFields.includes(target.field) ? target : undefined;
  }

  if (
    config.isFlat ||
    !config.measureNames.includes(target.measure) ||
    target.column_values.length !== config.colDimensionNames.length
  ) {
    return undefined;
  }

  const valid = target.column_values.every((entry, index) => {
    const dimension = config.colDimensionNames[index];
    return (
      entry.dimension === dimension &&
      columnDimensionAxes[dimension]?.includes(entry.value)
    );
  });

  return valid ? target : undefined;
}

/**
 * Interactive state always wins. An invalid legacy interactive ID deliberately
 * resolves to no sort instead of falling through to a Canvas default.
 */
export function getEffectivePivotSort(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): PivotSortTarget | undefined {
  if (config.pivot.sorting?.length) {
    return decodePivotSort(config.pivot.sorting, config, columnDimensionAxes);
  }
  return validatePivotSortTarget(
    config.defaultSort,
    config,
    columnDimensionAxes,
  );
}

/** Converts a stable target back to the current TanStack column ID for display. */
export function encodePivotSort(
  target: PivotSortTarget | undefined,
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> = {},
): SortingState {
  const validTarget = validatePivotSortTarget(
    target,
    config,
    columnDimensionAxes,
  );
  if (!validTarget) return [];

  if ("field" in validTarget) {
    return [{ id: validTarget.field, desc: validTarget.desc }];
  }

  const accessorParts: string[] = [];
  for (let index = 0; index < validTarget.column_values.length; index++) {
    const entry = validTarget.column_values[index];
    const valueIndex = columnDimensionAxes[entry.dimension]?.indexOf(
      entry.value,
    );
    if (valueIndex === undefined || valueIndex < 0) return [];
    accessorParts.push(`c${index}v${valueIndex}`);
  }

  const measureIndex = config.measureNames.indexOf(validTarget.measure);
  if (measureIndex < 0) return [];

  return [
    {
      id: `${accessorParts.join("_")}m${measureIndex}`,
      desc: validTarget.desc,
    },
  ];
}

export function pivotSortTargetsEqual(
  left: PivotSortTarget | undefined,
  right: PivotSortTarget | undefined,
): boolean {
  if (!left || !right || left.desc !== right.desc) return false;
  if ("field" in left || "field" in right) {
    return "field" in left && "field" in right && left.field === right.field;
  }
  if (
    left.measure !== right.measure ||
    left.column_values.length !== right.column_values.length
  ) {
    return false;
  }
  return left.column_values.every(
    (entry, index) =>
      entry.dimension === right.column_values[index].dimension &&
      entry.value === right.column_values[index].value,
  );
}
