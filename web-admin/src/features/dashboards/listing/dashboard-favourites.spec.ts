import { describe, expect, it } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  migrateLegacyDashboardFavourites,
  migrateLegacyRecentlyUsedDashboards,
} from "./dashboard-favourites";

const explore = (name: string): V1Resource => ({
  meta: { name: { kind: "rill.runtime.v1.Explore", name } },
});
const canvas = (name: string): V1Resource => ({
  meta: { name: { kind: "rill.runtime.v1.Canvas", name } },
});

describe("migrateLegacyDashboardFavourites", () => {
  const dashboards = [explore("Sales"), canvas("sales"), canvas("Ops")];

  it("returns undefined when every favourite already has a kind", () => {
    expect(
      migrateLegacyDashboardFavourites(
        ["rill.runtime.v1.Canvas/ops"],
        dashboards,
      ),
    ).toBeUndefined();
  });

  it("rewrites legacy names to kind/name, keeping order and unmatched names", () => {
    expect(
      migrateLegacyDashboardFavourites(
        ["ops", "unlisted", "rill.runtime.v1.Explore/sales", "ops"],
        dashboards,
      ),
    ).toEqual([
      "rill.runtime.v1.Canvas/ops",
      "unlisted",
      "rill.runtime.v1.Explore/sales",
    ]);
  });

  it("returns undefined when only unmatched legacy names remain", () => {
    expect(
      migrateLegacyDashboardFavourites(
        ["unlisted", "rill.runtime.v1.Canvas/ops"],
        dashboards,
      ),
    ).toBeUndefined();
  });

  it("expands a legacy name shared across kinds to every matching dashboard", () => {
    expect(migrateLegacyDashboardFavourites(["sales"], dashboards)).toEqual([
      "rill.runtime.v1.Explore/sales",
      "rill.runtime.v1.Canvas/sales",
    ]);
  });
});

describe("migrateLegacyRecentlyUsedDashboards", () => {
  const dashboards = [explore("Sales"), canvas("sales"), canvas("Ops")];

  it("returns undefined when every key already has a kind", () => {
    expect(
      migrateLegacyRecentlyUsedDashboards(
        { "rill.runtime.v1.Canvas/ops": 5 },
        dashboards,
      ),
    ).toBeUndefined();
  });

  it("rewrites legacy names, keeps unmatched names and the newest timestamp", () => {
    expect(
      migrateLegacyRecentlyUsedDashboards(
        { ops: 10, unlisted: 3, "rill.runtime.v1.Canvas/ops": 7 },
        dashboards,
      ),
    ).toEqual({ "rill.runtime.v1.Canvas/ops": 10, unlisted: 3 });
  });

  it("returns undefined when only unmatched legacy names remain", () => {
    expect(
      migrateLegacyRecentlyUsedDashboards(
        { unlisted: 3, "rill.runtime.v1.Canvas/ops": 7 },
        dashboards,
      ),
    ).toBeUndefined();
  });

  it("applies a legacy name shared across kinds to every matching dashboard", () => {
    expect(
      migrateLegacyRecentlyUsedDashboards({ sales: 4 }, dashboards),
    ).toEqual({
      "rill.runtime.v1.Explore/sales": 4,
      "rill.runtime.v1.Canvas/sales": 4,
    });
  });
});
