import {
  type ContextRecord,
  ContextTypeData,
  ChatContextEntryType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import type {
  RuntimeServiceCompleteBody,
  V1Message,
} from "@rilldata/web-common/runtime-client";
import { get, type Writable, writable } from "svelte/store";

export class MessageContext {
  public record: Writable<ContextRecord> = writable({});

  public static fromMessage(message: V1Message) {
    const context = new MessageContext();

    if (!message.contentType || !message.contentData) {
      return context;
    }
    try {
      const rawContext = JSON.parse(message.contentData);
      if (!rawContext?.analyst_agent_args) return context;
      Object.entries(ContextTypeData).forEach(([type, data]) => {
        const value = data.deserializer(rawContext?.analyst_agent_args);
        if (value) {
          context.set(type as keyof ContextRecord, value as any);
        }
      });
    } catch {
      context.clear();
    }

    return context;
  }

  public set<T extends keyof ContextRecord>(type: T, value: ContextRecord[T]) {
    this.record.update((r) => {
      return {
        ...r,
        [type]: value,
      };
    });
  }

  public delete(type: ChatContextEntryType) {
    this.record.update((r) => {
      delete r[type];
      return r;
    });
  }

  public clear() {
    this.record.set({});
  }

  public getRequestContext(): RuntimeServiceCompleteBody | undefined {
    const r = get(this.record);
    if (Object.keys(r).length === 0) return undefined;

    let context: RuntimeServiceCompleteBody = {};
    Object.entries(r).forEach(([t, v]) => {
      context = {
        ...context,
        ...ContextTypeData[t].serializer(v),
      };
    });

    return context;
  }
}
