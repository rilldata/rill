import { describe, it, expect } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { filterResourcesForDisplay, buildModelResourcesMap } from "./selectors";

function makeResource(
  kind: string,
  name: string,
  hidden?: boolean,
): V1Resource {
  return {
    meta: {
      name: { kind, name },
      hidden,
    },
  };
}

function makeModelResource(name: string, resultTable?: string): V1Resource {
  return {
    meta: { name: { kind: ResourceKind.Model, name } },
    model: {
      state: { resultTable },
    },
  };
}

describe("filterResourcesForDisplay", () => {
  it("returns empty array for undefined input", () => {
    expect(filterResourcesForDisplay(undefined)).toEqual([]);
  });

  it("returns empty array for empty input", () => {
    expect(filterResourcesForDisplay([])).toEqual([]);
  });

  it("keeps Source, Model, MetricsView, Explore resources", () => {
    const resources = [
      makeResource(ResourceKind.Source, "s1"),
      makeResource(ResourceKind.Model, "m1"),
      makeResource(ResourceKind.MetricsView, "mv1"),
      makeResource(ResourceKind.Explore, "e1"),
    ];
    expect(filterResourcesForDisplay(resources)).toHaveLength(4);
  });

  it("filters out ProjectParser", () => {
    const resources = [
      makeResource(ResourceKind.Source, "s1"),
      makeResource(ResourceKind.ProjectParser, "parser"),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("s1");
  });

  it("filters out RefreshTrigger", () => {
    const resources = [
      makeResource(ResourceKind.Model, "m1"),
      makeResource(ResourceKind.RefreshTrigger, "trigger1"),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(1);
  });

  it("filters out Component", () => {
    const resources = [
      makeResource(ResourceKind.Model, "m1"),
      makeResource(ResourceKind.Component, "comp1"),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(1);
  });

  it("filters out Migration", () => {
    const resources = [
      makeResource(ResourceKind.Model, "m1"),
      makeResource(ResourceKind.Migration, "mig1"),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(1);
  });

  it("filters out hidden resources", () => {
    const resources = [
      makeResource(ResourceKind.Source, "visible"),
      makeResource(ResourceKind.Source, "hidden_one", true),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("visible");
  });

  it("filters out all internal kinds at once", () => {
    const resources = [
      makeResource(ResourceKind.Source, "keep"),
      makeResource(ResourceKind.ProjectParser, "pp"),
      makeResource(ResourceKind.RefreshTrigger, "rt"),
      makeResource(ResourceKind.Component, "c"),
      makeResource(ResourceKind.Migration, "m"),
      makeResource(ResourceKind.Model, "keep2"),
    ];
    const result = filterResourcesForDisplay(resources);
    expect(result).toHaveLength(2);
    expect(result.map((r) => r.meta?.name?.name)).toEqual(["keep", "keep2"]);
  });
});

describe("buildModelResourcesMap", () => {
  it("returns empty map for undefined input", () => {
    const map = buildModelResourcesMap(undefined);
    expect(map.size).toBe(0);
  });

  it("returns empty map for empty input", () => {
    const map = buildModelResourcesMap([]);
    expect(map.size).toBe(0);
  });

  it("maps by resultTable (case-insensitive)", () => {
    const resource = makeModelResource("my_model", "MY_TABLE");
    const map = buildModelResourcesMap([resource]);

    expect(map.get("my_table")).toBe(resource);
  });

  it("maps by model name as fallback (case-insensitive)", () => {
    const resource = makeModelResource("My_Model", "result_tbl");
    const map = buildModelResourcesMap([resource]);

    expect(map.get("my_model")).toBe(resource);
    expect(map.get("result_tbl")).toBe(resource);
  });

  it("handles resource without resultTable (indexed by name only)", () => {
    const resource = makeModelResource("orphan_model");
    const map = buildModelResourcesMap([resource]);

    expect(map.get("orphan_model")).toBe(resource);
    expect(map.size).toBe(1);
  });

  it("maps multiple models correctly", () => {
    const r1 = makeModelResource("model_a", "table_a");
    const r2 = makeModelResource("model_b", "table_b");
    const map = buildModelResourcesMap([r1, r2]);

    expect(map.get("table_a")).toBe(r1);
    expect(map.get("model_a")).toBe(r1);
    expect(map.get("table_b")).toBe(r2);
    expect(map.get("model_b")).toBe(r2);
  });

  it("later entry overwrites earlier when keys collide", () => {
    const r1 = makeModelResource("shared_name", "table_1");
    const r2 = makeModelResource("shared_name", "table_2");
    const map = buildModelResourcesMap([r1, r2]);

    // model name key "shared_name" should point to the last one
    expect(map.get("shared_name")).toBe(r2);
    // but resultTable keys are distinct
    expect(map.get("table_1")).toBe(r1);
    expect(map.get("table_2")).toBe(r2);
  });
});
