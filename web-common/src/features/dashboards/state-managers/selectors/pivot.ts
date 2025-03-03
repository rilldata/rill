import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { filteredSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import { PivotChipType } from "../../pivot/types";
import { allDimensions } from "./dimensions";
import type { DashboardDataSources } from "./types";

export const pivotSelectors = {
  showPivot: ({ dashboard }: DashboardDataSources) => dashboard.pivot.active,
  rows: ({ dashboard }: DashboardDataSources) => dashboard.pivot.rows,
  originalColumns: ({ dashboard }: DashboardDataSources) =>
    dashboard.pivot.columns,
  columns: ({ dashboard }: DashboardDataSources) =>
    splitPivotChips(dashboard.pivot.columns),
  isFlat: ({ dashboard }: DashboardDataSources) =>
    dashboard.pivot.tableMode === "flat",
  measures: (dashData: DashboardDataSources) => {
    const measures = filteredSimpleMeasures(dashData)();
    const columnMeasures = splitPivotChips(
      dashData.dashboard.pivot.columns,
    ).measure;

    return measures
      .filter((m) => !columnMeasures.find((c) => c.id === m.name))
      .map((measure) => ({
        id: measure.name || "Unknown",
        title: measure.displayName || measure.name || "Unknown",
        type: PivotChipType.Measure,
        description: measure.description,
      }));
  },
  dimensions: ({
    validMetricsView,
    dashboard,
    validExplore,
  }: DashboardDataSources) => {
    {
      const dimensions = allDimensions({ validMetricsView, validExplore });

      const columnsDimensions = splitPivotChips(
        dashboard.pivot.columns,
      ).dimension;
      const rows = dashboard.pivot.rows;

      return dimensions
        .filter((d) => {
          return !(
            columnsDimensions.find((c) => c.id === d.name) ||
            rows.find((r) => r.id === d.name)
          );
        })
        .map((dimension) => ({
          id: dimension.name || dimension.column || "Unknown",
          title:
            dimension.displayName ||
            dimension.name ||
            dimension.column ||
            "Unknown",
          type: PivotChipType.Dimension,
          description: dimension.description,
        }));
    }
  },
};
