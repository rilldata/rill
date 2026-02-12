import { describe, it, expect } from "vitest";
import { resolveRootCauseErrorMessage } from "./error-utils";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

function makeResource(
  kind: string,
  name: string,
  opts: {
    reconcileError?: string;
    refs?: { kind: string; name: string }[];
  } = {},
): V1Resource {
  return {
    meta: {
      name: { kind, name },
      reconcileError: opts.reconcileError,
      refs: opts.refs,
    },
  };
}

const MODEL = "rill.runtime.v1.Model";
const METRICS_VIEW = "rill.runtime.v1.MetricsView";
const EXPLORE = "rill.runtime.v1.Explore";

describe("resolveRootCauseErrorMessage", () => {
  it("returns the root cause error when a direct dependency has an error", () => {
    const metricsView = makeResource(METRICS_VIEW, "orders_metrics", {
      reconcileError: 'table "orders" does not exist',
    });
    const explore = makeResource(EXPLORE, "orders_explore", {
      reconcileError: "dependency error",
      refs: [{ kind: METRICS_VIEW, name: "orders_metrics" }],
    });

    const result = resolveRootCauseErrorMessage(
      explore,
      [metricsView, explore],
      "dependency error",
    );

    expect(result).toBe(
      'Error in dependency orders_metrics: table "orders" does not exist',
    );
  });

  it("traverses multiple levels to find the root cause", () => {
    const model = makeResource(MODEL, "orders_model", {
      reconcileError: "invalid SQL: syntax error at position 42",
    });
    const metricsView = makeResource(METRICS_VIEW, "orders_metrics", {
      reconcileError: "dependency error",
      refs: [{ kind: MODEL, name: "orders_model" }],
    });
    const explore = makeResource(EXPLORE, "orders_explore", {
      reconcileError: "dependency error",
      refs: [{ kind: METRICS_VIEW, name: "orders_metrics" }],
    });

    const result = resolveRootCauseErrorMessage(
      explore,
      [model, metricsView, explore],
      "dependency error",
    );

    expect(result).toBe(
      "Error in dependency orders_model: invalid SQL: syntax error at position 42",
    );
  });

  it("returns the original error when the resource has no refs", () => {
    const model = makeResource(MODEL, "orders_model", {
      reconcileError: "invalid SQL",
    });

    const result = resolveRootCauseErrorMessage(model, [model], "invalid SQL");

    expect(result).toBe("invalid SQL");
  });

  it("returns the original error when no refs have errors", () => {
    const model = makeResource(MODEL, "orders_model");
    const metricsView = makeResource(METRICS_VIEW, "orders_metrics", {
      reconcileError: "invalid measure expression",
      refs: [{ kind: MODEL, name: "orders_model" }],
    });

    const result = resolveRootCauseErrorMessage(
      metricsView,
      [model, metricsView],
      "invalid measure expression",
    );

    expect(result).toBe("invalid measure expression");
  });

  it("returns the original error when refs list is empty", () => {
    const explore = makeResource(EXPLORE, "orders_explore", {
      reconcileError: "some error",
      refs: [],
    });

    const result = resolveRootCauseErrorMessage(
      explore,
      [explore],
      "some error",
    );

    expect(result).toBe("some error");
  });

  it("uses the first errored ref when multiple refs have errors", () => {
    const modelA = makeResource(MODEL, "orders_model", {
      reconcileError: "error A",
    });
    const modelB = makeResource(MODEL, "returns_model", {
      reconcileError: "error B",
    });
    const metricsView = makeResource(METRICS_VIEW, "orders_metrics", {
      reconcileError: "dependency error",
      refs: [
        { kind: MODEL, name: "orders_model" },
        { kind: MODEL, name: "returns_model" },
      ],
    });

    const result = resolveRootCauseErrorMessage(
      metricsView,
      [modelA, modelB, metricsView],
      "dependency error",
    );

    expect(result).toBe("Error in dependency orders_model: error A");
  });

  it("skips refs not found in allResources", () => {
    const model = makeResource(MODEL, "orders_model", {
      reconcileError: "invalid SQL",
    });
    const metricsView = makeResource(METRICS_VIEW, "orders_metrics", {
      reconcileError: "dependency error",
      refs: [
        { kind: MODEL, name: "deleted_model" },
        { kind: MODEL, name: "orders_model" },
      ],
    });

    const result = resolveRootCauseErrorMessage(
      metricsView,
      [model, metricsView],
      "dependency error",
    );

    expect(result).toBe("Error in dependency orders_model: invalid SQL");
  });
});
