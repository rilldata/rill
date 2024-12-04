import { filteredSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { DashboardDataSources } from "./types";
import { PivotChipType } from "../../pivot/types";
import { allDimensions } from "./dimensions";

export const pivotSelectors = {
  showPivot: ({ dashboard }: DashboardDataSources) => dashboard.pivot.active,
  rows: ({ dashboard }: DashboardDataSources) => dashboard.pivot.rows,
  columns: ({ dashboard }: DashboardDataSources) => dashboard.pivot.columns,
  measures: (dashData: DashboardDataSources) => {
    const measures = filteredSimpleMeasures(dashData)();
    const columns = dashData.dashboard.pivot.columns;

    return measures
      .filter((m) => !columns.measure.find((c) => c.id === m.name))
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

      const columns = dashboard.pivot.columns;
      const rows = dashboard.pivot.rows;

      return dimensions
        .filter((d) => {
          return !(
            columns.dimension.find((c) => c.id === d.name) ||
            rows.dimension.find((r) => r.id === d.name)
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
