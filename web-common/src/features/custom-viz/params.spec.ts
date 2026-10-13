import type {
  V1ComponentParam,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  buildDefaultArgs,
  getDeclaredParams,
  orderByOptions,
  paramChannels,
  paramToInputParam,
  prettyParamLabel,
  reconcileOrderByArg,
  optimisticRendererProps,
  substituteArgs,
  substituteArgsRecursively,
} from "./params";

const params: V1ComponentParam[] = [
  { name: "metrics_view", type: "metrics_view", required: true },
  {
    name: "measure",
    type: "measure",
    required: true,
    metricsViewParam: "metrics_view",
  },
  {
    name: "dim",
    type: "dimension",
    metricsViewParam: "metrics_view",
  },
  {
    name: "time_dim",
    type: "time_dimension",
    required: true,
    metricsViewParam: "metrics_view",
  },
  { name: "limit", type: "number", default: 500 },
  {
    name: "shape",
    type: "string",
    default: "line",
    options: ["line", "area"],
  },
  { name: "smooth", type: "boolean" },
];

const metricsViewSpec: V1MetricsViewSpec = {
  timeDimension: "ts",
  dimensions: [
    { name: "country", type: "DIMENSION_TYPE_CATEGORICAL" },
    { name: "updated_at", type: "DIMENSION_TYPE_TIME" },
    { name: "device", type: "DIMENSION_TYPE_CATEGORICAL" },
  ],
  measures: [{ name: "total" }, { name: "avg_price" }],
};

describe("paramToInputParam", () => {
  it("maps param types to inspector input types", () => {
    expect(paramToInputParam(params[0]).type).toBe("metrics");
    expect(paramToInputParam(params[1]).type).toBe("measure");
    expect(paramToInputParam(params[2]).type).toBe("dimension");
    expect(paramToInputParam(params[4]).type).toBe("number");
    expect(paramToInputParam(params[6]).type).toBe("boolean");
  });

  it("maps time_dimension params to a time-constrained dimension picker", () => {
    const input = paramToInputParam(params[3]);
    expect(input.type).toBe("dimension");
    expect(input.meta?.includeTime).toBe(true);
    expect(input.meta?.timeFieldsOnly).toBe(true);
  });

  it("maps scalar params with options to a select", () => {
    const input = paramToInputParam(params[5]);
    expect(input.type).toBe("select");
    expect(input.meta?.options).toEqual([
      { value: "line", label: "line" },
      { value: "area", label: "area" },
    ]);
    expect(input.meta?.default).toBe("line");
  });

  it("marks numeric select options for coercion by the inspector", () => {
    const input = paramToInputParam({
      name: "limit",
      type: "number",
      options: [10, 25, 50],
    });

    expect(input.type).toBe("select");
    expect(input.meta?.options).toEqual([
      { value: "10", label: "10" },
      { value: "25", label: "25" },
      { value: "50", label: "50" },
    ]);
    expect(input.meta?.paramType).toBe("number");
  });

  it("marks required params as non-optional", () => {
    expect(paramToInputParam(params[0]).optional).toBe(false);
    expect(paramToInputParam(params[4]).optional).toBe(true);
  });
});

describe("buildDefaultArgs", () => {
  it("computes smart defaults for each param type", () => {
    const args = buildDefaultArgs(params, "mv1", metricsViewSpec);
    expect(args).toEqual({
      metrics_view: "mv1",
      measure: "total",
      dim: "country",
      time_dim: "ts",
      limit: 500,
      shape: "line",
      // smooth has no default and stays unset
    });
  });

  it("does not reuse the same field for two params of the same type", () => {
    const twoMeasures: V1ComponentParam[] = [
      { name: "m1", type: "measure", metricsViewParam: "metrics_view" },
      { name: "m2", type: "measure", metricsViewParam: "metrics_view" },
    ];
    const args = buildDefaultArgs(twoMeasures, "mv1", metricsViewSpec);
    expect(args).toEqual({ m1: "total", m2: "avg_price" });
  });

  it("skips time-typed dimensions for plain dimension params", () => {
    const spec: V1MetricsViewSpec = {
      dimensions: [
        { name: "updated_at", type: "DIMENSION_TYPE_TIME" },
        { name: "country", type: "DIMENSION_TYPE_CATEGORICAL" },
      ],
      measures: [],
    };
    const args = buildDefaultArgs(
      [{ name: "dim", type: "dimension" }],
      "mv1",
      spec,
    );
    expect(args).toEqual({ dim: "country" });
  });

  it("leaves field params unset when the metrics view has no matching field", () => {
    const args = buildDefaultArgs(
      [
        { name: "metrics_view", type: "metrics_view" },
        { name: "time_dim", type: "time_dimension" },
      ],
      "mv1",
      { dimensions: [], measures: [] },
    );
    expect(args).toEqual({ metrics_view: "mv1" });
  });

  it("defaults the order_by param to the bound measure", () => {
    const args = buildDefaultArgs(
      [
        { name: "metrics_view", type: "metrics_view", required: true },
        { name: "x_axis", type: "dimension", required: true },
        { name: "y_axis", type: "measure", required: true },
        { name: "order_by", type: "string", required: true },
      ],
      "mv1",
      metricsViewSpec,
    );
    expect(args.order_by).toBe(args.y_axis);
    expect(args.order_by).toBe("total");
  });

  it("falls back to a bound dimension for order_by when there is no measure", () => {
    const args = buildDefaultArgs(
      [
        { name: "x_axis", type: "dimension", required: true },
        { name: "order_by", type: "string", required: true },
      ],
      "mv1",
      metricsViewSpec,
    );
    expect(args.order_by).toBe("country");
  });
});

describe("buildDefaultArgs with value counts", () => {
  const spec: V1MetricsViewSpec = {
    timeDimension: "ts",
    dimensions: [
      { name: "site", type: "DIMENSION_TYPE_CATEGORICAL" },
      { name: "domain", type: "DIMENSION_TYPE_CATEGORICAL" },
      { name: "os", type: "DIMENSION_TYPE_CATEGORICAL" },
      { name: "device_type", type: "DIMENSION_TYPE_CATEGORICAL" },
      { name: "placement", type: "DIMENSION_TYPE_CATEGORICAL" },
    ],
    measures: [
      { name: "requests", formatPreset: "humanize" },
      { name: "bid_floor", formatPreset: "currency_usd" },
      { name: "max_bid", formatPreset: "currency_usd" },
      { name: "impressions", formatPreset: "humanize" },
    ],
  };
  // Capped at the probe limit (51), as the probe reports them.
  const valueCounts = new Map([
    ["site", 51],
    ["domain", 51],
    ["os", 28],
    ["device_type", 5],
    ["placement", 2],
  ]);
  const field = (name: string, type: string): V1ComponentParam => ({
    name,
    type,
    required: true,
  });

  it("gives a series channel a field with a handful of values, picked before the axis", () => {
    const args = buildDefaultArgs(
      [field("x_axis", "dimension"), field("color", "dimension")],
      "mv",
      spec,
      { channels: { x_axis: "x", color: "color" }, valueCounts },
    );
    // The series picks first. With series on top, the axis wants at most a dozen values; none is
    // left in range, so it takes the nearest.
    expect(args).toEqual({ x_axis: "placement", color: "device_type" });
  });

  it("gives the two-sided series of a pyramid a field with exactly two values", () => {
    const args = buildDefaultArgs([field("color", "dimension")], "mv", spec, {
      channels: { color: "color" },
      chartType: "Pyramid Chart",
      valueCounts,
    });
    expect(args).toEqual({ color: "placement" });
  });

  it("gives the x axis of a slope chart the two periods it connects", () => {
    const args = buildDefaultArgs(
      [field("x_axis", "dimension"), field("color", "dimension")],
      "mv",
      spec,
      {
        channels: { x_axis: "x", color: "color" },
        chartType: "Slope Chart",
        valueCounts,
      },
    );
    expect(args).toEqual({ x_axis: "placement", color: "device_type" });
  });

  it("gives the detail dimension the field with the most values", () => {
    const args = buildDefaultArgs(
      [field("x_axis", "dimension"), field("detail", "dimension")],
      "mv",
      spec,
      { channels: { x_axis: "x" }, valueCounts },
    );
    expect(args).toEqual({ x_axis: "os", detail: "site" });
  });

  it("falls back to the nearest varying field when nothing is in range", () => {
    const args = buildDefaultArgs(
      [field("color", "dimension")],
      "mv",
      {
        ...spec,
        dimensions: spec.dimensions!.slice(0, 3),
      },
      { channels: { color: "color" }, valueCounts },
    );
    expect(args).toEqual({ color: "os" });
  });

  it("pairs a range end with a measure in the primary measure's format", () => {
    const args = buildDefaultArgs(
      [field("y_axis", "measure"), field("y_axis_end", "measure")],
      "mv",
      spec,
      { channels: { y_axis: "y", y_axis_end: "y2" } },
    );
    expect(args).toEqual({ y_axis: "requests", y_axis_end: "impressions" });
  });

  it("reuses a measure rather than leaving a required param unbound", () => {
    const ohlc = ["open", "high", "low", "close", "volume"].map((name) =>
      field(name, "measure"),
    );
    const args = buildDefaultArgs(ohlc, "mv", spec, {
      channels: { open: "open", high: "high", low: "low", close: "close" },
    });
    expect(Object.keys(args).sort()).toEqual(
      ["close", "high", "low", "open", "volume"].sort(),
    );
    expect(Object.values(args).every(Boolean)).toBe(true);
  });
});

describe("buildDefaultArgs with field defaults", () => {
  it("ignores a field default the metrics view does not have", () => {
    const args = buildDefaultArgs(
      [
        { name: "m", type: "measure", default: "total_bids" },
        { name: "d", type: "dimension", default: "device" },
      ],
      "mv1",
      metricsViewSpec,
    );
    expect(args).toEqual({ m: "total", d: "device" });
  });
});

describe("paramChannels", () => {
  it("maps each param to the channel whose encoding references it", () => {
    expect(
      paramChannels({
        spec: {
          encodings: {
            x: { field: "{{ .params.x_axis }}" },
            color: "{{ .params.color }}",
            y: [{ field: "{{ .params.low }}" }],
          },
        },
      }),
    ).toEqual({ x_axis: "x", color: "color", low: "y" });
    expect(paramChannels(undefined)).toEqual({});
  });
});

describe("orderByOptions", () => {
  it("offers the field values currently bound to the field params, deduped", () => {
    const declared: V1ComponentParam[] = [
      { name: "metrics_view", type: "metrics_view" },
      { name: "x_axis", type: "dimension" },
      { name: "y_axis", type: "measure" },
      { name: "color", type: "measure" },
      { name: "order_by", type: "string" },
      { name: "limit", type: "number" },
    ];
    const options = orderByOptions(declared, {
      metrics_view: "mv1",
      x_axis: "country",
      y_axis: "total",
      color: "total", // duplicate value: appears once
      order_by: "total",
      limit: 10,
    });
    expect(options.map((option) => option.value)).toEqual(["country", "total"]);
    expect(options[0].label).toContain("X Axis");
  });

  it("skips unbound field params", () => {
    const options = orderByOptions([{ name: "x_axis", type: "dimension" }], {});
    expect(options).toEqual([]);
  });
});

describe("reconcileOrderByArg", () => {
  const declared: V1ComponentParam[] = [
    { name: "metrics_view", type: "metrics_view", required: true },
    { name: "x_axis", type: "dimension", required: true },
    { name: "y_axis", type: "measure", required: true },
    { name: "order_by", type: "string", required: true },
  ];

  it("follows a re-bound field that order_by pointed at", () => {
    const args = {
      x_axis: "country",
      y_axis: "avg_price", // re-bound from "total"
      order_by: "total",
    };
    reconcileOrderByArg(declared, args, {
      name: "y_axis",
      previousValue: "total",
    });
    expect(args.order_by).toBe("avg_price");
  });

  it("follows a re-bound dimension too", () => {
    const args = { x_axis: "device", y_axis: "total", order_by: "country" };
    reconcileOrderByArg(declared, args, {
      name: "x_axis",
      previousValue: "country",
    });
    expect(args.order_by).toBe("device");
  });

  it("keeps order_by when it points at a different, still-bound field", () => {
    const args = { x_axis: "device", y_axis: "total", order_by: "total" };
    reconcileOrderByArg(declared, args, {
      name: "x_axis",
      previousValue: "country",
    });
    expect(args.order_by).toBe("total");
  });

  it("re-seeds from the measure when order_by no longer matches any bound field", () => {
    // e.g. the metrics view changed and dependents were cleared around it.
    const args: Record<string, unknown> = {
      x_axis: "device",
      y_axis: "total",
      order_by: "stale_field",
    };
    reconcileOrderByArg(declared, args);
    expect(args.order_by).toBe("total");
  });

  it("does not override an explicit order_by change", () => {
    const args = { x_axis: "country", y_axis: "total", order_by: "country" };
    reconcileOrderByArg(declared, args, {
      name: "order_by",
      previousValue: "total",
    });
    expect(args.order_by).toBe("country");
  });

  it("unsets order_by when no field is bound at all", () => {
    const args: Record<string, unknown> = { order_by: "total" };
    reconcileOrderByArg(declared, args);
    expect(args.order_by).toBeUndefined();
  });
});

describe("substituteArgs", () => {
  it("substitutes .params and .args placeholders", () => {
    expect(
      substituteArgs(
        "SELECT {{ .params.measure }} FROM {{ .args.metrics_view }} LIMIT {{.params.limit}}",
        { measure: "total", metrics_view: "mv1", limit: 100 },
      ),
    ).toBe("SELECT total FROM mv1 LIMIT 100");
  });

  it("leaves unknown placeholders untouched", () => {
    expect(substituteArgs("{{ .params.missing }}", {})).toBe(
      "{{ .params.missing }}",
    );
  });

  it("does not substitute other template namespaces", () => {
    expect(substituteArgs("{{ .env.name }}", { name: "x" })).toBe(
      "{{ .env.name }}",
    );
  });

  it("substitutes repeated placeholders", () => {
    expect(
      substituteArgs("{{ .params.m }} + {{ .params.m }}", { m: "x" }),
    ).toBe("x + x");
  });
});

describe("substituteArgsRecursively", () => {
  it("substitutes strings nested in maps and arrays", () => {
    expect(
      substituteArgsRecursively(
        {
          metrics_sql: ["SELECT {{ .params.m }} FROM mv"],
          spec: {
            chartType: "Bar Chart",
            encodings: { y: { field: "{{ .params.m }}" } },
          },
          count: 1,
        },
        { m: "total" },
      ),
    ).toEqual({
      metrics_sql: ["SELECT total FROM mv"],
      spec: {
        chartType: "Bar Chart",
        encodings: { y: { field: "total" } },
      },
      count: 1,
    });
  });
});

describe("optimisticRendererProps", () => {
  it("resolves properties that only reference params", () => {
    expect(
      optimisticRendererProps(
        {
          metrics_sql: "SELECT {{ .params.m }} FROM mv",
          spec: { encodings: { y: { field: "{{ .params.m }}" } } },
        },
        { m: "total" },
      ),
    ).toEqual({
      metrics_sql: "SELECT total FROM mv",
      spec: { encodings: { y: { field: "total" } } },
    });
  });

  it("gives up when a placeholder is left over", () => {
    // .fields resolves on the server against the metrics view. Substituting only the params would
    // leave the display name as a literal placeholder, which Vega then reads as part of a tooltip
    // expression and fails to parse, so there is nothing safe to render optimistically.
    expect(
      optimisticRendererProps(
        {
          vega_spec:
            '{"encoding":{"y":{"field":"{{ .params.m }}","title":"{{ .fields.m.display_name }}"}}}',
        },
        { m: "total" },
      ),
    ).toBeUndefined();

    // Likewise for an unbound param, and for the namespaces only the server can resolve.
    expect(
      optimisticRendererProps({ spec: { x: "{{ .params.missing }}" } }, {}),
    ).toBeUndefined();
    expect(
      optimisticRendererProps({ spec: { x: "{{ .env.name }}" } }, {}),
    ).toBeUndefined();
    expect(
      optimisticRendererProps({ spec: { x: "{{ .user.email }}" } }, {}),
    ).toBeUndefined();
  });

  it("returns undefined when there are no renderer properties", () => {
    expect(optimisticRendererProps(undefined, {})).toBeUndefined();
  });
});

describe("getDeclaredParams", () => {
  it("prefers the valid spec and falls back to the draft spec only when allowed", () => {
    const resource = {
      component: {
        state: { validSpec: { params: [{ name: "a" }] } },
        spec: { params: [{ name: "b" }] },
      },
    };
    expect(getDeclaredParams(resource).map((p) => p.name)).toEqual(["a"]);

    const draftOnly = {
      component: { spec: { params: [{ name: "b" }] } },
    };
    expect(getDeclaredParams(draftOnly)).toEqual([]);
    expect(getDeclaredParams(draftOnly, true).map((p) => p.name)).toEqual([
      "b",
    ]);
  });
});

describe("prettyParamLabel", () => {
  it("converts snake_case to title case", () => {
    expect(prettyParamLabel("metrics_view")).toBe("Metrics View");
    expect(prettyParamLabel("time_dim")).toBe("Time Dim");
  });
});
