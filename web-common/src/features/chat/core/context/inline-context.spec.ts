import {
  InlineContextType,
  type InlineContext,
  convertPromptValueToContext,
  convertContextToInlinePrompt,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import { describe, it, expect } from "vitest";

describe("should convert to and from inline prompt", () => {
  const testCases: {
    title: string;
    ctx: InlineContext;
    expectedPrompt: string;
  }[] = [
    {
      title: "metrics view",
      ctx: {
        type: InlineContextType.MetricsView,
        metricsView: "adbids",
        value: "adbids",
      },
      expectedPrompt: `<chat-reference>type="metricsView" metricsView="adbids"</chat-reference>`,
    },

    {
      title: "time range",
      ctx: {
        type: InlineContextType.TimeRange,
        timeRange: "2025-11-21T00:00:00Z",
        value: "2025-11-21T00:00:00Z",
      },
      expectedPrompt: `<chat-reference>type="timeRange" timeRange="2025-11-21T00:00:00Z"</chat-reference>`,
    },

    {
      title: "measure",
      ctx: {
        type: InlineContextType.Measure,
        metricsView: "adbids",
        measure: "impressions",
        value: "impressions",
      },
      expectedPrompt: `<chat-reference>type="measure" metricsView="adbids" measure="impressions"</chat-reference>`,
    },

    {
      title: "dimension",
      ctx: {
        type: InlineContextType.Dimension,
        metricsView: "adbids",
        dimension: "publisher",
        value: "publisher",
      },
      expectedPrompt: `<chat-reference>type="dimension" metricsView="adbids" dimension="publisher"</chat-reference>`,
    },

    {
      title: "model",
      ctx: {
        type: InlineContextType.Model,
        model: "adbids_model",
        value: "adbids_model",
      },
      expectedPrompt: `<chat-reference>type="model" model="adbids_model"</chat-reference>`,
    },

    {
      title: "column",
      ctx: {
        type: InlineContextType.Column,
        model: "adbids_model",
        column: "pub",
        value: "pub",
      },
      expectedPrompt: `<chat-reference>type="column" model="adbids_model" column="pub"</chat-reference>`,
    },
  ];

  for (const { title, ctx, expectedPrompt } of testCases) {
    it(title, () => {
      const prompt = convertContextToInlinePrompt(ctx);
      expect(prompt).toEqual(expectedPrompt);
      const convertedCtx = convertPromptValueToContext(
        expectedPrompt
          .replace("<chat-reference>", "")
          .replace("</chat-reference>", ""),
      );
      expect(convertedCtx).toEqual(ctx);
    });
  }
});
