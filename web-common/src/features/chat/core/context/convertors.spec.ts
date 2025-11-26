import {
  ChatContextEntryType,
  type InlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  convertContextToInlinePrompt,
  convertPromptValueToContext,
} from "@rilldata/web-common/features/chat/core/context/convertors.ts";
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
        metricsView: "adbids",
      },
      expectedPrompt: `<inline>type="metricsView" metricsView="adbids"</inline>`,
    },

    {
      title: "time range",
      ctx: {
        type: ChatContextEntryType.TimeRange,
        timeRange: "2025-11-21T00:00:00Z",
      },
      expectedPrompt: `<inline>type="timeRange" timeRange="2025-11-21T00:00:00Z"</inline>`,
    },

    {
      title: "measure",
      ctx: {
        type: ChatContextEntryType.Measure,
        metricsView: "adbids",
        measure: "impressions",
      },
      expectedPrompt: `<inline>type="measure" metricsView="adbids" measure="impressions"</inline>`,
    },

    {
      title: "dimension",
      ctx: {
        type: ChatContextEntryType.Dimension,
        metricsView: "adbids",
        dimension: "publisher",
      },
      expectedPrompt: `<inline>type="dimension" metricsView="adbids" dimension="publisher"</inline>`,
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
