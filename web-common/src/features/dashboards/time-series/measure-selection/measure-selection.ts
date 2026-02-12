import {
  InlineContextType,
  type InlineContext,
  convertContextToInlinePrompt,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { get, writable } from "svelte/store";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { derived } from "svelte/store";

export class MeasureSelection {
  public readonly measure = writable<string | null>(null);
  public readonly start = writable<Date | null>(null);
  public readonly end = writable<Date | null>(null);

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

  public clear() {
    this.measure.set(null);
    this.start.set(null);
    this.end.set(null);
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
      value: measure,
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

  public getEnabledStore() {
    return derived(
      [featureFlags.dashboardChat, getExploreNameStore()],
      ([dashboardChat, exploreName]) => {
        return Boolean(dashboardChat && exploreName);
      },
    );
  }
}

export const measureSelection = new MeasureSelection();
