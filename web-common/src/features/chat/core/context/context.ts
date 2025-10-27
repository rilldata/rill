import {
  ContextKeyToTypeMap,
  type ContextRecord,
  ContextTypeData,
  type ConversationContextEntry,
  ConversationContextType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import type {
  V1CompletionMessageContext,
  V1Message,
} from "@rilldata/web-common/runtime-client";
import { snakeToCamel } from "@rilldata/web-common/lib/string-utils.ts";
import { get, type Writable, writable } from "svelte/store";

export class ConversationContext {
  public data: Writable<ConversationContextEntry[]> = writable([]);
  public record: Writable<ContextRecord> = writable({});

  public set(type: ConversationContextType, value: string) {
    this.data.update((c) => {
      let idx = -1;
      let exists = false;
      for (idx = 0; idx < c.length && c[idx].type <= type; idx++) {
        exists = c[idx].type === type;
        if (exists) break;
      }

      const deleteCount = exists ? 1 : 0;
      c.splice(idx, deleteCount, { type, value });
      return c;
    });
    this.updateRecord();
  }

  public delete(type: ConversationContextType) {
    this.data.update((c) => {
      return c.filter((e) => e.type !== type);
    });
    this.updateRecord();
  }

  public clear() {
    this.data.set([]);
    this.updateRecord();
  }

  public override(context: ConversationContextEntry[]) {
    this.data.set([]);
    // Reuse set code to make sure we dedupe and maintain the order of entries.
    context.forEach((c) => this.set(c.type, c.value as any));
    this.updateRecord();
  }

  public parseContext(message: V1Message) {
    if (!message.contentType || !message.contentData) {
      this.clear();
      return;
    }
    try {
      const contentData = JSON.parse(message.contentData);
      const context: ConversationContextEntry[] = [];

      Object.entries(contentData.context ?? {}).forEach(([key, value]) => {
        const camelKey = snakeToCamel(key);
        const type = ContextKeyToTypeMap[camelKey];
        if (type) {
          context.push({ type, value: value as any });
        }
      });

      this.override(context);
    } catch {
      this.clear();
    }
  }

  public getRequestContext(): V1CompletionMessageContext | undefined {
    const c = get(this.data);
    if (Object.keys(c).length === 0) return undefined;

    const context: V1CompletionMessageContext = {};
    c.forEach((e) => {
      context[ContextTypeData[e.type].key] = e.value as any;
    });

    if (context.metricsView) {
      context.explore = context.metricsView;
    }

    return context;
  }

  private updateRecord() {
    this.record.set(
      get(this.data).reduce((r, v) => {
        r[v.type] = v.value;
        return r;
      }, {}),
    );
  }
}
