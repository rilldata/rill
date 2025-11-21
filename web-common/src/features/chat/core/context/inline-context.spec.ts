import { ChatContextEntryType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  convertContextToInlinePrompt,
  convertPromptValueToContext,
  type InlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { describe, it, expect } from "vitest";

describe("should convert to and from inline prompt", () => {
  const testCases: {
    title: string;
    ctx: InlineChatContext;
    expectedPrompt: string;
  }[] = [
    {
      title: "metrics view",
      ctx: {
        type: ChatContextEntryType.MetricsView,
        values: ["adbids"],
        label: "",
      },
      expectedPrompt: `<inline>type="metrics_view" name="adbids"</inline>`,
    },

    {
      title: "time range",
      ctx: {
        type: ChatContextEntryType.TimeRange,
        values: ["2025-11-21T00:00:00Z"],
        label: "",
      },
      expectedPrompt: `<inline>type="time_range" range="2025-11-21T00:00:00Z"</inline>`,
    },

    {
      title: "measure",
      ctx: {
        type: ChatContextEntryType.Measures,
        values: ["adbids", "impressions"],
        label: "",
      },
      expectedPrompt: `<inline>type="measures" metrics_view="adbids" name="impressions"</inline>`,
    },
  ];

  for (const { title, ctx, expectedPrompt } of testCases) {
    it(title, () => {
      const prompt = convertContextToInlinePrompt(ctx);
      expect(prompt).toEqual(expectedPrompt);
      const convertedCtx = convertPromptValueToContext(
        expectedPrompt.replace("<inline>", "").replace("</inline>", ""),
      );
      expect(convertedCtx).toEqual(ctx);
    });
  }
});
