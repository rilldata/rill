import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { describe, expect, it } from "vitest";
import { describeFilters } from "./header";

const expr = { cond: { op: "OPERATION_AND", exprs: [] } };

function manager(
  dimensions: Record<string, unknown>[],
  measures: Record<string, unknown>[],
): ExpressionFilterManager {
  return {
    sortedFilterManagers: { dimensions, measures },
  } as unknown as ExpressionFilterManager;
}

describe("describeFilters", () => {
  it("spells out dimension filters by mode", () => {
    const lines = describeFilters(
      manager(
        [
          {
            expr,
            label: "Country",
            mode: DimensionFilterMode.Select,
            selectedValues: ["US", "CA"],
            exclude: false,
            inputText: "",
          },
          {
            expr,
            label: "Device",
            mode: DimensionFilterMode.Select,
            selectedValues: ["bot"],
            exclude: true,
            inputText: "",
          },
          {
            expr,
            label: "Publisher",
            mode: DimensionFilterMode.Contains,
            selectedValues: [],
            exclude: false,
            inputText: "news",
          },
          {
            expr: undefined,
            label: "Unset",
            mode: DimensionFilterMode.Select,
            selectedValues: [],
            exclude: false,
            inputText: "",
          },
        ],
        [],
      ),
    );
    expect(lines).toStrictEqual([
      "Country: US, CA",
      "Device: excludes bot",
      'Publisher: contains "news"',
    ]);
  });

  it("summarizes long value lists", () => {
    const values = Array.from({ length: 25 }, (_, i) => `v${i}`);
    const [line] = describeFilters(
      manager(
        [
          {
            expr,
            label: "Country",
            mode: DimensionFilterMode.InList,
            selectedValues: values,
            exclude: false,
            inputText: "",
          },
        ],
        [],
      ),
    );
    expect(line).toBe(`Country: ${values.slice(0, 20).join(", ")} and 5 more`);
  });

  it("spells out measure filters", () => {
    const lines = describeFilters(
      manager(
        [],
        [
          {
            expr,
            label: "Impressions",
            operation: MeasureFilterOperation.GreaterThan,
            type: MeasureFilterType.Value,
            value1: "1000",
            value2: "",
          },
          {
            expr,
            label: "Spend",
            operation: MeasureFilterOperation.Between,
            type: MeasureFilterType.Value,
            value1: "10",
            value2: "20",
          },
          {
            expr,
            label: "CTR",
            operation: MeasureFilterOperation.LessThan,
            type: MeasureFilterType.PercentChange,
            value1: "5",
            value2: "",
          },
        ],
      ),
    );
    expect(lines).toStrictEqual([
      "Impressions > 1000",
      "Spend between 10 and 20",
      "CTR % change < 5%",
    ]);
  });
});
