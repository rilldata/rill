import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import { buildEvalFileContents, draftEvalCaseFromExchange } from "./draft-eval-case";

describe("buildEvalFileContents", () => {
  it("round-trips a multiline answer with colon-terminated lines", () => {
    // LLM answers routinely end lines with colons; in flow style YAML this would
    // silently re-parse as a nested map instead of a string.
    const answer =
      "The top countries by revenue are:\n\n1. US with $5M\n2. DK with $1M";
    const contents = buildEvalFileContents({
      name: "top_countries",
      question: "Which countries have the highest revenue?",
      expectedAnswer: answer,
      metricsView: "sales",
      measures: ["net_revenue"],
    });

    const parsed = parseDocument(contents).toJS();
    expect(parsed.type).toBe("eval");
    expect(parsed.cases).toHaveLength(1);
    expect(parsed.cases[0].name).toBe("top_countries");
    expect(parsed.cases[0].expect.answer).toBe(answer);
    expect(parsed.cases[0].expect.metrics_view).toBe("sales");
  });
});

describe("draftEvalCaseFromExchange", () => {
  it("extracts the prompt, answer and query expectations from an exchange", () => {
    const messages = [
      {
        id: "call1",
        type: "call",
        tool: "router_agent",
        contentData: JSON.stringify({ prompt: "Revenue by country?" }),
      },
      {
        id: "q1",
        parentId: "call1",
        type: "call",
        tool: "query_metrics_view",
        contentData: JSON.stringify({
          metrics_view: "sales",
          measures: [{ name: "net_revenue" }],
          dimensions: [{ name: "country" }],
        }),
      },
      {
        id: "res1",
        parentId: "call1",
        type: "result",
        tool: "router_agent",
        contentData: JSON.stringify({ response: "US leads with $5M." }),
      },
    ];

    const draft = draftEvalCaseFromExchange(messages, "res1");
    expect(draft).not.toBeNull();
    expect(draft?.question).toBe("Revenue by country?");
    expect(draft?.expectedAnswer).toBe("US leads with $5M.");
    expect(draft?.metricsView).toBe("sales");
    expect(draft?.measures).toEqual(["net_revenue"]);
    expect(draft?.dimensions).toEqual(["country"]);
  });
});
