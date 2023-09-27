import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
import PercentOfTotal from "./PercentOfTotal.svelte";

import { PERC_DIFF } from "../../../components/data-types/type-utils";
import type {
  MetricsViewDimension,
  MetricsViewMeasure,
  V1MetricsViewComparisonRow,
  V1MetricsViewComparisonValue,
  V1MetricsViewFilter,
  V1MetricsViewToplistResponseDataItem,
} from "../../../runtime-client";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
  humanizeDimTableValue,
} from "../humanize-numbers";
import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";

import type { SvelteComponent } from "svelte";
import { getDimensionColumn } from "../dashboard-utils";
import type { DimensionTableRow } from "./dimension-table-types";
import { getFilterForDimension } from "../selectors";
import { SortDirection, SortType } from "../proto-state/derived-types";

/** Returns an updated filter set for a given dimension on search */
export function updateFilterOnSearch(
  filterForDimension,
  searchText,
  dimensionName
) {
  const filterSet = JSON.parse(JSON.stringify(filterForDimension));
  const addNull = "null".includes(searchText);
  if (searchText !== "") {
    let foundDimension = false;

    filterSet["include"].forEach((filter) => {
      if (filter.name === dimensionName) {
        filter.like = [`%${searchText}%`];
        foundDimension = true;
        if (addNull) filter.in.push(null);
      }
    });

    if (!foundDimension) {
      filterSet["include"].push({
        name: dimensionName,
        in: addNull ? [null] : [],
        like: [`%${searchText}%`],
      });
    }
  } else {
    filterSet["include"] = filterSet["include"].filter((f) => f.in.length);
    filterSet["include"].forEach((f) => {
      delete f.like;
    });
  }
  return filterSet;
}

export function getDimensionFilterWithSearch(
  filters: V1MetricsViewFilter,
  searchText: string,
  dimensionName: string
) {
  const filterForDimension = getFilterForDimension(filters, dimensionName);

  return updateFilterOnSearch(filterForDimension, searchText, dimensionName);
}

export function computePercentOfTotal(
  values: V1MetricsViewToplistResponseDataItem[],
  total: number,
  measureName: string
) {
  for (const value of values) {
    if (total === 0 || total === null || total === undefined) {
      value[measureName + "_percent_of_total"] =
        PERC_DIFF.CURRENT_VALUE_NO_DATA;
    } else {
      value[measureName + "_percent_of_total"] =
        formatMeasurePercentageDifference(value[measureName] / total);
    }
  }

  return values;
}

export function getComparisonProperties(
  measureName: string,
  selectedMeasure: MetricsViewMeasure
): {
  component: typeof SvelteComponent;
  type: string;
  format: string;
  description: string;
} {
  if (measureName.includes("_delta_perc"))
    return {
      component: DeltaChangePercentage,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Perc. change over comparison period",
    };
  else if (measureName.includes("_delta")) {
    return {
      component: DeltaChange,
      type: "RILL_CHANGE",
      format: selectedMeasure.format,
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
}

export function estimateColumnCharacterWidths(
  columns: VirtualizedTableColumns[],
  rows: V1MetricsViewToplistResponseDataItem[]
) {
  const columnWidths: { [key: string]: number } = {};
  let largestColumnLength = 0;
  columns.forEach((column, i) => {
    // get values
    const values = rows
      .filter((row) => row[column.name] !== null)
      .map(
        (row) =>
          `${row["__formatted_" + column.name] || row[column.name]}`.length
      );
    values.sort();
    const largest = Math.max(...values);
    columnWidths[column.name] = largest;
    if (i != 0) {
      largestColumnLength = Math.max(
        largestColumnLength,
        column.label?.length || column.name.length
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
  config: VirtualizedTableConfig
) {
  const estimateColumnSize = columns.map((column, i) => {
    if (column.name.includes("delta")) return config.comparisonColumnWidth;
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
            config.minColumnWidth
          )
        )
      : /** if there isn't a longet string length for some reason, let's go with a
         * default column width. We should not be in this state.
         */
        config.defaultColumnWidth;
  });

  const measureColumnSizeSum = estimateColumnSize
    .slice(1)
    .reduce((a, b) => a + b, 0);

  /* Dimension column should expand to cover whole container */
  estimateColumnSize[0] = Math.max(
    containerWidth - measureColumnSizeSum - config.indexWidth,
    estimateColumnSize[0]
  );

  return estimateColumnSize;
}

export function prepareVirtualizedDimTableColumns(
  allMeasures: MetricsViewMeasure[],
  leaderboardMeasureName: string,
  referenceValues: { [key: string]: number },
  dimension: MetricsViewDimension,

  inputColumnNames: string[],
  timeComparison: boolean,
  validPercentOfTotal: boolean,
  sortType: SortType,
  sortDirection: SortDirection
): VirtualizedTableColumns[] {
  const measureNames = allMeasures.map((m) => m.name);
  const selectedMeasure = allMeasures.find(
    (m) => m.name === leaderboardMeasureName
  );
  const dimensionColumn = getDimensionColumn(dimension);

  // copy column names so we don't mutate the original
  const columnNames = [...inputColumnNames];

  addContextColumnNames(
    columnNames,
    timeComparison,
    validPercentOfTotal,
    selectedMeasure
  );
  // Make dimension the first column
  columnNames.unshift(dimensionColumn);

  return columnNames
    .map((name) => {
      const highlight =
        name === selectedMeasure.name ||
        name.endsWith("_delta") ||
        name.endsWith("_delta_perc") ||
        name.endsWith("_percent_of_total");

      let sorted = undefined;
      if (name.endsWith("_delta") && sortType === SortType.DELTA_ABSOLUTE) {
        sorted = sortDirection;
      } else if (
        name.endsWith("_delta_perc") &&
        sortType === SortType.DELTA_PERCENT
      ) {
        sorted = sortDirection;
      } else if (
        name.endsWith("_percent_of_total") &&
        sortType === SortType.PERCENT
      ) {
        sorted = sortDirection;
      } else if (name === selectedMeasure.name && sortType === SortType.VALUE) {
        sorted = sortDirection;
      }

      if (measureNames.includes(name)) {
        // Handle all regular measures
        const measure = allMeasures.find((m) => m.name === name);
        return {
          name,
          type: "INT",
          label: measure?.label || measure?.expression,
          description: measure?.description,
          total: referenceValues[measure.name] || 0,
          enableResize: false,
          format: measure?.format,
          highlight,
          sorted,
        };
      } else if (name === dimensionColumn) {
        // Handle dimension column
        return {
          name,
          type: "VARCHAR",
          label: dimension?.label,
          enableResize: true,
          highlight,
          sorted,
        };
      } else if (selectedMeasure) {
        // Handle delta and delta_perc
        const comparison = getComparisonProperties(name, selectedMeasure);
        return {
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
      return undefined;
    })
    .filter((column) => !!column);
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
  selectedMeasure: MetricsViewMeasure
) {
  const name = selectedMeasure?.name;

  const sortByColumnIndex = columnNames.indexOf(name);
  // Add comparison columns if available
  let percentOfTotalSpliceIndex = 1;
  if (timeComparison) {
    percentOfTotalSpliceIndex = 2;
    columnNames.splice(sortByColumnIndex + 1, 0, `${name}_delta`);

    // Only push percentage delta column if selected measure is not a percentage
    if (selectedMeasure?.format != FormatPreset.PERCENTAGE) {
      percentOfTotalSpliceIndex = 3;
      columnNames.splice(sortByColumnIndex + 2, 0, `${name}_delta_perc`);
    }
  }
  if (validPercentOfTotal) {
    columnNames.splice(
      sortByColumnIndex + percentOfTotalSpliceIndex,
      0,
      `${name}_percent_of_total`
    );
  }
}

/**
 * This function prepares the data for the dimension table
 * from data returned by the createQueryServiceMetricsViewComparisonToplist
 * API.
 *
 */
export function prepareDimensionTableRows(
  queryRows: V1MetricsViewComparisonRow[],
  measures: MetricsViewMeasure[],
  activeMeasureName: string,
  dimensionColumn: string,
  addDeltas: boolean,
  addPercentOfTotal: boolean,
  unfilteredTotal: number
): DimensionTableRow[] {
  if (!queryRows || !queryRows.length) return [];

  const formatMap = Object.fromEntries(
    measures.map((m) => [m.name, m.format as FormatPreset])
  );

  const tableRows: DimensionTableRow[] = queryRows.map((row) => {
    const rawVals: [string, number][] = row.measureValues.map((m) => [
      m.measureName,
      m.baseValue as number,
    ]);

    const formattedVals: [string, string | number][] = row.measureValues.map(
      (m) => [
        "__formatted_" + m.measureName,
        humanizeDimTableValue(m.baseValue as number, formatMap[m.measureName]),
      ]
    );

    const rowOut: DimensionTableRow = Object.fromEntries([
      [dimensionColumn, row.dimensionValue as string],
      ...rawVals,
      ...formattedVals,
    ]);

    if (addDeltas) {
      const activeMeasure = row.measureValues.find(
        (m) => m.measureName === activeMeasureName
      ) as V1MetricsViewComparisonValue;

      rowOut[`${activeMeasureName}_delta`] = humanizeDimTableValue(
        activeMeasure.deltaAbs as number,
        formatMap[activeMeasureName]
      );
      rowOut[`${activeMeasureName}_delta_perc`] =
        formatMeasurePercentageDifference(activeMeasure.deltaRel as number);
    }

    if (addPercentOfTotal) {
      const activeMeasure = row.measureValues.find(
        (m) => m.measureName === activeMeasureName
      ) as V1MetricsViewComparisonValue;
      const value = activeMeasure.baseValue as number;

      if (unfilteredTotal === 0 || !unfilteredTotal) {
        rowOut[activeMeasureName + "_percent_of_total"] =
          PERC_DIFF.CURRENT_VALUE_NO_DATA;
      } else {
        rowOut[activeMeasureName + "_percent_of_total"] =
          formatMeasurePercentageDifference(value / unfilteredTotal);
      }
    }

    return rowOut;
  });
  return tableRows;
}

export function getSelectedRowIndicesFromFilters(
  rows: DimensionTableRow[],
  filters: V1MetricsViewFilter,
  dimensionName: string,
  excludeMode: boolean
): number[] {
  const selectedDimValues =
    ((excludeMode
      ? filters.exclude.find((d) => d.name === dimensionName)?.in
      : filters.include.find((d) => d.name === dimensionName)
          ?.in) as string[]) ?? [];

  return selectedDimValues
    .map((label) => {
      return rows.findIndex((row) => row[dimensionName] === label);
    })
    .filter((i) => i >= 0);
}
