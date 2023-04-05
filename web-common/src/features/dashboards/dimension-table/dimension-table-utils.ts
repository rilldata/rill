import { PERC_DIFF } from "../../../components/data-types/type-utils";
import type {
  MetricsViewMeasure,
  V1MetricsViewToplistResponse,
  V1MetricsViewToplistResponseDataItem,
} from "../../../runtime-client";
import {
  formatMeasurePercentageDifference,
  NicelyFormattedTypes,
} from "../humanize-numbers";
import DeltaChange from "./DeltaChange.svelte";
import DeltaChangePercentage from "./DeltaChangePercentage.svelte";

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
      // console.log(filter.name, dimensionName);
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
  values
) {
  if (!values || !values.length) return filterForDimension;
  const filterValues = values.map((v) => v[dimensionName]);
  getFilterForComparsion(filterForDimension, dimensionName, filterValues);
}

// Custom sort that implements the following logic:
// Input - [ "measure_1", "measure_10", "measure_1_delta", "measure_1_delta_perc", "measure_2", "measure_0", "measure_20" ]
// Output -  [ "measure_0", "measure_1", "measure_1_delta", "measure_1_delta_perc", "measure_2", "measure_10", "measure_20" ]
export function customSortMeasures(measure1: string, measure2: string) {
  const a_num = parseFloat(
    measure1
      .replace("measure_", "")
      .replace("_delta", ".1")
      .replace("_delta_perc", ".2")
  );
  const b_num = parseFloat(
    measure2
      .replace("measure_", "")
      .replace("_delta", ".1")
      .replace("_delta_perc", ".2")
  );

  return a_num - b_num;
}

/** Takes previous and current data to construct comparison data
 * with fields named measure_x_delta and measure_x_delta_perc */
export function computeComparisonValues(
  comparisonData: V1MetricsViewToplistResponse,
  values: V1MetricsViewToplistResponseDataItem[]
) {
  if (comparisonData?.meta?.length !== 2) return values;

  const measureName = comparisonData?.meta[0].name.includes("measure_")
    ? comparisonData?.meta[0].name
    : comparisonData?.meta[1].name;
  const dimensionName = comparisonData?.meta[0].name.includes("measure_")
    ? comparisonData?.meta[1].name
    : comparisonData?.meta[0].name;

  const dimensionToValueMap = new Map(
    comparisonData?.data?.map((obj) => [obj[dimensionName], obj[measureName]])
  );

  for (const value of values) {
    const prevValue = dimensionToValueMap.get(value[dimensionName]);

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

export function getComparisonProperties(
  measureName: string,
  selectedMeasure: MetricsViewMeasure
) {
  if (measureName.includes("_delta_perc"))
    return {
      label: DeltaChangePercentage,
      type: "RILL_PERCENTAGE_CHANGE",
      format: NicelyFormattedTypes.PERCENTAGE,
      description: "Percentage change over compared period",
    };
  else if (measureName.includes("_delta")) {
    return {
      label: DeltaChange,
      type: "RILL_CHANGE",
      format: selectedMeasure.format,
      description: "Change over compared period",
    };
  }
}
