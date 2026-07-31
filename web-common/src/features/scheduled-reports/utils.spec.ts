import {
  getDashboardNameFromReport,
  getExistingReportInitialFormValues,
  getNewCanvasReportInitialFormValues,
  isCanvasReportSpec,
} from "@rilldata/web-common/features/scheduled-reports/utils";
import {
  V1ExportFormat,
  type V1ReportSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("getDashboardNameFromReport", () => {
  it("returns the canvas annotation for canvas reports", () => {
    const reportSpec: V1ReportSpec = {
      annotations: { canvas: "my_canvas" },
    };
    expect(getDashboardNameFromReport(reportSpec)).toEqual("my_canvas");
  });

  it("returns the explore annotation for explore reports", () => {
    const reportSpec: V1ReportSpec = {
      annotations: { explore: "my_explore" },
    };
    expect(getDashboardNameFromReport(reportSpec)).toEqual("my_explore");
  });

  it("falls back to the metrics view from the query args", () => {
    const reportSpec: V1ReportSpec = {
      annotations: {},
      queryArgsJson: JSON.stringify({ metrics_view: "my_metrics_view" }),
    };
    expect(getDashboardNameFromReport(reportSpec)).toEqual("my_metrics_view");
  });

  it("does not throw for query-less reports with empty queryArgsJson", () => {
    // Proto string fields default to "", which must not be passed to JSON.parse.
    const reportSpec: V1ReportSpec = {
      annotations: {},
      queryArgsJson: "",
    };
    expect(getDashboardNameFromReport(reportSpec)).toEqual("");
  });
});

describe("isCanvasReportSpec", () => {
  it("detects canvas reports via the canvas annotation", () => {
    expect(isCanvasReportSpec({ annotations: { canvas: "c1" } })).toBe(true);
    expect(isCanvasReportSpec({ annotations: { explore: "e1" } })).toBe(false);
    expect(isCanvasReportSpec({})).toBe(false);
  });
});

describe("getNewCanvasReportInitialFormValues", () => {
  it("defaults to PDF format with all PDF options enabled", () => {
    const values = getNewCanvasReportInitialFormValues("user@example.com");
    expect(values.exportFormat).toEqual(V1ExportFormat.EXPORT_FORMAT_PDF);
    expect(values.pdfIncludeFilters).toBe(true);
    expect(values.pdfAllTabs).toBe(true);
    expect(values.emailRecipients[0]).toEqual("user@example.com");
    expect(values.rows).toEqual([]);
    expect(values.columns).toEqual([]);
  });
});

describe("getExistingReportInitialFormValues", () => {
  it("extracts PDF options from the web_open_state annotation", () => {
    const reportSpec: V1ReportSpec = {
      displayName: "My Canvas Report",
      exportFormat: V1ExportFormat.EXPORT_FORMAT_PDF,
      annotations: {
        canvas: "c1",
        web_open_state: "tr=P7D&pdf_include_filters=false&pdf_all_tabs=true",
      },
    };
    const values = getExistingReportInitialFormValues(
      reportSpec,
      "user@example.com",
      {},
    );
    expect(values.exportFormat).toEqual(V1ExportFormat.EXPORT_FORMAT_PDF);
    expect(values.pdfIncludeFilters).toBe(false);
    expect(values.pdfAllTabs).toBe(true);
  });

  it("defaults PDF options to true when the state has none", () => {
    const values = getExistingReportInitialFormValues(
      { annotations: {} },
      "user@example.com",
      {},
    );
    expect(values.pdfIncludeFilters).toBe(true);
    expect(values.pdfAllTabs).toBe(true);
  });
});
