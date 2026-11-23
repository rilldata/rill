import {
  formatConversationTitle,
  stripPartialChatReference,
} from "@rilldata/web-common/features/chat/core/context/conversation-title.ts";
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
      title: "inherited object property as type is dropped",
      input: `<chat-reference>type="toString"</chat-reference> <chat-reference>type="constructor"</chat-reference>hello`,
      expected: "hello",
    },
    {
      title: "dimension values without values",
      input: `<chat-reference>type="dimensionValues" metricsView="orders" dimension="country"</chat-reference>`,
      expected: "Customer Country:",
    },
    {
      title: "dimension values with values attribute",
      input: `<chat-reference>type="dimensionValues" metricsView="orders" dimension="country" values="US,IN"</chat-reference>`,
      expected: "Customer Country:",
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
      title: "older title cut within the opening tag name",
      // The old backend truncated to `title[:47] + "..."`, so a reference starting at byte 41 was cut to `<chat-`.
      input: `${"a".repeat(40)} <chat-...`,
      expected: `${"a".repeat(40)}...`,
    },
    {
      title: "older title cut right after the opening angle bracket",
      input: `${"a".repeat(46)}<...`,
      expected: `${"a".repeat(46)}...`,
    },
    {
      title: "text resembling a tag name prefix without truncation is kept",
      input: "Is x <ch y?",
      expected: "Is x <ch y?",
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

describe("stripPartialChatReference", () => {
  it("drops a trailing partial reference", () => {
    expect(
      stripPartialChatReference(
        `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> vs <chat-reference>type="metricsVi...`,
      ),
    ).toBe(
      `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> vs`,
    );
  });

  it("drops a title cut within the opening tag name", () => {
    expect(stripPartialChatReference(`Compare <chat-...`)).toBe("Compare");
  });

  it("keeps complete references", () => {
    const title = `<chat-reference>type="metricsView" metricsView="orders"</chat-reference> by month`;
    expect(stripPartialChatReference(title)).toBe(title);
  });
});
