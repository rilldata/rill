import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
import { PERC_DIFF } from "../../../components/data-types/type-utils";
import type {
  MetricsViewMeasure,
  V1MetricsViewToplistResponse,
  V1MetricsViewToplistResponseDataItem,
} from "../../../runtime-client";
import {
  FormatPreset,
  formatMeasurePercentageDifference,
} from "../humanize-numbers";
import PercentOfTotal from "./PercentOfTotal.svelte";
import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
import type { VirtualizedTableConfig } from "@rilldata/web-common/components/virtualized-table/types";

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

/** Returns a filter set which takes the current filter set for the
 * dimension table and updates it to get all the same dimension values
 * in a previous period */
export function getFilterForComparsion(
  filterForDimension,
  dimensionName,
  filterValues
) {
  const comparisonFilterSet = JSON.parse(JSON.stringify(filterForDimension));

  if (!filterValues.length) return comparisonFilterSet;

  let foundDimension = false;
  comparisonFilterSet["include"].forEach((filter) => {
    if (filter.name === dimensionName) {
      foundDimension = true;
      filter.in = filterValues;
    }
  });

  if (!foundDimension) {
    comparisonFilterSet["include"].push({
      name: dimensionName,
      in: filterValues,
    });
  }
  return comparisonFilterSet;
}

export function getFilterForComparisonTable(
  filterForDimension,
  dimensionName,
  dimensionColumn,
  values
) {
  if (!values || !values.length) return filterForDimension;
  const filterValues = values.map((v) => v[dimensionColumn]);
  return getFilterForComparsion(
    filterForDimension,
    dimensionName,
    filterValues
  );
}

/** Takes previous and current data to construct comparison data
 * with fields named measure_x_delta and measure_x_delta_perc */
export function computeComparisonValues(
  comparisonData: V1MetricsViewToplistResponse,
  values: V1MetricsViewToplistResponseDataItem[],
  dimensionName: string,
  dimensionColumn: string,
  measureName: string
) {
  if (comparisonData?.meta?.length !== 2) return values;

  const dimensionToValueMap = new Map(
    comparisonData?.data?.map((obj) => [obj[dimensionColumn], obj[measureName]])
  );

  for (const value of values) {
    const prevValue = dimensionToValueMap.get(value[dimensionColumn]);

    if (prevValue === undefined) {
      value[measureName + "_delta"] = null;
      value[measureName + "_delta_perc"] = PERC_DIFF.PREV_VALUE_NO_DATA;
    } else if (prevValue === null) {
      value[measureName + "_delta"] = null;
      value[measureName + "_delta_perc"] = PERC_DIFF.PREV_VALUE_NULL;
    } else if (prevValue === 0) {
      value[measureName + "_delta"] = value[measureName];
      value[measureName + "_delta_perc"] = PERC_DIFF.PREV_VALUE_ZERO;
    } else {
      value[measureName + "_delta"] = value[measureName] - prevValue;
      value[measureName + "_delta_perc"] = formatMeasurePercentageDifference(
        (value[measureName] - prevValue) / prevValue
      );
    }
  }

  return values;
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
) {
  if (measureName.includes("_delta_perc"))
    return {
      label: DeltaChangePercentage,
      type: "RILL_PERCENTAGE_CHANGE",
      format: FormatPreset.PERCENTAGE,
      description: "Perc. change over comparison period",
    };
  else if (measureName.includes("_delta")) {
    return {
      label: DeltaChange,
      type: "RILL_CHANGE",
      format: selectedMeasure.format,
      description: "Change over comparison period",
    };
  } else if (measureName.includes("_percent_of_total")) {
    return {
      label: PercentOfTotal,
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
