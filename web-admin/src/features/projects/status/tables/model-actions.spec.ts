import { describe, it, expect } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { getAvailableModelActions } from "./model-actions";

function makeModel(opts?: {
  incremental?: boolean;
  partitionsResolver?: string;
  partitionsModelId?: string;
  partitionsHaveErrors?: boolean;
}): V1Resource {
  return {
    meta: { name: { kind: "rill.runtime.v1.Model", name: "test_model" } },
    model: {
      spec: {
        incremental: opts?.incremental,
        partitionsResolver: opts?.partitionsResolver,
      },
      state: {
        resultTable: "test_model",
        partitionsModelId: opts?.partitionsModelId,
        partitionsHaveErrors: opts?.partitionsHaveErrors,
      },
    },
  };
}

describe("getAvailableModelActions", () => {
  it("returns empty array for undefined resource", () => {
    expect(getAvailableModelActions(undefined)).toEqual([]);
  });

  it("returns base actions for a plain model", () => {
    const resource = makeModel();
    const actions = getAvailableModelActions(resource);

    expect(actions).toEqual(["describe", "viewLogs", "fullRefresh"]);
    expect(actions).not.toContain("incrementalRefresh");
    expect(actions).not.toContain("viewPartitions");
    expect(actions).not.toContain("refreshErrored");
  });

  it("includes incrementalRefresh for incremental model", () => {
    const resource = makeModel({ incremental: true });
    const actions = getAvailableModelActions(resource);

    expect(actions).toContain("incrementalRefresh");
    expect(actions).toContain("fullRefresh");
    expect(actions).not.toContain("viewPartitions");
  });

  it("includes viewPartitions for partitioned model", () => {
    const resource = makeModel({ partitionsResolver: "sql" });
    const actions = getAvailableModelActions(resource);

    expect(actions).toContain("viewPartitions");
    expect(actions).not.toContain("incrementalRefresh");
    expect(actions).not.toContain("refreshErrored");
  });

  it("includes viewPartitions and incrementalRefresh for partitioned + incremental model", () => {
    const resource = makeModel({
      partitionsResolver: "sql",
      incremental: true,
    });
    const actions = getAvailableModelActions(resource);

    expect(actions).toContain("viewPartitions");
    expect(actions).toContain("incrementalRefresh");
    expect(actions).toContain("fullRefresh");
    expect(actions).not.toContain("refreshErrored");
  });

  it("includes refreshErrored for partitioned model with errored partitions", () => {
    const resource = makeModel({
      partitionsResolver: "sql",
      partitionsModelId: "abc-123",
      partitionsHaveErrors: true,
    });
    const actions = getAvailableModelActions(resource);

    expect(actions).toContain("viewPartitions");
    expect(actions).toContain("refreshErrored");
    expect(actions).toContain("fullRefresh");
  });

  it("includes all actions for partitioned + incremental + errored model", () => {
    const resource = makeModel({
      partitionsResolver: "sql",
      incremental: true,
      partitionsModelId: "abc-123",
      partitionsHaveErrors: true,
    });
    const actions = getAvailableModelActions(resource);

    expect(actions).toContain("describe");
    expect(actions).toContain("viewLogs");
    expect(actions).toContain("viewPartitions");
    expect(actions).toContain("refreshErrored");
    expect(actions).toContain("fullRefresh");
    expect(actions).toContain("incrementalRefresh");
    expect(actions).toHaveLength(6);
  });

  it("does not show refreshErrored when partitionsHaveErrors is true but no partitionsModelId", () => {
    const resource = makeModel({
      partitionsResolver: "sql",
      partitionsHaveErrors: true,
      // no partitionsModelId
    });
    const actions = getAvailableModelActions(resource);

    expect(actions).not.toContain("refreshErrored");
  });
});
