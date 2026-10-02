import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, expect, it } from "vitest";
import { decodePivotSort, encodePivotSort } from "./pivot-sort";
import { getSortForAccessor } from "./pivot-utils";
import type { PivotDataStoreConfig } from "./types";

function makeConfig(
  overrides: Partial<PivotDataStoreConfig> = {},
): PivotDataStoreConfig {
  return {
    measureNames: ["revenue"],
    rowDimensionNames: ["country"],
    colDimensionNames: ["region", "segment"],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      sorting: [],
    },
    time: {
      timeStart: undefined,
      timeEnd: undefined,
      timeZone: "UTC",
      timeDimension: "timestamp",
    },
    isFlat: false,
    ...overrides,
  } as unknown as PivotDataStoreConfig;
}

describe("pivot sort targets", () => {
  it("decodes legacy accessors and follows reordered column axes", () => {
    const config = makeConfig();
    const target = decodePivotSort(
      [{ id: "c0v1_c1v0m0", desc: true }],
      config,
      { region: ["NA", "EU"], segment: ["Consumer", "Enterprise"] },
    );

    expect(target).toEqual({
      measure: "revenue",
      column_values: [
        { dimension: "region", value: "EU" },
        { dimension: "segment", value: "Consumer" },
      ],
      desc: true,
    });
    expect(
      encodePivotSort(target, config, {
        region: ["EU", "NA"],
        segment: ["Enterprise", "Consumer"],
      }),
    ).toEqual([{ id: "c0v0_c1v1m0", desc: true }]);
  });

  it("builds nested query sorting from a semantic default", () => {
    const config = makeConfig({
      colDimensionNames: ["region"],
      defaultSort: {
        measure: "revenue",
        column_values: [{ dimension: "region", value: "EU" }],
        desc: true,
      },
    });

    const result = getSortForAccessor("country", config, {
      region: ["NA", "EU"],
    });

    expect(result.sortPivotBy).toEqual([{ name: "revenue", desc: true }]);
    expect(result.where?.cond?.exprs?.[0]?.cond?.exprs?.[0]?.ident).toBe(
      "region",
    );
    expect(result.where?.cond?.exprs?.[0]?.cond?.exprs?.[1]?.val).toBe("EU");
  });

  it("ignores a stale legacy accessor without falling back to the default", () => {
    const config = makeConfig({
      colDimensionNames: ["region"],
      pivot: { sorting: [{ id: "c0v99m0", desc: true }] },
      defaultSort: { field: "revenue", desc: true },
    } as Partial<PivotDataStoreConfig>);

    expect(
      getSortForAccessor("country", config, { region: ["NA", "EU"] }),
    ).toMatchObject({ sortPivotBy: [] });
  });
});
