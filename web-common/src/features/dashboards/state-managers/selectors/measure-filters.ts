import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types";
import type { AtLeast } from "@rilldata/web-common/features/dashboards/state-managers/types";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { type MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

export const measureHasFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (measureName: string) =>
    dashData.dashboard.dimensionThresholdFilters.some((dtf) =>
      dtf.filters.some((f) => f.measure === measureName),
    );
};

export type MeasureFilterItem = {
  dimensionName: string;
  name: string;
  label: string;
  filter?: MeasureFilterEntry;
};
export const getMeasureFilterItems = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (measureIdMap: Map<string, MetricsViewSpecMeasureV2>) => {
    return getMeasureFilters(
      measureIdMap,
      dashData.dashboard.dimensionThresholdFilters,
    );
  };
};

export function getMeasureFilters(
  measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  dimensionThresholdFilters: DimensionThresholdFilter[],
) {
  const filteredMeasures = new Array<MeasureFilterItem>();
  const addedMeasure = new Set<string>();

  for (const dtf of dimensionThresholdFilters) {
    filteredMeasures.push(
      ...getMeasureFilterForDimension(
        measureIdMap,
        dtf.filters,
        dtf.name,
        addedMeasure,
      ),
    );
  }

  return filteredMeasures;
}

export function getMeasureFilterForDimension(
  measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  filters: MeasureFilterEntry[],
  name = "",
  addedMeasure = new Set<string>(),
) {
  if (!filters.length) return [];

  const filteredMeasures = new Array<MeasureFilterItem>();

  filters.forEach((filter) => {
    if (addedMeasure.has(filter.measure)) {
      return;
    }

    const measure = measureIdMap.get(filter.measure);
    if (!measure) {
      return;
    }
    addedMeasure.add(filter.measure);
    filteredMeasures.push({
      dimensionName: name,
      name: filter.measure,
      label: measure.displayName || measure.expression || filter.measure,
      filter,
    });
  });

  return filteredMeasures;
}

export const getAllMeasureFilterItems = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (
    measureFilterItems: Array<MeasureFilterItem>,
    measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  ) => {
    const allMeasureFilterItems = [...measureFilterItems];

    // if the temporary filter is a dimension filter add it
    if (
      dashData.dashboard.temporaryFilterName &&
      measureIdMap.has(dashData.dashboard.temporaryFilterName)
    ) {
      allMeasureFilterItems.push({
        dimensionName: "",
        name: dashData.dashboard.temporaryFilterName,
        label: getMeasureDisplayName(
          measureIdMap.get(dashData.dashboard.temporaryFilterName),
        ),
      });
    }

    return allMeasureFilterItems;
  };
};

export const measureFilterSelectors = {
  measureHasFilter,
  getMeasureFilterItems,
  getAllMeasureFilterItems,
};
