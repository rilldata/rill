import {
  type ConversationContextEntry,
  ConversationContextType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
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

  public getContexts() {
    const measure = get(this.measure);
    const start = get(this.start);
    const end = get(this.end);

    const contextEntries: ConversationContextEntry[] = [];

    if (measure) {
      contextEntries.push({
        type: ConversationContextType.Measure,
        value: measure,
      });
    }

    if (start && end) {
      const timeRange = `${start.toISOString()} - ${end.toISOString()}`;
      contextEntries.push({
        type: ConversationContextType.TimeRange,
        value: timeRange,
      });
    } else if (start) {
      const timestamp = start.toISOString();
      contextEntries.push({
        type: ConversationContextType.TimeRange,
        value: timestamp,
      });
    }

    return contextEntries;
  }
}

export const measureSelection = new MeasureSelection();
