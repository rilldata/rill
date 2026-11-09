import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { describe, expect, it } from "vitest";
import { getDashboardFromEmbedRoute } from "./embed-route-utils";

describe("getDashboardFromEmbedRoute", () => {
  it("resolves explore routes", () => {
    expect(
      getDashboardFromEmbedRoute("/-/embed/explore/[name]", { name: "e1" }),
    ).toEqual({ name: "e1", kind: ResourceKind.Explore });
  });

  it("resolves canvas routes", () => {
    expect(
      getDashboardFromEmbedRoute("/-/embed/canvas/[name]", { name: "c1" }),
    ).toEqual({ name: "c1", kind: ResourceKind.Canvas });
  });

  it("resolves personal canvas routes as canvases", () => {
    expect(
      getDashboardFromEmbedRoute("/-/embed/personal/[name]", { name: "p1" }),
    ).toEqual({ name: "p1", kind: ResourceKind.Canvas });
  });

  it("returns null for non-dashboard routes", () => {
    expect(getDashboardFromEmbedRoute("/-/embed", {})).toBeNull();
    expect(
      getDashboardFromEmbedRoute("/-/embed/reports/[report]", {
        report: "r1",
      }),
    ).toBeNull();
  });
});
