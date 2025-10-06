import {
  type ContextRecord,
  ContextTypeData,
  extractContextEntry,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  type ConversationContextEntry,
  ConversationContextType,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { get, type Writable, writable } from "svelte/store";

const contextRegex = /\s*<context>([\s\S]*?)<\/context>/m;

export class ConversationContext {
  public data: Writable<ConversationContextEntry[]> = writable([]);
  public record: Writable<ContextRecord> = writable({});

  public static cleanContext(prompt: string) {
    return prompt.replace(contextRegex, "");
  }

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
    this.data.set(context);
    this.updateRecord();
  }

  public parseMessages(messages: V1Message[]) {
    // Find the last message from user.
    const userMessage = messages.findLast((m) => m.role === "user");
    const prompt = userMessage?.content?.[0]?.text;
    if (!prompt) return;

    // Extract the context part
    const contextMatch = contextRegex.exec(prompt);
    if (!contextMatch?.[1]) return;

    // Convert the context parts and set
    contextMatch[1].split("\n").forEach((line) => {
      const contextEntry = extractContextEntry(line);
      if (!contextEntry) return;
      this.set(contextEntry.type, contextEntry.value);
    });
  }

  public toString() {
    const c = get(this.data);

    if (c.length === 0) return "";

    const contextPart = get(this.data)
      .map((e) => `${ContextTypeData[e.type].prompt}: "${e.value}"`)
      .join("\n");

    return `\n\n<context>\n${contextPart}\n</context>`;
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
