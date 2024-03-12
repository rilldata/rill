import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import { prepareMeasureFilterResolutions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { timeControlsState } from "@rilldata/web-common/features/dashboards/state-managers/selectors/time-range";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types";
import type { AtLeast } from "@rilldata/web-common/features/dashboards/state-managers/types";
import {
  forEachExpression,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  MetricsViewSpecMeasureV2,
  V1Expression,
} from "@rilldata/web-common/runtime-client";

export const getMeasureFilterForDimensionIndex = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (dimensionName: string) =>
    dashData.dashboard.dimensionThresholdFilters.findIndex(
      (dtf) => dtf.name === dimensionName,
    );
};

export const additionalMeasures = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  const measures = new Set<string>([dashData.dashboard.leaderboardMeasureName]);
  dashData.dashboard.dimensionThresholdFilters.forEach(({ filter }) => {
    forEachExpression(filter, (e) => {
      if (e.ident) {
        measures.add(e.ident);
      }
    });
  });
  return [...measures];
};

const matchHavingExpressionByName = (e: V1Expression, name: string) =>
  matchExpressionByName(e, name) ||
  (e.cond?.exprs?.length && matchExpressionByName(e.cond?.exprs[0], name));

export const getHavingFilterExpression = (filter: V1Expression, name: string) =>
  filter?.cond?.exprs?.find((e) => matchHavingExpressionByName(e, name));

export const getHavingFilterExpressionIndex = (
  filter: V1Expression,
  name: string,
) =>
  filter?.cond?.exprs?.findIndex((e) => matchHavingExpressionByName(e, name)) ??
  -1;

export const measureHasFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return (measureName: string) =>
    dashData.dashboard.dimensionThresholdFilters.some(
      (dtf) => getHavingFilterExpression(dtf.filter, measureName) !== undefined,
    );
};

export type MeasureFilterItem = {
  dimensionName: string;
  name: string;
  label: string;
  expr?: V1Expression;
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
        dtf.filter,
        dtf.name,
        addedMeasure,
      ),
    );
  }

  return filteredMeasures;
}

export function getMeasureFilterForDimension(
  measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  filter: V1Expression | undefined,
  name = "",
  addedMeasure = new Set<string>(),
) {
  if (!filter) return [];

  const filteredMeasures = new Array<MeasureFilterItem>();

  forEachExpression(filter, (e, depth) => {
    if (depth > 0 || !e.cond?.exprs?.length) {
      return;
    }
    const ident =
      e.cond?.exprs?.[0].ident ?? e.cond?.exprs?.[0].cond?.exprs?.[0].ident;
    if (
      ident === undefined ||
      addedMeasure.has(ident) ||
      !measureIdMap.has(ident)
    ) {
      return;
    }
    const measure = measureIdMap.get(ident);
    if (!measure) {
      return;
    }
    addedMeasure.add(ident);
    filteredMeasures.push({
      dimensionName: name,
      name: ident,
      label: measure.label || measure.expression || ident,
      expr: e,
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

export const getResolvedFilterForMeasureFilters = (
  dashData: DashboardDataSources,
) => {
  return prepareMeasureFilterResolutions(
    dashData.dashboard,
    timeControlsState(dashData),
    dashData.queryClient,
  );
};

export const hasAtLeastOneMeasureFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">,
) => {
  return dashData.dashboard.dimensionThresholdFilters.some(
    (dtf) => dtf.filter.cond?.exprs?.length,
  );
};

export const measureFilterSelectors = {
  measureHasFilter,
  getMeasureFilterItems,
  getAllMeasureFilterItems,
  getResolvedFilterForMeasureFilters,
  hasAtLeastOneMeasureFilter,
};
