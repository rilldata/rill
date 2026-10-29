import { getExistingAlertInitialFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values";
import {
  V1ExportFormat,
  type V1AlertSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

function getAlertSpec(overrides: Partial<V1AlertSpec>): V1AlertSpec {
  return {
    displayName: "Alert",
    resolverProperties: {
      query_name: "MetricsViewAggregation",
      query_args_json: JSON.stringify({
        metricsView: "mv1",
        measures: [{ name: "total" }],
        dimensions: [],
      }),
    },
    annotations: { web_open_path: "/explore/e1" },
    ...overrides,
  };
}

describe("getExistingAlertInitialFormValues", () => {
  it("reads the attachment options", () => {
    const values = getExistingAlertInitialFormValues(
      getAlertSpec({
        exportAttach: true,
        exportFormat: V1ExportFormat.EXPORT_FORMAT_XLSX,
        exportLimit: "100",
        exportIncludeHeader: true,
      }),
      "mv1",
    );
    expect(values.exportAttach).toBe(true);
    expect(values.exportFormat).toEqual(V1ExportFormat.EXPORT_FORMAT_XLSX);
    expect(values.exportLimit).toEqual("100");
    expect(values.exportIncludeHeader).toBe(true);
  });

  it("defaults to CSV without a row limit for alerts without an attachment", () => {
    const values = getExistingAlertInitialFormValues(getAlertSpec({}), "mv1");
    expect(values.exportAttach).toBe(false);
    expect(values.exportFormat).toEqual(V1ExportFormat.EXPORT_FORMAT_CSV);
    expect(values.exportLimit).toEqual("");
    expect(values.exportIncludeHeader).toBe(false);
  });
});
