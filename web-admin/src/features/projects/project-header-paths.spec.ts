import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { UNTAGGED_KEY, UNTAGGED_LABEL } from "../dashboards/listing/selectors";
import {
  buildDashboardHref,
  buildDashboardPathOption,
  buildDashboardSubOption,
  buildTagPathsOptions,
  buildVisualizationOptions,
  getAllDashboardTags,
  groupDashboardsByTag,
  hasUntaggedDashboards,
  sortDashboardResources,
} from "./project-header-paths";

function makeExplore(
  name: string,
  tags: string[] = [],
  displayName?: string,
): V1Resource {
  return {
    meta: {
      name: { kind: "rill.runtime.v1.Explore", name },
      tags,
    },
    explore: {
      spec: displayName ? { displayName } : {},
    },
  } as V1Resource;
}

function makeCanvas(
  name: string,
  tags: string[] = [],
  displayName?: string,
): V1Resource {
  return {
    meta: {
      name: { kind: "rill.runtime.v1.Canvas", name },
      tags,
    },
    canvas: {
      spec: displayName ? { displayName } : {},
    },
  } as V1Resource;
}

const ORG = "acme";
const PROJ = "data";

describe("sortDashboardResources", () => {
  it("orders canvas dashboards before explores, then alphabetically", () => {
    const result = sortDashboardResources([
      makeExplore("zeta"),
      makeCanvas("orders"),
      makeExplore("alpha"),
      makeCanvas("billing"),
    ]);
    expect(result.map((r) => r.meta?.name?.name)).toEqual([
      "billing",
      "orders",
      "alpha",
      "zeta",
    ]);
  });

  it("does not mutate the input array", () => {
    const input = [makeExplore("b"), makeExplore("a")];
    const inputBefore = [...input];
    sortDashboardResources(input);
    expect(input).toEqual(inputBefore);
  });
});

describe("getAllDashboardTags", () => {
  it("returns the union of tags across resources, sorted alphabetically", () => {
    const tags = getAllDashboardTags([
      makeExplore("a", ["sales", "marketing"]),
      makeCanvas("b", ["marketing", "ops"]),
      makeExplore("c", []),
    ]);
    expect(tags).toEqual(["marketing", "ops", "sales"]);
  });

  it("returns an empty array when no resources have tags", () => {
    expect(getAllDashboardTags([makeExplore("a"), makeCanvas("b")])).toEqual(
      [],
    );
  });
});

describe("hasUntaggedDashboards", () => {
  it("is true when any resource has no tags", () => {
    expect(
      hasUntaggedDashboards([makeExplore("a", ["sales"]), makeCanvas("b", [])]),
    ).toBe(true);
  });

  it("is false when all resources have at least one tag", () => {
    expect(
      hasUntaggedDashboards([
        makeExplore("a", ["sales"]),
        makeCanvas("b", ["ops"]),
      ]),
    ).toBe(false);
  });
});

describe("groupDashboardsByTag", () => {
  it("places multi-tag dashboards into every tag bucket", () => {
    const sales = makeExplore("a", ["sales", "marketing"]);
    const ops = makeCanvas("b", ["ops"]);
    const result = groupDashboardsByTag([sales, ops]);
    expect(result.get("sales")).toEqual([sales]);
    expect(result.get("marketing")).toEqual([sales]);
    expect(result.get("ops")).toEqual([ops]);
  });

  it("omits untagged dashboards entirely", () => {
    const result = groupDashboardsByTag([
      makeExplore("a", []),
      makeCanvas("b", ["ops"]),
    ]);
    expect(result.has(UNTAGGED_KEY)).toBe(false);
    expect(result.get("ops")?.map((r) => r.meta?.name?.name)).toEqual(["b"]);
  });
});

describe("buildDashboardHref", () => {
  it("builds an explore href with the tag query param", () => {
    expect(buildDashboardHref(makeExplore("orders"), "sales", ORG, PROJ)).toBe(
      "/acme/data/explore/orders?tags=sales",
    );
  });

  it("builds a canvas href with the tag query param", () => {
    expect(buildDashboardHref(makeCanvas("dash"), "ops", ORG, PROJ)).toBe(
      "/acme/data/canvas/dash?tags=ops",
    );
  });

  it("omits the tag param when tag is UNTAGGED_KEY", () => {
    expect(
      buildDashboardHref(makeExplore("orders"), UNTAGGED_KEY, ORG, PROJ),
    ).toBe("/acme/data/explore/orders");
  });

  it("URL-encodes tags with special characters", () => {
    expect(buildDashboardHref(makeExplore("orders"), "a/b c", ORG, PROJ)).toBe(
      "/acme/data/explore/orders?tags=a%2Fb%20c",
    );
  });
});

describe("buildDashboardSubOption", () => {
  it("uses the explore display name when present", () => {
    const [key, option] = buildDashboardSubOption(
      makeExplore("orders", [], "Orders Dashboard"),
      "sales",
      ORG,
      PROJ,
    );
    expect(key).toBe("orders");
    expect(option.label).toBe("Orders Dashboard");
    expect(option.href).toBe("/acme/data/explore/orders?tags=sales");
  });

  it("falls back to the resource name when display name is missing", () => {
    const [, option] = buildDashboardSubOption(
      makeCanvas("billing"),
      UNTAGGED_KEY,
      ORG,
      PROJ,
    );
    expect(option.label).toBe("billing");
    expect(option.href).toBe("/acme/data/canvas/billing");
  });
});

describe("buildTagPathsOptions", () => {
  const explore = makeExplore("orders", ["sales"], "Orders");
  const canvas = makeCanvas("billing", ["sales", "finance"], "Billing");
  const untagged = makeExplore("notes", [], "Notes");
  const sorted = sortDashboardResources([explore, canvas, untagged]);
  const dashboardsByTag = groupDashboardsByTag(sorted);
  const allDashboardTags = getAllDashboardTags(sorted);

  it("includes one entry per tag with sub-options for each dashboard in the tag", () => {
    const map = buildTagPathsOptions({
      allDashboardTags,
      dashboardsByTag,
      hasUntaggedDashboard: false,
      activeTag: undefined,
      organization: ORG,
      project: PROJ,
    });
    const sales = map.get("sales");
    expect(sales.label).toBe("sales");
    expect(sales.href).toBe("/acme/data?tags=sales");
    expect(Array.from(sales.subOptions.keys())).toEqual(["billing", "orders"]);
    expect(sales.subOptions.get("billing").href).toBe(
      "/acme/data/canvas/billing?tags=sales",
    );
  });

  it("includes the untagged entry when any dashboard is untagged", () => {
    const map = buildTagPathsOptions({
      allDashboardTags,
      dashboardsByTag,
      hasUntaggedDashboard: true,
      activeTag: undefined,
      organization: ORG,
      project: PROJ,
    });
    const untaggedEntry = map.get(UNTAGGED_KEY);
    expect(untaggedEntry.label).toBe(UNTAGGED_LABEL);
    expect(untaggedEntry.href).toBe(
      `/acme/data?tags=${encodeURIComponent(UNTAGGED_KEY)}`,
    );
  });

  it("includes the untagged entry when the active tag is UNTAGGED_KEY even without untagged dashboards", () => {
    const map = buildTagPathsOptions({
      allDashboardTags,
      dashboardsByTag,
      hasUntaggedDashboard: false,
      activeTag: UNTAGGED_KEY,
      organization: ORG,
      project: PROJ,
    });
    expect(map.has(UNTAGGED_KEY)).toBe(true);
  });

  it("omits the untagged entry when no untagged dashboard exists and it is not active", () => {
    const map = buildTagPathsOptions({
      allDashboardTags,
      dashboardsByTag,
      hasUntaggedDashboard: false,
      activeTag: "sales",
      organization: ORG,
      project: PROJ,
    });
    expect(map.has(UNTAGGED_KEY)).toBe(false);
  });
});

describe("buildDashboardPathOption", () => {
  it("returns explore section and ResourceKind for an explore", () => {
    const option = buildDashboardPathOption(
      makeExplore("orders", [], "Orders"),
    );
    expect(option.label).toBe("Orders");
    expect(option.section).toBe("explore");
    expect(option.depth).toBe(2);
  });

  it("returns canvas section for a canvas resource", () => {
    const option = buildDashboardPathOption(makeCanvas("billing"));
    expect(option.section).toBe("canvas");
    expect(option.label).toBe("billing");
  });
});

describe("buildVisualizationOptions", () => {
  const explore = makeExplore("orders", ["sales"], "Orders");
  const canvas = makeCanvas("billing", ["finance"], "Billing");
  const sortedVisualizations = sortDashboardResources([explore, canvas]);
  const dashboardsByTag = groupDashboardsByTag(sortedVisualizations);

  it("returns all dashboards when tagAsFolders is off", () => {
    const map = buildVisualizationOptions({
      sortedVisualizations,
      dashboardsByTag,
      activeTag: "sales",
      tagAsFolders: false,
    });
    expect(Array.from(map.keys())).toEqual(["billing", "orders"]);
  });

  it("returns all dashboards when no active tag is set", () => {
    const map = buildVisualizationOptions({
      sortedVisualizations,
      dashboardsByTag,
      activeTag: undefined,
      tagAsFolders: true,
    });
    expect(Array.from(map.keys())).toEqual(["billing", "orders"]);
  });

  it("scopes dashboards to the active tag when tagAsFolders is on", () => {
    const map = buildVisualizationOptions({
      sortedVisualizations,
      dashboardsByTag,
      activeTag: "sales",
      tagAsFolders: true,
    });
    expect(Array.from(map.keys())).toEqual(["orders"]);
  });

  it("returns an empty map when active tag has no dashboards", () => {
    const map = buildVisualizationOptions({
      sortedVisualizations,
      dashboardsByTag,
      activeTag: "missing",
      tagAsFolders: true,
    });
    expect(map.size).toBe(0);
  });
});
