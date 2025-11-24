import { ChatContextEntryType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { get, writable } from "svelte/store";
import {
  convertContextToInlinePrompt,
  type InlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

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

  public startAnomalyExplanationChat(metricsViewName: string) {
    if (!this.hasSelection()) return;
    const measure = get(this.measure)!;

    const measureMention = convertContextToInlinePrompt({
      type: ChatContextEntryType.Measures,
      values: [metricsViewName, measure],
    });

    const start = get(this.start)?.toISOString();
    const end = get(this.end)?.toISOString();
    if (!start) return;

    const timeRangeCtx = <InlineChatContext>{
      type: ChatContextEntryType.TimeRange,
      values: [],
    };
    if (end) {
      timeRangeCtx.values.push(`${start} to ${end}`);
    } else {
      timeRangeCtx.values.push(start);
    }
    const timeRangeMention = convertContextToInlinePrompt(timeRangeCtx);

    const prompt =
      `Explain what drives ${measureMention}, ${timeRangeMention}. ` +
      `What dimensions have noticeably changed, as compared to other time windows?`;

    sidebarActions.startChat(prompt);
  }
}

export const measureSelection = new MeasureSelection();
