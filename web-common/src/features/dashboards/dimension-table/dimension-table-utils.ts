import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  copyFilterExpression,
  createAndExpression,
  createInExpression,
  createLikeExpression,
  createOrExpression,
  filterExpressions,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1MetricsViewAggregationResponseDataItem,
  V1Operation,
} from "../../../runtime-client";
import PercentOfTotal from "./PercentOfTotal.svelte";

import { PERC_DIFF } from "../../../components/data-types/type-utils";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1Expression,
  V1MetricsViewToplistResponseDataItem,
} from "../../../runtime-client";

import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";

import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import type { SvelteComponent } from "svelte";
import { SortType } from "../proto-state/derived-types";
import { getFiltersForOtherDimensions } from "../selectors";
import type { ExploreState } from "web-common/src/features/dashboards/stores/explore-state";
import type { DimensionTableRow } from "./dimension-table-types";
import type { DimensionTableConfig } from "./DimensionTableConfig";

/** Returns an updated filter set for a given dimension on search */
export function updateFilterOnSearch(
  filterForDimension: V1Expression,
  searchText: string,
  dimensionName: string,
): V1Expression | undefined {
  if (!filterForDimension) return undefined;
  // create a copy
  const addNull = "null".includes(searchText);
  if (searchText !== "") {
    let cond: V1Expression;
    if (addNull) {
      cond = createOrExpression([
        // TODO: do we need a `IS NULL` expression?
        createInExpression(dimensionName, [null]),
        createLikeExpression(dimensionName, `%${searchText}%`),
      ]);
    } else {
      cond = createLikeExpression(dimensionName, `%${searchText}%`);
    }

    filterForDimension = copyFilterExpression(filterForDimension);
    const filterIdx = filterForDimension.cond?.exprs?.findIndex((e) =>
      matchExpressionByName(e, dimensionName),
    );
    if (filterIdx === undefined || filterIdx === -1) {
      filterForDimension.cond?.exprs?.push(cond);
    } else {
      filterForDimension.cond?.exprs?.splice(filterIdx, 0, cond);
    }
  } else {
    filterForDimension =
      filterExpressions(
        filterForDimension,
        (e) =>
          e.cond?.op !== V1Operation.OPERATION_LIKE &&
          e.cond?.op !== V1Operation.OPERATION_NLIKE,
      ) ?? createAndExpression([]);
  }
  return filterForDimension;
}

export function getDimensionFilterWithSearch(
  filters: V1Expression,
  searchText: string,
  dimensionName: string,
) {
  const filterForDimension =
    getFiltersForOtherDimensions(filters, dimensionName) ??
    createAndExpression([]);

  return updateFilterOnSearch(filterForDimension, searchText, dimensionName);
}

export function computePercentOfTotal(
  values: V1MetricsViewToplistResponseDataItem[],
  total: number,
  measureName: string,
) {
  for (const value of values) {
    if (total === 0 || total === null || total === undefined) {
      value[measureName + "_percent_of_total"] =
        PERC_DIFF.CURRENT_VALUE_NO_DATA;
    } else {
      value[measureName + "_percent_of_total"] =
        formatMeasurePercentageDifference(
          (value[measureName] as number) / total,
        );
    }
  }

  return values;
}

export function getComparisonProperties(
  measureName: string,
  selectedMeasure: MetricsViewSpecMeasure,
): {
  component: typeof SvelteComponent<any>;
  type: string;
  format: string;
  description: string;
} {
  if (measureName.includes("_delta_perc")) {
    return {
      component: DeltaChangePercentage,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Percentage change over comparison period",
    };
  } else if (measureName.includes("_delta")) {
    return {
      component: DeltaChange,
      type: "RILL_CHANGE",
      format: selectedMeasure.formatPreset ?? FormatPreset.HUMANIZE,
      description: "Change over comparison period",
    };
  } else if (measureName.includes("_percent_of_total")) {
    return {
      component: PercentOfTotal,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Percent of total",
    };
  }
  throw new Error(
    "Invalid measure name, getComparisonProperties must only be called on context columns",
  );
}

export function estimateColumnCharacterWidths(
  columns: VirtualizedTableColumns[],
  rows: V1MetricsViewToplistResponseDataItem[],
) {
  const columnWidths: { [key: string]: number } = {};
  let largestColumnLength = 0;
  columns.forEach((column, i) => {
    // get values
    const values = rows
      .filter((row) => row[column.name] !== null)
      .map(
        (row) =>
          `${row["__formatted_" + column.name] || row[column.name]}`.length,
      );
    values.sort();
    const largest = Math.max(...values);
    columnWidths[column.name] = largest;
    if (i != 0) {
      largestColumnLength = Math.max(
        largestColumnLength,
        column.label?.length || column.name.length,
      );
    }
  });
  return { columnWidths, largestColumnLength };
}

/** this is a perceived character width value, in pixels, when our monospace
 * font is 12px high. */
const CHARACTER_WIDTH = 7;
const CHARACTER_X_PAD = 16 * 2;
const HEADER_ICON_WIDTHS = 16;
const HEADER_X_PAD = CHARACTER_X_PAD;
const HEADER_FLEX_SPACING = 14;
// const CHARACTER_LIMIT_FOR_WRAPPING = 9;

export function estimateColumnSizes(
  columns: VirtualizedTableColumns[],
  columnWidths: {
    [key: string]: number;
  },
  containerWidth: number,
  config: DimensionTableConfig,
): number[] {
  const estimatedColumnSizes = columns.map((column, i) => {
    if (column.name.includes("delta")) return config.comparisonColumnWidth;
    if (column.name.includes("percent_of_total"))
      return config.comparisonColumnWidth;
    if (i != 0) return config.defaultColumnWidth;

    const largestStringLength =
      columnWidths[column.name] * CHARACTER_WIDTH + CHARACTER_X_PAD;

    /** The header width is largely a function of the total number of characters in the column.*/
    const headerWidth =
      (column.label?.length || column.name.length) * CHARACTER_WIDTH +
      HEADER_ICON_WIDTHS +
      HEADER_X_PAD +
      HEADER_FLEX_SPACING;

    /** If the header is bigger than the largestStringLength and that's not at threshold, default to threshold.
     * This will prevent the case where we have very long column names for very short column values.
     */
    const effectiveHeaderWidth =
      headerWidth > 160 && largestStringLength < 160
        ? config.minHeaderWidthWhenColumsAreSmall
        : headerWidth;

    return largestStringLength
      ? Math.min(
          config.maxColumnWidth,
          Math.max(
            largestStringLength,
            effectiveHeaderWidth,
            /** All columns must be minColumnWidth regardless of user settings. */
            config.minColumnWidth,
          ),
        )
      : /** if there isn't a longet string length for some reason, let's go with a
         * default column width. We should not be in this state.
         */
        config.defaultColumnWidth;
  });

  return estimatedColumnSizes;
}

export function prepareVirtualizedDimTableColumns(
  exploreState: ExploreState,
  allMeasures: MetricsViewSpecMeasure[],
  maxValues: { [key: string]: number },
  dimension: MetricsViewSpecDimension,
  timeComparison: boolean,
  validPercentOfTotal: boolean,
  activeMeasures?: string[],
): VirtualizedTableColumns[] {
  const sortType = exploreState.dashboardSortType;
  const sortDirection = exploreState.sortDirection;

  const measureNames = allMeasures.map((m) => m.name);
  const leaderboardSortByMeasureName =
    exploreState.leaderboardSortByMeasureName;
  const selectedMeasure = allMeasures.find(
    (m) => m.name === leaderboardSortByMeasureName,
  );

  const dimensionColumn = dimension.name ?? "";

  // copy column names so we don't mutate the original
  const columnNames = exploreState.visibleMeasures.filter((m) =>
    allMeasures.some((am) => am.name === m),
  );

  // Show context columns based on selected context columns and time comparison settings
  if (selectedMeasure) {
    // If activeMeasures is provided and leaderboardShowContextForAllMeasures is true, add context columns for each active measure
    if (
      activeMeasures?.length &&
      exploreState.leaderboardShowContextForAllMeasures
    ) {
      activeMeasures.forEach((measureName) => {
        const measure = allMeasures.find((m) => m.name === measureName);
        if (measure) {
          addContextColumnNames(
            columnNames,
            timeComparison,
            validPercentOfTotal,
            measure,
          );
        }
      });
    } else {
      // Only add context columns for the leaderboardSortByMeasureName
      addContextColumnNames(
        columnNames,
        timeComparison,
        validPercentOfTotal,
        selectedMeasure,
      );
    }
  }

  // Make dimension the first column
  columnNames.unshift(dimensionColumn);

  const columns = columnNames
    .map((name) => {
      // Determine if this column is related to the selected measure
      const isSelectedMeasureColumn = name === selectedMeasure?.name;
      const isSelectedMeasureDelta = name === `${selectedMeasure?.name}_delta`;
      const isSelectedMeasureDeltaPerc =
        name === `${selectedMeasure?.name}_delta_perc`;
      const isSelectedMeasurePercent =
        name === `${selectedMeasure?.name}_percent_of_total`;

      // Determine highlighting
      let highlight = false;
      if (sortType === SortType.DIMENSION) {
        highlight = name === dimensionColumn;
      } else {
        highlight =
          isSelectedMeasureColumn ||
          isSelectedMeasureDelta ||
          isSelectedMeasureDeltaPerc ||
          isSelectedMeasurePercent;
      }

      // Determine sorting
      let sorted;
      if (sortType === SortType.DIMENSION && name === dimensionColumn) {
        sorted = sortDirection;
      } else if (sortType === SortType.VALUE && isSelectedMeasureColumn) {
        sorted = sortDirection;
      } else if (
        sortType === SortType.DELTA_ABSOLUTE &&
        isSelectedMeasureDelta
      ) {
        sorted = sortDirection;
      } else if (
        sortType === SortType.DELTA_PERCENT &&
        isSelectedMeasureDeltaPerc
      ) {
        sorted = sortDirection;
      } else if (sortType === SortType.PERCENT && isSelectedMeasurePercent) {
        sorted = sortDirection;
      }

      let columnOut: VirtualizedTableColumns | undefined = undefined;
      if (measureNames.includes(name)) {
        // Handle all regular measures
        const measure = allMeasures.find((m) => m.name === name);
        columnOut = {
          name,
          type: "INT",
          label: measure?.displayName || measure?.expression,
          description: measure?.description,
          max: maxValues[measure?.name ?? ""] || 0,
          enableResize: false,
          format: measure?.formatPreset,
          highlight,
          sorted,
        };
      } else if (name === dimensionColumn) {
        // Handle dimension column
        columnOut = {
          name,
          type: "VARCHAR",
          label: dimension?.displayName,
          enableResize: true,
          highlight,
          sorted,
        };
      } else if (selectedMeasure !== undefined) {
        // Handle delta, delta_perc, and percent_of_total columns
        const comparison = getComparisonProperties(name, selectedMeasure);
        columnOut = {
          name,
          type: comparison.type,
          label: comparison.component,
          description: comparison.description,
          enableResize: false,
          format: comparison.format,
          highlight,
          sorted,
        };
      }
      return columnOut;
    })
    .filter((column) => column !== undefined);

  // cast is safe, because we filtered out undefined columns
  return columns ?? [];
}

/**
 * Splices the context column names into the list of dimension
 * table column names.
 *
 * This mutates the columnNames array.
 */
export function addContextColumnNames(
  columnNames: string[],
  timeComparison: boolean,
  validPercentOfTotal: boolean,
  selectedMeasure: MetricsViewSpecMeasure,
) {
  const name = selectedMeasure?.name;
  if (!name) return;

  const sortByColumnIndex = columnNames.indexOf(name);
  const isPercent = selectedMeasure?.formatPreset === FormatPreset.PERCENTAGE;
  let nextIndex = sortByColumnIndex + 1;

  // 1. Add percent of total first (if applicable)
  if (
    validPercentOfTotal &&
    !isPercent &&
    selectedMeasure.validPercentOfTotal
  ) {
    columnNames.splice(nextIndex, 0, `${name}_percent_of_total`);
    nextIndex++;
  }

  // 2. Add absolute change and percent change if time comparison is enabled
  if (timeComparison) {
    // Add absolute change
    columnNames.splice(nextIndex, 0, `${name}_delta`);
    nextIndex++;

    // Add percent change (if measure is not already a percentage)
    if (!isPercent) {
      columnNames.splice(nextIndex, 0, `${name}_delta_perc`);
    }
  }
}

function castUnknownToNumberOrNull(val: unknown): number | null {
  if (typeof val === "number") return val;
  if (val === null || val === undefined) return null;
  console.warn(
    `castUnknownNumberOrNull should only be used to cast unknowns that should be numbers, null, or undefined to numbers or null. Got: ${val}`,
  );
  return val as number;
}

/**
 * This function prepares the data for the dimension table
 * from data returned by the createQueryServiceMetricsViewComparison
 * API.
 *
 */
export function prepareDimensionTableRows(
  queryRows: V1MetricsViewAggregationResponseDataItem[],
  // all of the measures defined for this metrics spec,
  // including those that are not visible
  allMeasuresForSpec: MetricsViewSpecMeasure[],
  activeMeasureName: string,
  dimensionColumn: string,
  addDeltas: boolean,
  addPercentOfTotal: boolean,
  unfilteredTotal: number | { [key: string]: number },
): DimensionTableRow[] {
  if (!queryRows || !queryRows.length) return [];

  const formattersForMeasures: { [key: string]: (val: number) => string } =
    Object.fromEntries(
      allMeasuresForSpec.map((m) => [m.name, createMeasureValueFormatter(m)]),
    );

  const tableRows: DimensionTableRow[] = queryRows
    .filter((row) => row[activeMeasureName] !== undefined)
    .map((row) => {
      // cast is safe since we filtered out rows without measureValues
      const rawVals: [string, number | null][] = allMeasuresForSpec
        .filter((m) => m.name! in row)
        .map((m) => [m.name!, castUnknownToNumberOrNull(row[m.name!])]);

      const formattedVals: [string, string | number | PERC_DIFF][] =
        rawVals.map(([name, val]) => [
          "__formatted_" + name,
          val !== null
            ? formattersForMeasures[name](val)
            : PERC_DIFF.CURRENT_VALUE_NO_DATA,
        ]);

      const rowOut: DimensionTableRow = Object.fromEntries([
        [dimensionColumn, row[dimensionColumn] as string],
        ...rawVals,
        ...formattedVals,
      ]);

      if (addDeltas) {
        // Process deltas for all measures that have comparison data
        allMeasuresForSpec.forEach((measure) => {
          if (!measure.name) return;

          const deltaAbs = row[measure.name + ComparisonDeltaAbsoluteSuffix];
          if (deltaAbs !== undefined) {
            rowOut[`${measure.name}_delta`] =
              castUnknownToNumberOrNull(deltaAbs);
            rowOut[`__formatted_${measure.name}_delta`] =
              deltaAbs !== null
                ? formattersForMeasures[measure.name](deltaAbs as number)
                : PERC_DIFF.PREV_VALUE_NO_DATA;
          }

          const deltaRel = row[measure.name + ComparisonDeltaRelativeSuffix];
          if (deltaRel !== undefined) {
            rowOut[`${measure.name}_delta_perc`] =
              castUnknownToNumberOrNull(deltaRel);
            rowOut[`__formatted_${measure.name}_delta_perc`] =
              deltaRel !== null
                ? formatMeasurePercentageDifference(deltaRel as number)
                : PERC_DIFF.PREV_VALUE_NO_DATA;
          }
        });
      }

      if (addPercentOfTotal) {
        // Calculate percent of total for all measures
        allMeasuresForSpec.forEach((measure) => {
          if (!measure.name) return;
          const value = castUnknownToNumberOrNull(row[measure.name]);
          const total =
            typeof unfilteredTotal === "number"
              ? unfilteredTotal
              : (unfilteredTotal[measure.name] ?? 0);

          if (value === null || total === 0 || !total) {
            rowOut[measure.name + "_percent_of_total"] =
              PERC_DIFF.CURRENT_VALUE_NO_DATA;
            rowOut[`__formatted_${measure.name}_percent_of_total`] =
              PERC_DIFF.CURRENT_VALUE_NO_DATA;
          } else {
            rowOut[measure.name + "_percent_of_total"] = value / total;
            rowOut[`__formatted_${measure.name}_percent_of_total`] =
              formatMeasurePercentageDifference(value / total);
          }
        });
      }

      return rowOut;
    });
  return tableRows;
}
