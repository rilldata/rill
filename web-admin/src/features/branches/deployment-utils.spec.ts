import { describe, it, expect } from "vitest";
import { V1DeploymentStatus, type V1Deployment } from "../../client";
import { deduplicateDeployments } from "./deployment-utils";

function makeDeployment(overrides: Partial<V1Deployment> = {}): V1Deployment {
  return {
    id: "d-1",
    branch: "main",
    environment: "dev",
    status: V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
    updatedOn: "2026-01-01T00:00:00Z",
    ...overrides,
  } as V1Deployment;
}

describe("deduplicateDeployments", () => {
  it("returns empty array for empty input", () => {
    expect(deduplicateDeployments([])).toEqual([]);
  });

  it("passes through deployments with distinct branches", () => {
    const a = makeDeployment({ id: "a", branch: "main" });
    const b = makeDeployment({ id: "b", branch: "feature-x" });
    expect(deduplicateDeployments([a, b])).toEqual([a, b]);
  });

  it("keeps the most recently updated deployment per branch", () => {
    const older = makeDeployment({
      id: "old",
      branch: "feature-x",
      updatedOn: "2026-01-01T00:00:00Z",
    });
    const newer = makeDeployment({
      id: "new",
      branch: "feature-x",
      updatedOn: "2026-01-02T00:00:00Z",
    });
    const result = deduplicateDeployments([older, newer]);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("new");
  });

  it("keeps the newer deployment regardless of input order", () => {
    const older = makeDeployment({
      id: "old",
      branch: "feature-x",
      updatedOn: "2026-01-01T00:00:00Z",
    });
    const newer = makeDeployment({
      id: "new",
      branch: "feature-x",
      updatedOn: "2026-01-02T00:00:00Z",
    });
    const result = deduplicateDeployments([newer, older]);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("new");
  });

  it("groups deployments with undefined branch under the same key", () => {
    const a = makeDeployment({
      id: "a",
      branch: undefined,
      updatedOn: "2026-01-01T00:00:00Z",
    });
    const b = makeDeployment({
      id: "b",
      branch: undefined,
      updatedOn: "2026-01-02T00:00:00Z",
    });
    const result = deduplicateDeployments([a, b]);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("b");
  });

  it("applies exclude predicate", () => {
    const active = makeDeployment({ id: "a", branch: "main" });
    const deleted = makeDeployment({
      id: "b",
      branch: "feature-x",
      status: V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    });
    const result = deduplicateDeployments(
      [active, deleted],
      (d) => d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    );
    expect(result).toEqual([active]);
  });

  it("exclude runs before dedup: excluded newer entry does not shadow older", () => {
    const older = makeDeployment({
      id: "old",
      branch: "feature-x",
      updatedOn: "2026-01-01T00:00:00Z",
    });
    const newerDeleted = makeDeployment({
      id: "new",
      branch: "feature-x",
      updatedOn: "2026-01-02T00:00:00Z",
      status: V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    });
    const result = deduplicateDeployments(
      [older, newerDeleted],
      (d) => d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    );
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("old");
  });

  it("handles undefined updatedOn gracefully", () => {
    const noDate = makeDeployment({
      id: "a",
      branch: "feature-x",
      updatedOn: undefined,
    });
    const withDate = makeDeployment({
      id: "b",
      branch: "feature-x",
      updatedOn: "2026-01-01T00:00:00Z",
    });
    const result = deduplicateDeployments([noDate, withDate]);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe("b");
  });
});
