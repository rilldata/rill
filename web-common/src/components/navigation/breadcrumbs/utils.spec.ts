import { describe, expect, it } from "vitest";
import { getCarryOverSubRoute } from "./utils";

describe("getCarryOverSubRoute", () => {
  it("carries over a plain sub-route under the new prefix", () => {
    expect(
      getCarryOverSubRoute("/[organization]/-/settings", "/[organization]"),
    ).toBe("/-/settings");
  });

  it("returns '' when the sub-route contains a variable segment", () => {
    expect(
      getCarryOverSubRoute(
        "/[organization]/[project]/explore/[name]",
        "/[organization]",
      ),
    ).toBe("");
  });

  it("returns '' when there is no sub-route past the new prefix", () => {
    expect(getCarryOverSubRoute("/[organization]", "/[organization]")).toBe("");
  });

  it("returns '' for edit session sub-routes", () => {
    expect(
      getCarryOverSubRoute(
        "/[organization]/[project]/-/edit/files",
        "/[organization]/[project]",
      ),
    ).toBe("");
  });
});
