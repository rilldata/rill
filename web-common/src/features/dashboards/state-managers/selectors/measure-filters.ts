import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types";
import type { AtLeast } from "@rilldata/web-common/features/dashboards/state-managers/types";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";

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
  measures?: Map<string, MetricsViewSpecMeasure>;
  dimensions?: MetricsViewSpecDimension[];
  filter?: MeasureFilterEntry;
  pinned?: boolean;
  required?: boolean;
  missingRequired?: boolean;
  metricsViewNames?: string[];
};

export const getMeasureFilterItems = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (
    measureIdMap: Map<string, MetricsViewSpecMeasure>,
    pinnedFilters: Set<string> = new Set(),
    requiredFilters: Set<string> = new Set(),
  ) => {
    return getMeasureFilters(
      measureIdMap,
      dashData.dashboard.dimensionThresholdFilters,
      pinnedFilters,
      requiredFilters,
    );
  };
};

export function getMeasureFilters(
  measureIdMap: Map<string, MetricsViewSpecMeasure>,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  pinnedFilters: Set<string>,
  requiredFilters: Set<string>,
) {
  const filteredMeasures = new Array<MeasureFilterItem>();
  const addedMeasure = new Set<string>();

  for (const dtf of dimensionThresholdFilters) {
    const required = dtf.filters[0]
      ? requiredFilters.has(dtf.filters[0].measure)
      : false;
    filteredMeasures.push(
      ...getMeasureFilterForDimension(
        measureIdMap,
        dtf.filters,
        dtf.name,
        addedMeasure,
        required ||
          (dtf.filters[0] ? pinnedFilters.has(dtf.filters[0].measure) : false),
        required,
      ),
    );
  }

  pinnedFilters.forEach((pinnedFilter) => {
    const existing = filteredMeasures.some((mfi) => mfi.name === pinnedFilter);
    if (existing || !measureIdMap.has(pinnedFilter)) return;

    filteredMeasures.push({
      dimensionName: "",
      name: pinnedFilter,
      label: getMeasureDisplayName(measureIdMap.get(pinnedFilter)),
      pinned: true,
    });
  });

  requiredFilters.forEach((requiredFilter) => {
    const existing = filteredMeasures.some(
      (mfi) => mfi.name === requiredFilter,
    );
    if (existing || !measureIdMap.has(requiredFilter)) return;

    filteredMeasures.push({
      dimensionName: "",
      name: requiredFilter,
      label: getMeasureDisplayName(measureIdMap.get(requiredFilter)),
      pinned: true,
      required: true,
    });
  });

  return filteredMeasures;
}

export function getMeasureFilterForDimension(
  measureIdMap: Map<string, MetricsViewSpecMeasure>,
  filters: MeasureFilterEntry[],
  name = "",
  addedMeasure = new Set<string>(),
  pinned = false,
  required = false,
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
      pinned,
      required,
    });
  });

  return filteredMeasures;
}

export const getAllMeasureFilterItems = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (
    measureFilterItems: Array<MeasureFilterItem>,
    measureIdMap: Map<string, MetricsViewSpecMeasure>,
    pinnedFilters: Set<string> = new Set(),
    requiredFilters: Set<string> = new Set(),
  ) => {
    const allMeasureFilterItems = [...measureFilterItems];

    // if the temporary filter is a dimension filter add it
    if (
      dashData.dashboard.temporaryFilterName &&
      measureIdMap.has(dashData.dashboard.temporaryFilterName) &&
      allMeasureFilterItems.every(
        (mfi) => mfi.name !== dashData.dashboard.temporaryFilterName,
      )
    ) {
      const required = requiredFilters.has(
        dashData.dashboard.temporaryFilterName,
      );
      allMeasureFilterItems.push({
        dimensionName: "",
        name: dashData.dashboard.temporaryFilterName,
        label: getMeasureDisplayName(
          measureIdMap.get(dashData.dashboard.temporaryFilterName),
        ),
        pinned:
          required || pinnedFilters.has(dashData.dashboard.temporaryFilterName),
        required,
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
