import { describe, it, expect } from "vitest";
import { searchIndex, groupResults } from "./search-orchestrator";
import type { SearchableItem } from "./types";

const items: SearchableItem[] = [
  {
    name: "acme-analytics",
    type: "project",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics",
  },
  {
    name: "acme-marketing",
    type: "project",
    projectName: "acme-marketing",
    orgName: "acme",
    route: "/acme/acme-marketing",
  },
  {
    name: "Revenue Overview",
    type: "explore",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/explore/revenue-overview",
  },
  {
    name: "Campaign Tracker",
    type: "canvas",
    projectName: "acme-marketing",
    orgName: "acme",
    route: "/acme/acme-marketing/canvas/campaign-tracker",
  },
  {
    name: "Weekly Revenue Report",
    type: "report",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/-/reports/weekly-revenue-report",
  },
  {
    name: "Revenue Drop Alert",
    type: "alert",
    projectName: "acme-analytics",
    orgName: "acme",
    route: "/acme/acme-analytics/-/alerts/revenue-drop-alert",
  },
];

describe("searchIndex", () => {
  it("returns empty groups for queries shorter than 2 chars", () => {
    const result = searchIndex(items, "a");
    expect(result.projects).toHaveLength(0);
    expect(result.dashboards).toHaveLength(0);
    expect(result.reports).toHaveLength(0);
    expect(result.alerts).toHaveLength(0);
  });

  it("matches projects by name (case-insensitive)", () => {
    const result = searchIndex(items, "acme");
    expect(result.projects).toHaveLength(2);
  });

  it("matches dashboards by name", () => {
    const result = searchIndex(items, "revenue");
    expect(result.dashboards).toHaveLength(1);
    expect(result.dashboards[0].name).toBe("Revenue Overview");
  });

  it("groups explore and canvas under dashboards", () => {
    const result = searchIndex(items, "er"); // matches "Tracker" and "Overview"
    expect(
      result.dashboards.every(
        (d) => d.type === "explore" || d.type === "canvas",
      ),
    ).toBe(true);
  });

  it("matches reports", () => {
    const result = searchIndex(items, "weekly");
    expect(result.reports).toHaveLength(1);
    expect(result.reports[0].name).toBe("Weekly Revenue Report");
  });

  it("matches alerts", () => {
    const result = searchIndex(items, "drop");
    expect(result.alerts).toHaveLength(1);
    expect(result.alerts[0].name).toBe("Revenue Drop Alert");
  });

  it("limits results to 5 per group", () => {
    const manyProjects: SearchableItem[] = Array.from({ length: 10 }, (_, i) => ({
      name: `project-${i}`,
      type: "project" as const,
      projectName: `project-${i}`,
      orgName: "acme",
      route: `/acme/project-${i}`,
    }));
    const result = searchIndex(manyProjects, "project");
    expect(result.projects).toHaveLength(5);
  });

  it("returns empty groups for empty query", () => {
    const result = searchIndex(items, "");
    expect(result.projects).toHaveLength(0);
    expect(result.dashboards).toHaveLength(0);
  });

  it("matches across name and project name", () => {
    const result = searchIndex(items, "marketing");
    expect(result.projects).toHaveLength(1);
    expect(result.dashboards).toHaveLength(1);
  });
});

describe("groupResults", () => {
  it("separates items into correct groups", () => {
    const grouped = groupResults(items);
    expect(grouped.projects).toHaveLength(2);
    expect(grouped.dashboards).toHaveLength(2);
    expect(grouped.reports).toHaveLength(1);
    expect(grouped.alerts).toHaveLength(1);
  });
});
