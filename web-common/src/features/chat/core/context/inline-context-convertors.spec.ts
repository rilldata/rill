import {
  InlineContextType,
  type InlineContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import {
  convertContextToInlinePrompt,
  convertPromptValueToContext,
  convertPromptWithInlineContextToComponents,
} from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";
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
      },
      expectedPrompt: `<chat-reference>type="metricsView" metricsView="adbids"</chat-reference>`,
    },

    {
      title: "time range",
      ctx: {
        type: InlineContextType.TimeRange,
        timeRange: "2025-11-21T00:00:00Z",
      },
      expectedPrompt: `<chat-reference>type="timeRange" timeRange="2025-11-21T00:00:00Z"</chat-reference>`,
    },

    {
      title: "measure",
      ctx: IMPRESSIONS_INLINE_CONTEXT,
      expectedPrompt: IMPRESSIONS_INLINE_PROMPT,
    },

    {
      title: "dimension",
      ctx: PUBLISHER_INLINE_CONTEXT,
      expectedPrompt: PUBLISHER_INLINE_PROMPT,
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

describe("should convert prompt to correct text or components", () => {
  const testCases: {
    title: string;
    prompt: string;
    textOrComponents: ReturnType<
      typeof convertPromptWithInlineContextToComponents
    >;
  }[] = [
    {
      title: "Single line valid prompt",
      prompt: `What are the top ${PUBLISHER_INLINE_PROMPT} and ${DOMAIN_INLINE_PROMPT}?`,
      textOrComponents: [
        [
          { type: "text", text: "What are the top " },
          { type: "context", context: PUBLISHER_INLINE_CONTEXT },
          { type: "text", text: " and " },
          { type: "context", context: DOMAIN_INLINE_CONTEXT },
          { type: "text", text: "?" },
        ],
      ],
    },
    {
      title: "Single line malformed prompt",
      prompt: `What are the top ${PUBLISHER_INLINE_PROMPT} and ${DOMAIN_INLINE_PROMPT.substring(0, DOMAIN_INLINE_PROMPT.length / 2)}?`,
      textOrComponents: [
        [
          { type: "text", text: "What are the top " },
          { type: "context", context: PUBLISHER_INLINE_CONTEXT },
          {
            type: "text",
            text: ` and ${DOMAIN_INLINE_PROMPT.substring(0, DOMAIN_INLINE_PROMPT.length / 2)}?`,
          },
        ],
      ],
    },
    {
      title: "Multi line valid prompt",
      prompt: `What are the top ${PUBLISHER_INLINE_PROMPT} and ${DOMAIN_INLINE_PROMPT}?\nUse ${IMPRESSIONS_INLINE_PROMPT} as a measure.`,
      textOrComponents: [
        [
          { type: "text", text: "What are the top " },
          { type: "context", context: PUBLISHER_INLINE_CONTEXT },
          { type: "text", text: " and " },
          { type: "context", context: DOMAIN_INLINE_CONTEXT },
          { type: "text", text: "?" },
        ],
        [
          { type: "text", text: "Use " },
          { type: "context", context: IMPRESSIONS_INLINE_CONTEXT },
          { type: "text", text: " as a measure." },
        ],
      ],
    },
  ];

  for (const { title, prompt, textOrComponents } of testCases) {
    it(title, () => {
      const convertedTextOrComponents =
        convertPromptWithInlineContextToComponents(prompt);
      expect(convertedTextOrComponents).toEqual(textOrComponents);
    });
  }
});

const IMPRESSIONS_INLINE_PROMPT = `<chat-reference>type="measure" metricsView="adbids" measure="impressions"</chat-reference>`;
const IMPRESSIONS_INLINE_CONTEXT = {
  type: InlineContextType.Measure,
  metricsView: "adbids",
  measure: "impressions",
} satisfies InlineContext;
const PUBLISHER_INLINE_PROMPT = `<chat-reference>type="dimension" metricsView="adbids" dimension="publisher"</chat-reference>`;
const PUBLISHER_INLINE_CONTEXT = {
  type: InlineContextType.Dimension,
  metricsView: "adbids",
  dimension: "publisher",
} satisfies InlineContext;
const DOMAIN_INLINE_PROMPT = `<chat-reference>type="dimension" metricsView="adbids" dimension="domain"</chat-reference>`;
const DOMAIN_INLINE_CONTEXT = {
  type: InlineContextType.Dimension,
  metricsView: "adbids",
  dimension: "domain",
} satisfies InlineContext;
