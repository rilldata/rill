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

  describe("the creator's saved email", () => {
    const spec = getAlertSpec({
      notifiers: [
        {
          connector: "email",
          properties: {
            recipients: ["Creator@example.com", "other@example.com"],
          },
        },
      ],
      annotations: {
        web_open_path: "/explore/e1",
        admin_embed_owner_id: "ext_abc",
        admin_embed_owner_email: "creator@example.com",
      },
    });

    it("is in its own field rather than in the recipients", () => {
      const values = getExistingAlertInitialFormValues(
        spec,
        "mv1",
        "editor@example.com",
        true,
      );
      expect(values.emailRecipients).toEqual(["other@example.com", ""]);
      // Outside of an embed, the editor's email doesn't replace the saved email
      expect(values.embedOwnerEmail).toEqual("creator@example.com");
    });

    it("is empty when the field isn't shown", () => {
      const values = getExistingAlertInitialFormValues(spec, "mv1", "", false);
      expect(values.embedOwnerEmail).toEqual("");
    });
  });

  it("reads the rows limit", () => {
    const values = getExistingAlertInitialFormValues(
      getAlertSpec({ notifyRowsLimit: 20 }),
      "mv1",
    );
    expect(values.notifyRowsLimit).toEqual("20");
  });

  it("defaults the rows limit for alerts without one", () => {
    const values = getExistingAlertInitialFormValues(getAlertSpec({}), "mv1");
    expect(values.notifyRowsLimit).toEqual("5");
  });
});
