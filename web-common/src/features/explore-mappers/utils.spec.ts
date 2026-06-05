import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils.ts";
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
    expect(getExploreName("/explore/My%20Explore?execution_time=2026-06-04")).toBe(
      "My Explore",
    );
  });

  it("stops at a trailing slash", () => {
    expect(getExploreName("/explore/My%20Explore/")).toBe("My Explore");
  });

  it("returns an empty string when the path has no explore segment", () => {
    expect(getExploreName("/canvas/my_canvas")).toBe("");
  });
});
