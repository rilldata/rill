import {
  getDashboardPath,
  getExploreName,
  getSpecQuery,
} from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { describe, expect, it } from "vitest";

describe("getExploreName", () => {
  it("decodes a simple explore name", () => {
    expect(getExploreName("/explore/my_explore")).toBe("my_explore");
  });

  it("decodes spaces encoded as %20", () => {
    expect(getExploreName("/explore/My%20Explore")).toBe("My Explore");
  });

  it("preserves parentheses and hyphens (does not truncate at non-word characters)", () => {
    expect(
      getExploreName(
        "/explore/Sales%20Overview%20-%20Region%20(EMEA%20under%20review)",
      ),
    ).toBe("Sales Overview - Region (EMEA under review)");
  });

  it("stops at a query string boundary", () => {
    expect(
      getExploreName("/explore/My%20Explore?execution_time=2026-06-04"),
    ).toBe("My Explore");
  });

  it("stops at a trailing slash", () => {
    expect(getExploreName("/explore/My%20Explore/")).toBe("My Explore");
  });

  it("returns an empty string when the path has no explore segment", () => {
    expect(getExploreName("/canvas/my_canvas")).toBe("");
  });
});

describe("getDashboardPath", () => {
  it("builds Rill Cloud paths", () => {
    const ctx = { organization: "org", project: "proj" };
    expect(getDashboardPath(ctx, "explore", "e")).toBe("/org/proj/explore/e");
    expect(getDashboardPath(ctx, "canvas", "c")).toBe("/org/proj/canvas/c");
  });

  it("builds Rill Cloud share paths when there is a token", () => {
    const ctx = { organization: "org", project: "proj", token: "tok" };
    expect(getDashboardPath(ctx, "explore", "e")).toBe(
      "/org/proj/-/share/tok/explore/e",
    );
    expect(getDashboardPath(ctx, "canvas", "c")).toBe(
      "/org/proj/-/share/tok/canvas/c",
    );
  });

  it("builds embed paths", () => {
    const ctx = { embedded: true };
    expect(getDashboardPath(ctx, "explore", "e")).toBe("/-/embed/explore/e");
    expect(getDashboardPath(ctx, "canvas", "c")).toBe("/-/embed/canvas/c");
  });

  it("escapes the dashboard name", () => {
    const ctx = { organization: "org", project: "proj" };
    expect(getDashboardPath(ctx, "explore", "sales#2026")).toBe(
      "/org/proj/explore/sales%232026",
    );
    expect(getDashboardPath(ctx, "canvas", "sales?region")).toBe(
      "/org/proj/canvas/sales%3Fregion",
    );
  });
});

describe("getSpecQuery", () => {
  it("prefers resolver properties", () => {
    expect(
      getSpecQuery({
        resolverProperties: { query_name: "A", query_args_json: "{}" },
        queryName: "B",
        queryArgsJson: "[]",
      }),
    ).toEqual({ queryName: "A", queryArgsJson: "{}" });
  });

  it("falls back to the legacy fields", () => {
    expect(getSpecQuery({ queryName: "B", queryArgsJson: "[]" })).toEqual({
      queryName: "B",
      queryArgsJson: "[]",
    });
    expect(
      getSpecQuery({
        resolverProperties: { query_name: "", query_args_json: "" },
        queryName: "B",
        queryArgsJson: "[]",
      }),
    ).toEqual({ queryName: "B", queryArgsJson: "[]" });
  });

  it("returns empty strings without a query", () => {
    expect(getSpecQuery({})).toEqual({ queryName: "", queryArgsJson: "" });
  });
});
