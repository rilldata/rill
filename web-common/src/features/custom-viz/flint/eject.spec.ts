import { describe, expect, it } from "vitest";
import {
  V1TimeGrain,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { compileFlintSpec, type FlintChartSpec } from "./compile";
import { deriveFlintFields } from "./semantic-types";
import { ejectToVegaSpec, type EjectFieldBinding } from "./eject";

/**
 * Ejection runs against real compiler output rather than a hand-written spec, since what needs
 * templating is whatever the compiler happened to bake in: encodings, transforms, expressions,
 * axis and legend titles, and formatter names.
 */

const metricsView: V1MetricsViewSpec = {
  timeDimension: "timestamp",
  measures: [
    {
      name: "bid_price_sum",
      displayName: "Sum of Bid Price",
      formatPreset: "currency_usd",
    },
    { name: "total_records", displayName: "Total records" },
  ],
  dimensions: [{ name: "publisher", displayName: "Publisher" }],
};

const rows = Array.from({ length: 6 }, (_, i) => ({
  timestamp: `2024-0${i + 1}-01T00:00:00Z`,
  publisher: ["Yahoo", "Google", "Microsoft"][i % 3],
  bid_price_sum: 1000 + i * 37,
  total_records: 40 + i,
}));

const bindings: EjectFieldBinding[] = [
  {
    param: "x_axis",
    field: "publisher",
    displayName: "Publisher",
    type: "dimension",
  },
  {
    param: "y_axis",
    field: "bid_price_sum",
    displayName: "Sum of Bid Price",
    type: "measure",
  },
  {
    param: "time_dim",
    field: "timestamp",
    displayName: "Time",
    type: "time_dimension",
  },
];

function eject(spec: FlintChartSpec, columns: string[]) {
  const compiled = compileFlintSpec({
    spec,
    rows,
    fields: deriveFlintFields(
      columns,
      metricsView,
      V1TimeGrain.TIME_GRAIN_MONTH,
    ),
    size: { width: 600, height: 400 },
    measureFields: ["bid_price_sum", "total_records"],
  });
  expect(compiled.error).toBeNull();
  const json = ejectToVegaSpec(
    compiled.spec as unknown as Record<string, unknown>,
    bindings,
  );
  return { json, spec: JSON.parse(json) as Record<string, any> };
}

function ejectBar() {
  return eject(
    {
      chartType: "Bar Chart",
      encodings: {
        x: { field: "publisher", sortBy: "y", sortOrder: "descending" },
        y: { field: "bid_price_sum" },
      },
    },
    ["publisher", "bid_price_sum"],
  );
}

describe("ejectToVegaSpec", () => {
  it("templates field names, display names and formatter names", () => {
    const { spec } = ejectBar();

    expect(spec.encoding.x.field).toBe("{{ .params.x_axis }}");
    expect(spec.encoding.x.title).toBe("{{ .fields.x_axis.display_name }}");
    expect(spec.encoding.y.field).toBe("{{ .params.y_axis }}");
    expect(spec.encoding.y.title).toBe("{{ .fields.y_axis.display_name }}");
    expect(spec.encoding.y.formatType).toBe("{{ .fields.y_axis.format_type }}");
    expect(spec.encoding.y.axis.formatType).toBe(
      "{{ .fields.y_axis.format_type }}",
    );
  });

  it("leaves no bound field name hard-coded", () => {
    const { json } = ejectBar();

    for (const { field } of bindings) {
      // Bound names may only appear inside the template references that reproduce them.
      const stripped = json.replace(/\{\{[^}]*\}\}/g, "");
      expect(stripped).not.toContain(field);
    }
    expect(json).not.toContain("Sum of Bid Price");
  });

  it("points the data at the renderer's named dataset", () => {
    const { spec } = ejectBar();

    expect(spec.data).toEqual({ name: "query1" });
    // The rows the compiler read must not be written into the component file.
    expect(JSON.stringify(spec)).not.toContain("Yahoo");
  });

  it("declares the Vega-Lite schema", () => {
    const { spec } = ejectBar();
    expect(spec.$schema).toBe(
      "https://vega.github.io/schema/vega-lite/v5.json",
    );
  });

  it("drops the fixed view size the compiler fills in", () => {
    const { spec } = ejectBar();

    expect(spec.config.view?.continuousWidth).toBeUndefined();
    expect(spec.config.view?.continuousHeight).toBeUndefined();
    // A step width is the compiler's density decision, not a container measurement.
    expect(spec.width).toEqual({ step: 40 });
  });

  it("templates field names embedded in transforms and expressions", () => {
    // A waterfall chart compiles to window transforms and calculate expressions that reference
    // the bound fields as `datum['<field>']`, not just as encoding fields.
    const { json, spec } = eject(
      {
        chartType: "Waterfall Chart",
        encodings: { x: { field: "publisher" }, y: { field: "bid_price_sum" } },
      },
      ["publisher", "bid_price_sum"],
    );

    const transforms = JSON.stringify(spec.transform);
    expect(transforms).toContain("datum['{{ .params.y_axis }}']");
    expect(transforms).toContain('"field":"{{ .params.y_axis }}"');
    const stripped = json.replace(/\{\{[^}]*\}\}/g, "");
    expect(stripped).not.toContain("bid_price_sum");
  });

  it("keeps hard-coded labels that only read like a display name", () => {
    // The waterfall legend is titled "Type" by the compiler, on an encoding bound to a derived
    // column. Only a title on the encoding that binds the field is treated as its display name.
    const { spec } = eject(
      {
        chartType: "Waterfall Chart",
        encodings: { x: { field: "publisher" }, y: { field: "bid_price_sum" } },
      },
      ["publisher", "bid_price_sum"],
    );

    const legendTitles = JSON.stringify(spec.layer).match(
      /"legend":\s*\{\s*"title":\s*"([^"]*)"/,
    );
    expect(legendTitles?.[1]).toBe("Type");
  });

  it("templates the time dimension's substituted display name", () => {
    const { spec } = eject(
      {
        chartType: "Line Chart",
        encodings: {
          x: { field: "timestamp" },
          y: { field: "bid_price_sum" },
        },
      },
      ["timestamp", "bid_price_sum"],
    );

    expect(spec.encoding.x.field).toBe("{{ .params.time_dim }}");
    // A metrics view has no display name for its timestamp column, so charts label it "Time".
    expect(spec.encoding.x.title).toBe("{{ .fields.time_dim.display_name }}");
  });

  it("does not split one field name inside another", () => {
    const compiled = {
      encoding: {
        x: { field: "price" },
        y: { field: "price_sum" },
        color: { field: "price_sum_ratio" },
      },
    };
    const overlapping: EjectFieldBinding[] = [
      { param: "a", field: "price", type: "measure" },
      { param: "b", field: "price_sum", type: "measure" },
    ];

    const spec = JSON.parse(ejectToVegaSpec(compiled, overlapping)) as Record<
      string,
      any
    >;

    expect(spec.encoding.x.field).toBe("{{ .params.a }}");
    expect(spec.encoding.y.field).toBe("{{ .params.b }}");
    // Not a bound field, so it stays as it is rather than being partly rewritten.
    expect(spec.encoding.color.field).toBe("price_sum_ratio");
  });

  it("does not rewrite a template reference it just inserted", () => {
    // A param named after the field it binds makes the inserted reference contain the needle.
    const spec = JSON.parse(
      ejectToVegaSpec({ encoding: { x: { field: "publisher" } } }, [
        { param: "publisher", field: "publisher", type: "dimension" },
      ]),
    ) as Record<string, any>;

    expect(spec.encoding.x.field).toBe("{{ .params.publisher }}");
  });
});
