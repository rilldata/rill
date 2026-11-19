import { describe, expect, it } from "vitest";
import { getCanvasReportOpenUrl, getOpenExploreName } from "./open-urls";

const embedded = { embedded: true };

describe("getCanvasReportOpenUrl", () => {
  it("returns undefined for explore reports", () => {
    expect(
      getCanvasReportOpenUrl({ annotations: { explore: "e" } }, embedded),
    ).toBeUndefined();
  });

  it("applies the captured canvas state", () => {
    expect(
      getCanvasReportOpenUrl(
        { annotations: { canvas: "c", web_open_state: "tr=P7D&f=x" } },
        embedded,
      ),
    ).toBe("/-/embed/canvas/c?tr=P7D&f=x");
  });

  it("strips internal report params", () => {
    expect(
      getCanvasReportOpenUrl(
        {
          annotations: {
            canvas: "c",
            web_open_state: "tr=P7D&pdf_include_filters=true&pdf_all_tabs=true",
          },
        },
        embedded,
      ),
    ).toBe("/-/embed/canvas/c?tr=P7D");
  });

  it("omits the query string when there is no state", () => {
    expect(
      getCanvasReportOpenUrl(
        { annotations: { canvas: "c" } },
        { organization: "org", project: "proj", token: "tok" },
      ),
    ).toBe("/org/proj/-/share/tok/canvas/c");
  });
});

describe("getOpenExploreName", () => {
  it("prefers the explore annotation", () => {
    expect(
      getOpenExploreName({
        annotations: { explore: "e", web_open_path: "/explore/old" },
      }),
    ).toBe("e");
  });

  it("falls back to web_open_path", () => {
    expect(
      getOpenExploreName({ annotations: { web_open_path: "/explore/old" } }),
    ).toBe("old");
  });

  it("returns an empty name without annotations", () => {
    expect(getOpenExploreName({})).toBe("");
  });

  it("returns the explore of an alert", () => {
    // Alerts created from the UI only set web_open_path
    expect(
      getOpenExploreName({
        resolverProperties: { query_name: "MetricsViewAggregation" },
        annotations: { web_open_path: "/explore/e1" },
      }),
    ).toBe("e1");
  });
});
