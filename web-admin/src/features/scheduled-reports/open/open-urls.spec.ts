import { describe, expect, it } from "vitest";
import {
  EmbedDashboardPaths,
  getCanvasReportOpenUrl,
  getCloudDashboardPaths,
  getOpenExploreName,
  getOpenQuery,
} from "./open-urls";

describe("dashboard paths", () => {
  it("builds Rill Cloud paths", () => {
    const paths = getCloudDashboardPaths("org", "proj");
    expect(paths.explore("e")).toBe("/org/proj/explore/e");
    expect(paths.canvas("c")).toBe("/org/proj/canvas/c");
  });

  it("builds Rill Cloud share paths when there is a token", () => {
    const paths = getCloudDashboardPaths("org", "proj", "tok");
    expect(paths.explore("e")).toBe("/org/proj/-/share/tok/explore/e");
    expect(paths.canvas("c")).toBe("/org/proj/-/share/tok/canvas/c");
  });

  it("builds embed paths", () => {
    expect(EmbedDashboardPaths.explore("e")).toBe("/-/embed/explore/e");
    expect(EmbedDashboardPaths.canvas("c")).toBe("/-/embed/canvas/c");
  });
});

describe("getCanvasReportOpenUrl", () => {
  it("returns undefined for explore reports", () => {
    expect(
      getCanvasReportOpenUrl(
        { annotations: { explore: "e" } },
        EmbedDashboardPaths,
      ),
    ).toBeUndefined();
  });

  it("applies the captured canvas state", () => {
    expect(
      getCanvasReportOpenUrl(
        { annotations: { canvas: "c", web_open_state: "tr=P7D&f=x" } },
        EmbedDashboardPaths,
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
        EmbedDashboardPaths,
      ),
    ).toBe("/-/embed/canvas/c?tr=P7D");
  });

  it("omits the query string when there is no state", () => {
    expect(
      getCanvasReportOpenUrl(
        { annotations: { canvas: "c" } },
        getCloudDashboardPaths("org", "proj", "tok"),
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
});

describe("getOpenQuery", () => {
  it("prefers resolver properties", () => {
    expect(
      getOpenQuery({
        resolverProperties: { query_name: "A", query_args_json: "{}" },
        queryName: "B",
        queryArgsJson: "[]",
      }),
    ).toEqual({ queryName: "A", queryArgsJson: "{}" });
  });

  it("falls back to the legacy fields", () => {
    expect(getOpenQuery({ queryName: "B", queryArgsJson: "[]" })).toEqual({
      queryName: "B",
      queryArgsJson: "[]",
    });
  });
});
