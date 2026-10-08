import { describe, expect, it } from "vitest";
import {
  applyFieldConfigToMetricsView,
  columnSettingsCapabilities,
  fieldConfig,
  fieldConfigs,
  fieldName,
  fieldNames,
  mergeFieldList,
  normalizeRowDimensionEntries,
  resolveColumnStyles,
  setFieldConfig,
  sortableFieldNames,
  sortingFromSpec,
  sortingToSpec,
  stripColumnDimensionKeys,
  wrapLinesFromSpec,
  type PivotFieldEntry,
} from "./field-config";
import type { PivotSpec, TableSpec } from "./index";

const isMeasure = (name: string) => ["impressions", "spend"].includes(name);

const tableSpec: TableSpec = {
  metrics_view: "ads",
  columns: [
    { name: "campaign", width: 260, wrap: true, label: "Campaign name" },
    "flight_start",
    { name: "impressions", width: 90, format_d3: ".3s", align: "center" },
  ],
};

const pivotSpec: PivotSpec = {
  metrics_view: "ads",
  measures: [{ name: "impressions", width: 120 }, "spend"],
  row_dimensions: [{ name: "campaign", width: 260, wrap: true }, "device"],
  col_dimensions: [{ name: "country", label: "Country name" }],
};

describe("fieldName / fieldNames", () => {
  it("reads names from strings and objects and skips malformed entries", () => {
    expect(fieldName("a")).toBe("a");
    expect(fieldName({ name: "b", width: 10 })).toBe("b");
    expect(fieldName(null)).toBe("");
    expect(fieldName({ width: 10 } as never)).toBe("");
    expect(
      fieldNames(["a", { name: "b" }, 5 as never, { name: 3 } as never]),
    ).toEqual(["a", "b"]);
    expect(fieldNames(undefined)).toEqual([]);
    expect(fieldNames("a" as never)).toEqual([]);
  });
});

describe("fieldConfig / fieldConfigs", () => {
  it("returns sanitized objects keyed by name", () => {
    expect(fieldConfig(tableSpec.columns, "campaign")).toEqual({
      name: "campaign",
      width: 260,
      wrap: true,
      label: "Campaign name",
    });
    expect(fieldConfig(tableSpec.columns, "flight_start")).toEqual({
      name: "flight_start",
    });
    expect(fieldConfig(tableSpec.columns, "missing")).toBeUndefined();
    expect(Object.keys(fieldConfigs(pivotSpec))).toEqual([
      "impressions",
      "spend",
      "campaign",
      "device",
      "country",
    ]);
  });

  // The canvas YAML is hand-editable, so malformed values must be dropped
  // instead of crashing the canvas.
  it("drops malformed override values", () => {
    const entries = [
      {
        name: "campaign",
        width: "wide",
        wrap: "yes",
        align: "middle",
        label: "",
        format_preset: 4,
      } as never,
    ];
    expect(fieldConfig(entries, "campaign")).toEqual({ name: "campaign" });
  });
});

describe("mergeFieldList", () => {
  it("keeps the objects of surviving names in the new order", () => {
    expect(
      mergeFieldList(tableSpec.columns, ["impressions", "campaign", "spend"]),
    ).toEqual([
      { name: "impressions", width: 90, format_d3: ".3s", align: "center" },
      { name: "campaign", width: 260, wrap: true, label: "Campaign name" },
      "spend",
    ]);
  });

  it("works without a previous list", () => {
    expect(mergeFieldList(undefined, ["a"])).toEqual(["a"]);
  });
});

describe("setFieldConfig", () => {
  it("upserts keys, removes nulled keys and collapses to a string", () => {
    let list: PivotFieldEntry[] = ["a", "b"];
    list = setFieldConfig(list, "b", { width: 140 });
    expect(list).toEqual(["a", { name: "b", width: 140 }]);
    list = setFieldConfig(list, "b", { label: "B", width: null });
    expect(list).toEqual(["a", { name: "b", label: "B" }]);
    list = setFieldConfig(list, "b", { label: null });
    expect(list).toEqual(["a", "b"]);
  });

  it("appends a missing name", () => {
    expect(setFieldConfig(["a"], "c", { align: "left" })).toEqual([
      "a",
      { name: "c", align: "left" },
    ]);
  });

  it("leaves other entries untouched", () => {
    const list = setFieldConfig(tableSpec.columns, "flight_start", {
      width: 140,
    });
    expect(list[0]).toBe(tableSpec.columns[0]);
    expect(list[1]).toEqual({ name: "flight_start", width: 140 });
  });
});

describe("stripColumnDimensionKeys", () => {
  it("keeps only the name and label", () => {
    expect(
      stripColumnDimensionKeys({
        name: "country",
        width: 100,
        wrap: true,
        align: "left",
        label: "Country",
      }),
    ).toEqual({ name: "country", label: "Country" });
    expect(stripColumnDimensionKeys({ name: "country", width: 100 })).toBe(
      "country",
    );
    expect(stripColumnDimensionKeys("country")).toBe("country");
  });
});

describe("columnSettingsCapabilities", () => {
  it("offers width to table columns, measures and the first row dimension", () => {
    expect(columnSettingsCapabilities("columns", 2, false)).toEqual({
      width: true,
      wrap: true,
      align: true,
      format: false,
    });
    expect(columnSettingsCapabilities("columns", 0, true)).toEqual({
      width: true,
      wrap: false,
      align: true,
      format: true,
    });
    expect(columnSettingsCapabilities("measures", 1, true).width).toBe(true);
    expect(columnSettingsCapabilities("row_dimensions", 0, false)).toEqual({
      width: true,
      wrap: true,
      align: false,
      format: false,
    });
    expect(columnSettingsCapabilities("row_dimensions", 1, false).width).toBe(
      false,
    );
    expect(columnSettingsCapabilities("col_dimensions", 0, false)).toEqual({
      width: false,
      wrap: false,
      align: false,
      format: false,
    });
  });
});

describe("wrapLinesFromSpec", () => {
  it("defaults to 2 and clamps into 1–5", () => {
    expect(wrapLinesFromSpec({ ...tableSpec })).toBe(2);
    expect(wrapLinesFromSpec({ ...tableSpec, wrap_lines: 4 })).toBe(4);
    expect(wrapLinesFromSpec({ ...tableSpec, wrap_lines: 0 })).toBe(1);
    expect(wrapLinesFromSpec({ ...tableSpec, wrap_lines: 12 })).toBe(5);
    expect(wrapLinesFromSpec({ ...tableSpec, wrap_lines: "3" as never })).toBe(
      2,
    );
  });
});

describe("resolveColumnStyles", () => {
  it("keys table styles by column and clamps widths to the role bounds", () => {
    expect(resolveColumnStyles(tableSpec, isMeasure)).toEqual({
      campaign: { width: 260, wrap: true },
      impressions: { width: 90, align: "center" },
    });
    expect(
      resolveColumnStyles(
        {
          ...tableSpec,
          columns: [
            { name: "campaign", width: 900 },
            { name: "impressions", width: 10, wrap: true },
          ],
        },
        isMeasure,
      ),
    ).toEqual({
      campaign: { width: 600 },
      // Measures never wrap, and their width is clamped to the measure bounds.
      impressions: { width: 60 },
    });
  });

  it("keys pivot styles by measure and by the first row dimension only", () => {
    expect(resolveColumnStyles(pivotSpec, isMeasure)).toEqual({
      impressions: { width: 120 },
      campaign: { width: 260, wrap: true },
    });
    const reordered: PivotSpec = {
      ...pivotSpec,
      row_dimensions: ["device", { name: "campaign", width: 260 }],
    };
    expect(resolveColumnStyles(reordered, isMeasure)).toEqual({
      impressions: { width: 120 },
    });
  });

  it("ignores align on the row header", () => {
    const spec: PivotSpec = {
      ...pivotSpec,
      row_dimensions: [{ name: "campaign", width: 200, align: "center" }],
    };
    expect(resolveColumnStyles(spec, isMeasure).campaign).toEqual({
      width: 200,
    });
  });
});

describe("normalizeRowDimensionEntries", () => {
  it("drops align everywhere and width/wrap after the first entry", () => {
    expect(
      normalizeRowDimensionEntries([
        { name: "campaign", width: 200, wrap: true, align: "center" },
        { name: "device", width: 150, wrap: true, label: "Device" },
        "country",
      ]),
    ).toEqual([
      { name: "campaign", width: 200, wrap: true },
      { name: "device", label: "Device" },
      "country",
    ]);
  });
});

describe("applyFieldConfigToMetricsView", () => {
  const measures = [
    {
      name: "impressions",
      displayName: "Impressions",
      formatPreset: "humanize",
      formatD3Locale: { decimal: "," },
    },
    { name: "spend", displayName: "Spend", formatD3: ",.2f" },
  ];
  const dimensions = [
    { name: "campaign", displayName: "Campaign" },
    { name: "country", column: "country", displayName: "Country" },
  ];

  it("overrides labels and replaces the format fields as a set", () => {
    const result = applyFieldConfigToMetricsView(measures, dimensions, {
      ...pivotSpec,
      measures: [
        { name: "impressions", format_d3: ".3s" },
        { name: "spend", format_preset: "currency_usd", label: "Cost" },
      ],
    });
    expect(result.measures).toEqual([
      {
        name: "impressions",
        displayName: "Impressions",
        formatD3: ".3s",
        formatD3Locale: { decimal: "," },
      },
      { name: "spend", displayName: "Cost", formatPreset: "currency_usd" },
    ]);
    expect(result.dimensions).toEqual([
      { name: "campaign", displayName: "Campaign" },
      { name: "country", column: "country", displayName: "Country name" },
    ]);
  });

  it("returns the inputs untouched without overrides", () => {
    const spec: TableSpec = { metrics_view: "ads", columns: ["campaign"] };
    const result = applyFieldConfigToMetricsView(measures, dimensions, spec);
    expect(result.measures).toBe(measures);
    expect(result.dimensions).toBe(dimensions);
  });
});

describe("sorting", () => {
  it("lists the sortable fields", () => {
    expect(sortableFieldNames(tableSpec)).toEqual([
      "campaign",
      "flight_start",
      "impressions",
    ]);
    expect(sortableFieldNames(pivotSpec)).toEqual([
      "impressions",
      "spend",
      "campaign",
      "device",
    ]);
  });

  it("builds the initial sorting with role-based default directions", () => {
    expect(
      sortingFromSpec({ ...tableSpec, sort_by: "impressions" }, isMeasure),
    ).toEqual([{ id: "impressions", desc: true }]);
    expect(
      sortingFromSpec({ ...tableSpec, sort_by: "campaign" }, isMeasure),
    ).toEqual([{ id: "campaign", desc: false }]);
    expect(
      sortingFromSpec(
        { ...tableSpec, sort_by: "campaign", sort_dir: "desc" },
        isMeasure,
      ),
    ).toEqual([{ id: "campaign", desc: true }]);
    expect(sortingFromSpec(tableSpec, isMeasure)).toEqual([]);
    expect(
      sortingFromSpec({ ...tableSpec, sort_by: "missing" }, isMeasure),
    ).toEqual([]);
  });

  it("maps any pivot row dimension to the row header column", () => {
    expect(
      sortingFromSpec({ ...pivotSpec, sort_by: "device" }, isMeasure),
    ).toEqual([{ id: "campaign", desc: false }]);
    expect(
      sortingFromSpec({ ...pivotSpec, sort_by: "spend" }, isMeasure),
    ).toEqual([{ id: "spend", desc: true }]);
  });

  it("targets the comparison column only while the widget compares", () => {
    const spec: TableSpec = {
      ...tableSpec,
      sort_by: "impressions",
      sort_comparison: "delta",
    };
    expect(sortingFromSpec(spec, isMeasure, true)).toEqual([
      { id: "impressions__delta_abs", desc: true },
    ]);
    expect(sortingFromSpec(spec, isMeasure, false)).toEqual([
      { id: "impressions", desc: true },
    ]);
    expect(
      sortingFromSpec(
        { ...spec, sort_comparison: "percent_change", sort_dir: "asc" },
        isMeasure,
        true,
      ),
    ).toEqual([{ id: "impressions__delta_rel", desc: false }]);
    // Row dimensions have no comparison columns.
    expect(
      sortingFromSpec(
        { ...pivotSpec, sort_by: "campaign", sort_comparison: "delta" },
        isMeasure,
        true,
      ),
    ).toEqual([{ id: "campaign", desc: false }]);
  });

  it("converts a clicked header back to the spec keys", () => {
    expect(sortingToSpec([{ id: "spend", desc: false }], pivotSpec)).toEqual({
      sort_by: "spend",
      sort_dir: "asc",
      sort_comparison: undefined,
    });
    expect(
      sortingToSpec([{ id: "spend__delta_rel", desc: true }], pivotSpec),
    ).toEqual({
      sort_by: "spend",
      sort_dir: "desc",
      sort_comparison: "percent_change",
    });
    expect(sortingToSpec([], pivotSpec)).toEqual({
      sort_by: undefined,
      sort_dir: undefined,
      sort_comparison: undefined,
    });
    // A measure under one column-dimension value has no YAML form.
    expect(sortingToSpec([{ id: "c0v1m0", desc: true }], pivotSpec)).toBe(
      undefined,
    );
  });
});
