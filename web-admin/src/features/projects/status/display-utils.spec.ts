import { describe, expect, it } from "vitest";
import { formatConnectorName, getResourceKindTagColor } from "./display-utils";

describe("display-utils", () => {
  describe("formatConnectorName", () => {
    const testCases: [string | undefined, string][] = [
      [undefined, "—"],
      ["duckdb", "DuckDB"],
      ["clickhouse", "ClickHouse"],
      ["druid", "Druid"],
      ["pinot", "Pinot"],
      ["openai", "OpenAI"],
      ["claude", "Claude"],
    ];

    for (const [input, expected] of testCases) {
      it(`formatConnectorName(${JSON.stringify(input)}) = ${expected}`, () => {
        expect(formatConnectorName(input)).toEqual(expected);
      });
    }
  });

  describe("getResourceKindTagColor", () => {
    const testCases: [string, string][] = [
      ["rill.runtime.v1.MetricsView", "blue"],
      ["rill.runtime.v1.Model", "green"],
      ["rill.runtime.v1.Report", "orange"],
      ["rill.runtime.v1.Source", "purple"],
      ["rill.runtime.v1.Theme", "magenta"],
      ["rill.runtime.v1.Unknown", "gray"],
      ["some.other.kind", "gray"],
    ];

    for (const [kind, expectedColor] of testCases) {
      it(`getResourceKindTagColor(${kind}) = ${expectedColor}`, () => {
        expect(getResourceKindTagColor(kind)).toEqual(expectedColor);
      });
    }
  });
});
