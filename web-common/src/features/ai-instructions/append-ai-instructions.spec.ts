import { describe, expect, it } from "vitest";
import { parse } from "yaml";
import {
  appendAIInstructions,
  appendYAMLMeasure,
} from "./append-ai-instructions";

describe("appendAIInstructions", () => {
  it("appends to existing instructions and preserves comments", () => {
    const content = `# Project config
display_name: My project
ai_instructions: |
  Revenue means net_revenue.
olap_connector: duckdb
`;
    const out = appendAIInstructions(
      content,
      "abc refers to the net_revenue measure.",
    );
    expect(out).toContain("# Project config");
    const parsed = parse(out);
    expect(parsed.ai_instructions).toBe(
      "Revenue means net_revenue.\n\nabc refers to the net_revenue measure.\n",
    );
    expect(parsed.display_name).toBe("My project");
    expect(parsed.olap_connector).toBe("duckdb");
  });

  it("creates the property when absent, in block style", () => {
    const out = appendAIInstructions(
      "display_name: My project\n",
      "Always use fiscal quarters.",
    );
    expect(out).toContain("ai_instructions: |");
    expect(parse(out).ai_instructions).toBe("Always use fiscal quarters.\n");
  });

  it("works on an empty file", () => {
    const out = appendAIInstructions("", "Always use fiscal quarters.");
    expect(parse(out).ai_instructions).toBe("Always use fiscal quarters.\n");
  });

  it("rejects multi-document files", () => {
    expect(() => appendAIInstructions("a: 1\n---\nb: 2\n", "rule")).toThrow(
      "Multi-document",
    );
  });

  it("rejects non-mapping files", () => {
    expect(() => appendAIInstructions("- a\n- b\n", "rule")).toThrow(
      "not a YAML mapping",
    );
  });
});

describe("appendYAMLMeasure", () => {
  const measureYAML =
    "name: avg_order_value\nexpression: SUM(revenue) / COUNT(*)\ndescription: Average order value.";

  it("appends to an existing measures list and preserves comments", () => {
    const content = `type: metrics_view
model: sales
measures:
  # The main revenue measure
  - name: net_revenue
    expression: SUM(revenue)
`;
    const out = appendYAMLMeasure(content, measureYAML);
    expect(out).toContain("# The main revenue measure");
    const parsed = parse(out);
    expect(parsed.measures).toHaveLength(2);
    expect(parsed.measures[1]).toEqual({
      name: "avg_order_value",
      expression: "SUM(revenue) / COUNT(*)",
      description: "Average order value.",
    });
  });

  it("creates the measures list when absent", () => {
    const out = appendYAMLMeasure("type: metrics_view\nmodel: sales\n", measureYAML);
    const parsed = parse(out);
    expect(parsed.measures).toHaveLength(1);
    expect(parsed.measures[0].name).toBe("avg_order_value");
  });

  it("rejects a non-mapping measure definition", () => {
    expect(() => appendYAMLMeasure("type: metrics_view\n", "- a\n- b\n")).toThrow(
      "must be a YAML mapping",
    );
  });

  it("rejects a file whose measures property is not a list", () => {
    expect(() => appendYAMLMeasure("measures: nope\n", measureYAML)).toThrow(
      "not a list",
    );
  });
});
