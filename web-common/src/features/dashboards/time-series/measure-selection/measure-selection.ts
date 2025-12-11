import {
  InlineContextType,
  type InlineContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { get, writable } from "svelte/store";
import { convertContextToInlinePrompt } from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";
import type {
  GraphicScale,
  SimpleDataGraphicConfiguration,
} from "@rilldata/web-common/components/data-graphic/state/types";

export class MeasureSelection {
  public readonly measure = writable<string | null>(null);
  public readonly start = writable<Date | null>(null);
  public readonly end = writable<Date | null>(null);
  // Calculated x,y coordinates of the measure selection point.
  // This uses GraphicScale and SimpleDataGraphicConfiguration that is not available outside `SimpleDataGraphic`.
  public readonly x = writable<number | null>(null);
  public readonly y = writable<number | null>(null);

  public setStart(measure: string, start: Date) {
    this.measure.set(measure);
    this.start.set(start);
    this.end.set(null);
  }

  public setRange(measure: string, start: Date, end: Date) {
    this.measure.set(measure);
    this.start.set(start);
    this.end.set(end);
  }

  /**
   * Calculate the point on the graph where the measure selection should be drawn.
   * scaler and config are only available within the `MeasureSelection` wrapped in `SimpleDataGraphic`.
   * But it is used in `ExplainButton` that is outside the `SimpleDataGraphic` wrapper to avoid click issues.
   * That is why this updates the x & y stores directly.
   *
   * @param start
   * @param end
   * @param scaler
   * @param config
   */
  public calculatePoint(
    start: Date,
    end: Date | null,
    scaler: GraphicScale,
    config: SimpleDataGraphicConfiguration,
  ) {
    const startX = scaler(start);
    const endX = end ? scaler(end) : startX;

    const x = Math.round((startX + endX) / 2);
    const y = config.bottom;

    this.x.set(x);
    this.y.set(y);
  }

  public clear() {
    this.measure.set(null);
    this.start.set(null);
    this.end.set(null);
    this.x.set(null);
    this.y.set(null);
  }

  public hasSelection() {
    return Boolean(get(this.measure));
  }

  public isRangeSelection() {
    return Boolean(get(this.end));
  }

  public startAnomalyExplanationChat(metricsView: string) {
    if (!this.hasSelection()) return;
    const measure = get(this.measure)!;

    const measureMention = convertContextToInlinePrompt({
      type: InlineContextType.Measure,
      metricsView,
      measure,
    });

    const start = get(this.start)?.toISOString();
    const end = get(this.end)?.toISOString();
    if (!start) return;

    const timeRangeCtx = <InlineContext>{
      type: InlineContextType.TimeRange,
    };
    if (end) {
      timeRangeCtx.timeRange = `${start} to ${end}`;
    } else {
      timeRangeCtx.timeRange = start;
    }
    const timeRangeMention = convertContextToInlinePrompt(timeRangeCtx);

    const prompt =
      `Explain what drives ${measureMention}, ${timeRangeMention}. ` +
      `What dimensions have noticeably changed, as compared to other time windows?`;

    sidebarActions.startChat(prompt);
  }
}

export const measureSelection = new MeasureSelection();
