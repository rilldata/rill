import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { getChartData } from "@rilldata/web-common/features/components/charts/data-provider";
import { type Readable } from "svelte/store";
import type { ChartDataResult } from "../../../components/charts/types";

/**
 * Convenience wrapper for using getChartData with canvas context.
 */
export function getChartDataForCanvas(
  ctx: CanvasStore,
  component: BaseChart<ChartSpec>,
  config: ChartSpec,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
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
  });
}
