import { describe, it, expect } from "vitest";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getResourceStatus, filterResources } from "./utils";

function makeResource(
  kind: string,
  name: string,
  opts?: {
    reconcileError?: string;
    reconcileStatus?: V1ReconcileStatus;
  },
): V1Resource {
  return {
    meta: {
      name: { kind, name },
      reconcileError: opts?.reconcileError ?? "",
      reconcileStatus:
        opts?.reconcileStatus ?? V1ReconcileStatus.RECONCILE_STATUS_IDLE,
    },
  };
}

describe("getResourceStatus", () => {
  it("returns 'error' when resource has reconcileError", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileError: "something broke",
    });
    expect(getResourceStatus(r)).toBe("error");
  });

  it("returns 'warn' for PENDING status", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_PENDING,
    });
    expect(getResourceStatus(r)).toBe("warn");
  });

  it("returns 'warn' for RUNNING status", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    });
    expect(getResourceStatus(r)).toBe("warn");
  });

  it("returns 'ok' for IDLE status", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
    });
    expect(getResourceStatus(r)).toBe("ok");
  });

  it("returns 'ok' for UNSPECIFIED status", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED,
    });
    expect(getResourceStatus(r)).toBe("ok");
  });

  it("error takes priority over warn status", () => {
    const r = makeResource(ResourceKind.Model, "m1", {
      reconcileError: "error msg",
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    });
    expect(getResourceStatus(r)).toBe("error");
  });
});

describe("filterResources", () => {
  const resources = [
    makeResource(ResourceKind.Source, "my_source"),
    makeResource(ResourceKind.Model, "my_model"),
    makeResource(ResourceKind.Model, "errored_model", {
      reconcileError: "failed",
    }),
    makeResource(ResourceKind.MetricsView, "my_metrics", {
      reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_PENDING,
    }),
  ];

  it("returns all resources with no filters", () => {
    const result = filterResources(resources, [], "", []);
    expect(result).toHaveLength(4);
  });

  it("returns empty for undefined input", () => {
    expect(filterResources(undefined, [], "", [])).toEqual([]);
  });

  it("filters by single kind", () => {
    const result = filterResources(resources, [ResourceKind.Model], "", []);
    expect(result).toHaveLength(2);
    expect(result.every((r) => r.meta?.name?.kind === ResourceKind.Model)).toBe(
      true,
    );
  });

  it("filters by multiple kinds", () => {
    const result = filterResources(
      resources,
      [ResourceKind.Source, ResourceKind.MetricsView],
      "",
      [],
    );
    expect(result).toHaveLength(2);
  });

  it("filters by search text (case-insensitive)", () => {
    const result = filterResources(resources, [], "MY_MOD", []);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("my_model");
  });

  it("filters by status 'error'", () => {
    const result = filterResources(resources, [], "", ["error"]);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("errored_model");
  });

  it("filters by status 'warn'", () => {
    const result = filterResources(resources, [], "", ["warn"]);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("my_metrics");
  });

  it("filters by status 'ok'", () => {
    const result = filterResources(resources, [], "", ["ok"]);
    expect(result).toHaveLength(2);
  });

  it("combines kind + status + search filters", () => {
    const result = filterResources(resources, [ResourceKind.Model], "errored", [
      "error",
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].meta?.name?.name).toBe("errored_model");
  });

  it("returns empty when no resources match", () => {
    const result = filterResources(resources, [], "nonexistent", []);
    expect(result).toEqual([]);
  });
});
