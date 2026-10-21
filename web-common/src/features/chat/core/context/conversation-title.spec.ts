import { formatConversationTitle } from "@rilldata/web-common/features/chat/core/context/conversation-title.ts";
import type { InlineContextMetadata } from "@rilldata/web-common/features/chat/core/context/metadata.ts";
import { describe, expect, it } from "vitest";

const meta: InlineContextMetadata = {
  metricsViewSpecs: {
    orders: {
      metricsViewSpec: { displayName: "Sales" },
      measures: { revenue: { name: "revenue", displayName: "Net Revenue" } },
      dimensions: {
        country: { name: "country", displayName: "Customer Country" },
      },
    },
  },
  canvasSpecs: { overview: { displayName: "Sales Overview" } },
  componentSpecs: {},
};

const emptyMeta: InlineContextMetadata = {
  metricsViewSpecs: {},
  canvasSpecs: {},
  componentSpecs: {},
};

describe("formatConversationTitle", () => {
  const testCases: {
    title: string;
    input: string;
    meta?: InlineContextMetadata;
    expected: string;
  }[] = [
    {
      title: "plain text",
      input: "How did revenue change?",
      expected: "How did revenue change?",
    },
    {
      title: "metrics view",
      input: `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> by month`,
      expected: "Sales by month",
    },
    {
      title: "measure",
      input: `What drives <chat-reference>type="measure" metricsView="orders" measure="revenue"</chat-reference>?`,
      expected: "What drives Net Revenue?",
    },
    {
      title: "dimension",
      input: `Revenue by <chat-reference>type="dimension" metricsView="orders" dimension="country"</chat-reference>`,
      expected: "Revenue by Customer Country",
    },
    {
      title: "canvas",
      input: `Summarize <chat-reference>type="canvas" canvas="overview"</chat-reference>`,
      expected: "Summarize Sales Overview",
    },
    {
      title: "skill",
      input: `<chat-reference>type="skill" skill="monthly-close"</chat-reference> for March`,
      expected: "/monthly-close for March",
    },
    {
      title: "column",
      input: `Nulls in <chat-reference>type="column" model="orders" column="revenue"</chat-reference>`,
      expected: "Nulls in revenue",
    },
    {
      title: "falls back to names without metadata",
      input: `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> by month`,
      meta: emptyMeta,
      expected: "orders by month",
    },
    {
      title: "unknown type is dropped",
      input: `<chat-reference>type="unknown" unknown="x"</chat-reference> hello`,
      expected: "hello",
    },
    {
      title: "reference truncated by the backend",
      input: `<chat-reference>type="metricsView" metricsView="orders"</chat-reference>...`,
      expected: "Sales...",
    },
    {
      title: "partial reference in an older title is dropped",
      input: `Compare <chat-reference>type="metricsView" metricsVi...`,
      expected: "Compare...",
    },
    {
      title: "title that is only a partial reference",
      input: `<chat-reference>type="metricsView" metricsVi...`,
      expected: "",
    },
  ];

  for (const { title, input, meta: m, expected } of testCases) {
    it(title, () => {
      expect(formatConversationTitle(input, m ?? meta)).toBe(expected);
    });
  }
});
