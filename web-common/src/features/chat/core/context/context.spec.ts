import { ConversationContextType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
import { get } from "svelte/store";
import { describe, it, expect } from "vitest";

describe("chat context", () => {
  it("should set and update context", () => {
    const context = new ConversationContext();

    context.set(ConversationContextType.MetricsView, "mv_0");
    context.set(ConversationContextType.Measures, "mes_0");

    expect(get(context.data)).toEqual([
      { type: ConversationContextType.MetricsView, value: "mv_0" },
      { type: ConversationContextType.Measures, value: "mes_0" },
    ]);
    expect(get(context.record)).toEqual({
      [ConversationContextType.MetricsView]: "mv_0",
      [ConversationContextType.Measures]: "mes_0",
    });

    context.set(ConversationContextType.MetricsView, "mv_1");
    context.set(ConversationContextType.Measures, "mes_1");

    expect(get(context.data)).toEqual([
      { type: ConversationContextType.MetricsView, value: "mv_1" },
      { type: ConversationContextType.Measures, value: "mes_1" },
    ]);
    expect(get(context.record)).toEqual({
      [ConversationContextType.MetricsView]: "mv_1",
      [ConversationContextType.Measures]: "mes_1",
    });
  });
});
