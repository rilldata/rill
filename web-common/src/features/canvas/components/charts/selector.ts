import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { getChartData } from "@rilldata/web-common/features/components/charts/data-provider";
import type { Readable } from "svelte/store";
import type { ChartDataResult } from "../../../components/charts/types";
import type { ExpressionState } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

/**
 * Convenience wrapper for using getChartData with canvas context.
 * @param ctx
 * @param component
 * @param config
 * @param filterStore
 * @param timeControlStore
 * @param themeModeStore - Reactive store tracking if theme mode is dark (for light/dark toggle)
 * @param visible
 */
export function getChartDataForCanvas(
  ctx: CanvasStore,
  component: BaseChart<CanvasChartSpec>,
  config: CanvasChartSpec,
  filterStore: Readable<ExpressionState>,
  timeControlStore: Readable<TimeControlState>,
  themeModeStore: Readable<boolean>,
  visible: Readable<boolean>,
): Readable<ChartDataResult> {
  const chartDataQuery = component.createChartDataQuery(
    ctx,
    filterStore,
    timeControlStore,
    visible,
  );

  return getChartData({
    config,
    chartDataQuery,
    getDomainValues: () => component.getChartDomainValues(),
    metricsView: ctx.canvasEntity.metricsView,
    themeStore: ctx.canvasEntity.theme,
    timeControlStore,
    themeModeStore,
  });
}
