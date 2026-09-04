import { describe, expect, it } from "vitest";
import {
  invalidDimensionEncodingError,
  missingRendererError,
  missingSelectedParamsError,
} from "./import-with-ai";

// The static consistency check the import loop runs on generated components:
// templated metrics_sql is never executed while the component reconciles alone,
// so an ORDER BY or encoding referencing an unselected field param only fails at
// render time. The check catches it during import and feeds it back to the agent.

const PARAMS = [
  { name: "metrics_view", type: "metrics_view" },
  { name: "x_axis", type: "dimension" },
  { name: "y_axis", type: "dimension" },
  { name: "size", type: "measure" },
];

describe("missingRendererError", () => {
  it("flags a component without any renderer (draft-tolerated by the reconciler)", () => {
    const error = missingRendererError(undefined, undefined);
    expect(error).toContain("TOP-LEVEL custom_chart");
  });

  it("flags a non-custom_chart renderer by name", () => {
    const error = missingRendererError("renderer", {});
    expect(error).toContain('"renderer"');
  });

  it("flags a custom_chart missing metrics_sql or spec", () => {
    expect(
      missingRendererError("custom_chart", {
        spec: { chartType: "Bar Chart" },
      }),
    ).toContain("metrics_sql");
    expect(
      missingRendererError("custom_chart", { metrics_sql: "SELECT 1" }),
    ).toContain("spec");
  });

  it("flags a spec with no chartType or no encodings", () => {
    expect(
      missingRendererError("custom_chart", {
        metrics_sql: "SELECT 1",
        spec: { encodings: { x: { field: "a" } } },
      }),
    ).toContain("chartType");
    expect(
      missingRendererError("custom_chart", {
        metrics_sql: "SELECT 1",
        spec: { chartType: "Bar Chart", encodings: {} },
      }),
    ).toContain("encodings");
  });

  it("passes a complete custom_chart", () => {
    expect(
      missingRendererError("custom_chart", {
        metrics_sql: "SELECT 1",
        spec: {
          chartType: "Bar Chart",
          encodings: { x: { field: "{{ .params.x_axis }}" } },
        },
      }),
    ).toBeNull();
  });
});

describe("missingSelectedParamsError", () => {
  it("flags a field param that is ordered by and encoded but not selected", () => {
    // Real failure case: a generated 2D histogram ordered by the size measure
    // without selecting it.
    const error = missingSelectedParamsError(PARAMS, {
      metrics_sql: [
        "SELECT {{ .params.x_axis }}, {{ .params.y_axis }} FROM {{ .params.metrics_view }} ORDER BY {{ .params.size }} DESC LIMIT 10",
      ],
      spec: {
        chartType: "Scatter Plot",
        encodings: {
          x: { field: "{{ .params.x_axis }}" },
          y: { field: "{{ .params.y_axis }}" },
          size: { field: "{{ .params.size }}" },
        },
      },
    });
    expect(error).toContain("{{ .params.size }}");
    expect(error).toContain("SELECT list");
  });

  it("passes when every referenced field param is selected", () => {
    const error = missingSelectedParamsError(PARAMS, {
      metrics_sql: [
        "SELECT {{ .params.x_axis }}, {{ .params.y_axis }}, {{ .params.size }} FROM {{ .params.metrics_view }} ORDER BY {{ .params.size }} DESC LIMIT 10",
      ],
      spec: { encodings: { size: { field: "{{ .params.size }}" } } },
    });
    expect(error).toBeNull();
  });

  it("ignores metrics_view and scalar params referenced outside the SELECT list", () => {
    const error = missingSelectedParamsError(
      [...PARAMS, { name: "limit", type: "number" }],
      {
        metrics_sql:
          "SELECT {{ .params.x_axis }}, {{ .params.size }} FROM {{ .params.metrics_view }} ORDER BY {{ .params.size }} DESC LIMIT {{ .params.limit }}",
        spec: { encodings: { x: { field: "{{ .params.x_axis }}" } } },
      },
    );
    expect(error).toBeNull();
  });

  it("flags quantitative/temporal types hardcoded onto dimension params", () => {
    // Dimensions are categorical, so typing them numerically renders a blank chart.
    const error = invalidDimensionEncodingError(PARAMS, {
      spec: {
        chartType: "Scatter Plot",
        encodings: {
          x: { field: "{{ .params.x_axis }}", type: "quantitative" },
          y: { field: "{{ .params.y_axis }}", type: "temporal" },
          size: { field: "{{ .params.size }}", type: "quantitative" },
        },
      },
    });
    expect(error).toContain('"x_axis"');
    expect(error).toContain('"y_axis"');
    expect(error).not.toContain('"size"'); // Measures may be quantitative.
    expect(error).toContain('remove the "type"');
  });

  it("accepts encodings that leave the type to Rill's semantics", () => {
    const error = invalidDimensionEncodingError(PARAMS, {
      spec: {
        encodings: {
          x: { field: "{{ .params.x_axis }}" },
          y: { field: "{{ .params.y_axis }}", type: "nominal" },
          size: { field: "{{ .params.size }}", type: "quantitative" },
        },
      },
    });
    expect(error).toBeNull();
  });

  it("is lenient when there is no SQL or no field params", () => {
    expect(missingSelectedParamsError(PARAMS, undefined)).toBeNull();
    expect(
      missingSelectedParamsError(PARAMS, { spec: { chartType: "Bar Chart" } }),
    ).toBeNull();
    expect(
      missingSelectedParamsError(
        [{ name: "metrics_view", type: "metrics_view" }],
        {
          metrics_sql: "SELECT a FROM {{ .params.metrics_view }}",
        },
      ),
    ).toBeNull();
  });
});
