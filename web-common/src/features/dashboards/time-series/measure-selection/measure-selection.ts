import { ChatContextEntryType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { convertContextToInlinePrompt } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { get, writable } from "svelte/store";

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

  public startAnomalyExplanationChat() {
    if (!this.hasSelection()) return;
    const measure = get(this.measure)!;

    const measureMention = convertContextToInlinePrompt({
      type: ChatContextEntryType.Measures,
      value: measure,
      subValue: null,
      label: "",
    });

    const start = get(this.start)?.toISOString();
    const end = get(this.end)?.toISOString();
    if (!start) return;

    const timeRangeCtx = {
      type: ChatContextEntryType.TimeRange,
      value: "",
      subValue: null,
      label: "",
    };
    if (end) {
      timeRangeCtx.value = `${start} to ${end}`;
    } else {
      timeRangeCtx.value = start;
    }
    const timeRangeMention = convertContextToInlinePrompt(timeRangeCtx);

    const prompt =
      `Explain what drives ${measureMention}, ${timeRangeMention}. ` +
      `What dimensions have noticeably changed, as compared to other time windows?`;

    sidebarActions.startChat(prompt);
  }
}

export const measureSelection = new MeasureSelection();
