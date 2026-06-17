import { describe, it, expect } from "vitest";
import { buildRoute } from "./route-builders";

describe("buildRoute", () => {
  it("builds project route", () => {
    expect(buildRoute("project", "acme", "analytics", "analytics")).toBe(
      "/acme/analytics",
    );
  });

  it("builds explore dashboard route", () => {
    expect(
      buildRoute("explore", "acme", "analytics", "revenue-overview"),
    ).toBe("/acme/analytics/explore/revenue-overview");
  });

  it("builds canvas dashboard route", () => {
    expect(
      buildRoute("canvas", "acme", "analytics", "campaign-tracker"),
    ).toBe("/acme/analytics/canvas/campaign-tracker");
  });

  it("builds report route", () => {
    expect(buildRoute("report", "acme", "analytics", "weekly-report")).toBe(
      "/acme/analytics/-/reports/weekly-report",
    );
  });

  it("builds alert route", () => {
    expect(buildRoute("alert", "acme", "analytics", "revenue-drop")).toBe(
      "/acme/analytics/-/alerts/revenue-drop",
    );
  });
});
