import {
  type ConversationContextEntry,
  ConversationContextType,
} from "@rilldata/web-common/features/chat/core/types.ts";
import { get, type Writable, writable } from "svelte/store";

const ContextTypeData: Record<ConversationContextType, { label: string }> = {
  [ConversationContextType.MetricsView]: {
    label: "Metrics View",
  },
  [ConversationContextType.TimeRange]: {
    label: "Time Range",
  },
  [ConversationContextType.Measures]: {
    label: "Measures",
  },
};

export class ConversationContext {
  public context: Writable<ConversationContextEntry[]> = writable([]);

  public set(type: ConversationContextType, value: string) {
    this.context.update((c) => {
      let idx = -1;
      let exists = false;
      for (idx = 0; idx < c.length && c[idx].type <= type; idx++) {
        exists = c[idx].type === type;
      }

      const deleteCount = exists ? 1 : 0;
      c.splice(idx, deleteCount, { type, value });
      return c;
    });
  }

  public delete(type: ConversationContextType) {
    this.context.update((c) => {
      return c.filter((e) => e.type !== type);
    });
  }

  public clear() {
    this.context.set([]);
  }

  public override(context: ConversationContextEntry[]) {
    this.context.set(context);
  }

  public toString() {
    const c = get(this.context);

    if (c.length === 0) return "";

    const contextPart = get(this.context)
      .map((e) => `${ContextTypeData[e.type].label}: ${e.value}`)
      .join("\n");

    return `\n\n${contextPart}`;
  }
}
