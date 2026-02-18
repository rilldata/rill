import { describe, expect, it } from "vitest";
import {
  filterResources,
  getResourceStatus,
} from "./resource-filter-utils";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

function makeResource(
  name: string,
  kind: string,
  opts: { error?: string; status?: V1ReconcileStatus } = {},
): V1Resource {
  return {
    meta: {
      name: { name, kind },
      reconcileError: opts.error,
      reconcileStatus: opts.status ?? V1ReconcileStatus.RECONCILE_STATUS_IDLE,
    },
  } as V1Resource;
}

describe("getResourceStatus", () => {
  it("returns 'error' when resource has a reconcile error", () => {
    const r = makeResource("src1", ResourceKind.Source, {
      error: "connection failed",
    });
    expect(getResourceStatus(r)).toBe("error");
  });

  it("returns 'warn' when status is PENDING", () => {
    const r = makeResource("model1", ResourceKind.Model, {
      status: V1ReconcileStatus.RECONCILE_STATUS_PENDING,
    });
    expect(getResourceStatus(r)).toBe("warn");
  });

  it("returns 'warn' when status is RUNNING", () => {
    const r = makeResource("model1", ResourceKind.Model, {
      status: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    });
    expect(getResourceStatus(r)).toBe("warn");
  });

  it("returns 'ok' when status is IDLE with no error", () => {
    const r = makeResource("src1", ResourceKind.Source);
    expect(getResourceStatus(r)).toBe("ok");
  });
});

describe("filterResources", () => {
  const resources = [
    makeResource("orders", ResourceKind.Source),
    makeResource("users", ResourceKind.Source),
    makeResource("orders_model", ResourceKind.Model, {
      status: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    }),
    makeResource("dashboard1", ResourceKind.Explore),
    makeResource("broken_source", ResourceKind.Source, {
      error: "parse error",
    }),
  ];

  it("returns all resources when no filters applied", () => {
    const result = filterResources(resources, [], "", []);
    expect(result).toHaveLength(5);
  });

  it("returns empty array for undefined input", () => {
    const result = filterResources(undefined, [], "", []);
    expect(result).toEqual([]);
  });

  it("filters by type", () => {
    const result = filterResources(
      resources,
      [ResourceKind.Source],
      "",
      [],
    );
    expect(result).toHaveLength(3);
    expect(result.every((r) => r.meta?.name?.kind === ResourceKind.Source)).toBe(
      true,
    );
  });

  it("filters by multiple types", () => {
    const result = filterResources(
      resources,
      [ResourceKind.Source, ResourceKind.Explore],
      "",
      [],
    );
    expect(result).toHaveLength(4);
  });

  it("filters by search string (case-insensitive)", () => {
    const result = filterResources(resources, [], "ORDER", []);
    expect(result).toHaveLength(2);
    expect(result.map((r) => r.meta?.name?.name)).toContain("orders");
    expect(result.map((r) => r.meta?.name?.name)).toContain("orders_model");
  });

  it("filters by status", () => {
    const result = filterResources(resources, [], "", ["error"]);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("broken_source");
  });

  it("filters by warn status", () => {
    const result = filterResources(resources, [], "", ["warn"]);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("orders_model");
  });

  it("combines type + search + status filters", () => {
    const result = filterResources(
      resources,
      [ResourceKind.Source],
      "broken",
      ["error"],
    );
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("broken_source");
  });

  it("returns empty when no resources match combined filters", () => {
    const result = filterResources(
      resources,
      [ResourceKind.Explore],
      "",
      ["error"],
    );
    expect(result).toHaveLength(0);
  });
});
