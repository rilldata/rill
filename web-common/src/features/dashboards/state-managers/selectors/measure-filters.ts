import { matchExpressionByName } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { DashboardDataSources } from "@rilldata/web-common/features/dashboards/state-managers/selectors/types";
import type { AtLeast } from "@rilldata/web-common/features/dashboards/state-managers/types";
import { forEachExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import type { V1Expression } from "@rilldata/web-common/runtime-client";

export const getHavingFilterExpression = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((name: string) => V1Expression | undefined) => {
  return (name: string) =>
    dashData.dashboard.havingFilter?.cond?.exprs?.find((e) =>
      matchExpressionByName(e, name)
    );
};

export const getHavingFilterExpressionIndex = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
): ((name: string) => number | undefined) => {
  return (name: string) =>
    dashData.dashboard.havingFilter?.cond?.exprs?.findIndex(
      (e) =>
        matchExpressionByName(e, name) ||
        (e.cond?.exprs?.length && matchExpressionByName(e.cond?.exprs[0], name))
    );
};

export const measureHasFilter = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
) => {
  return (measureName: string) =>
    getHavingFilterExpression(dashData)(measureName) !== undefined;
};

export type FilteredMeasure = {
  name: string;
  label: string;
  expr: V1Expression;
};
export const getAllMeasureFilters = (
  dashData: AtLeast<DashboardDataSources, "dashboard">
) => {
  return (measures: Map<string, MetricsViewSpecMeasureV2>) => {
    const filteredMeasures = new Array<FilteredMeasure>();
    const addedMeasure = new Set<string>();
    forEachExpression(dashData.dashboard.havingFilter, (e) => {
      if (!e.cond?.exprs?.length) {
        return;
      }
      const ident = e.cond?.exprs?.[0].ident;
      if (
        ident === undefined ||
        addedMeasure.has(ident) ||
        !measures.has(ident)
      ) {
        return;
      }
      const measure = measures.get(ident);
      if (!measure) {
        return;
      }
      addedMeasure.add(ident);
      filteredMeasures.push({
        name: ident,
        label: measure.label || measure.expression || ident,
        expr: e,
      });
    });
    return filteredMeasures;
  };
};

export const measureFilterSelectors = {
  measureHasFilter,
  getAllMeasureFilters,
};
