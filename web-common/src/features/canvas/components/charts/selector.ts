import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { getChartData } from "@rilldata/web-common/features/components/charts/data-provider";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { Readable } from "svelte/store";
import type { ChartDataResult } from "../../../components/charts/types";

/**
 * Convenience wrapper for using getChartData with canvas context.
 * @param themeModeStore - Reactive store tracking if theme mode is dark (for light/dark toggle)
 */
export function getChartDataForCanvas(
  ctx: CanvasStore,
  component: BaseChart<CanvasChartSpec>,
  config: CanvasChartSpec,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
  themeModeStore: Readable<boolean>,
): Readable<ChartDataResult> {
  const chartDataQuery = component.createChartDataQuery(
    ctx,
    timeAndFilterStore,
  );

  return getChartData({
    config,
    chartDataQuery,
    getDomainValues: () => component.getChartDomainValues(),
    metricsView: ctx.canvasEntity.metricsView,
    themeStore: ctx.canvasEntity.theme,
    timeAndFilterStore,
    themeModeStore,
  });
}
