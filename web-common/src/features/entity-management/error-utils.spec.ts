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

describe("resolveRootCauseErrorMessage", () => {
  it("returns the root cause error when a direct dependency has an error", () => {
    const source = makeResource("Source", "raw_data", {
      reconcileError: "connection refused",
    });
    const model = makeResource("Model", "clean_data", {
      reconcileError: "dependency error",
      refs: [{ kind: "Source", name: "raw_data" }],
    });

    const result = resolveRootCauseErrorMessage(
      model,
      [source, model],
      "fallback",
    );

    expect(result).toBe("raw_data: connection refused");
  });

  it("traverses multiple levels to find the root cause", () => {
    const source = makeResource("Source", "raw_data", {
      reconcileError: "503 Service Unavailable",
    });
    const model = makeResource("Model", "clean_data", {
      reconcileError: "dependency error",
      refs: [{ kind: "Source", name: "raw_data" }],
    });
    const explore = makeResource("Explore", "dashboard", {
      reconcileError: "dependency error",
      refs: [{ kind: "Model", name: "clean_data" }],
    });

    const result = resolveRootCauseErrorMessage(
      explore,
      [source, model, explore],
      "fallback",
    );

    expect(result).toBe("raw_data: 503 Service Unavailable");
  });

  it("returns the fallback when the resource has no refs", () => {
    const resource = makeResource("Source", "raw_data", {
      reconcileError: "connection refused",
    });

    const result = resolveRootCauseErrorMessage(
      resource,
      [resource],
      "fallback",
    );

    expect(result).toBe("fallback");
  });

  it("returns the fallback when no refs have errors", () => {
    const source = makeResource("Source", "raw_data");
    const model = makeResource("Model", "clean_data", {
      reconcileError: "some error",
      refs: [{ kind: "Source", name: "raw_data" }],
    });

    const result = resolveRootCauseErrorMessage(
      model,
      [source, model],
      "fallback",
    );

    expect(result).toBe("fallback");
  });

  it("returns the fallback when refs list is empty", () => {
    const resource = makeResource("Model", "clean_data", {
      reconcileError: "some error",
      refs: [],
    });

    const result = resolveRootCauseErrorMessage(
      resource,
      [resource],
      "fallback",
    );

    expect(result).toBe("fallback");
  });

  it("uses the first errored ref when multiple refs have errors", () => {
    const sourceA = makeResource("Source", "source_a", {
      reconcileError: "error A",
    });
    const sourceB = makeResource("Source", "source_b", {
      reconcileError: "error B",
    });
    const model = makeResource("Model", "clean_data", {
      reconcileError: "dependency error",
      refs: [
        { kind: "Source", name: "source_a" },
        { kind: "Source", name: "source_b" },
      ],
    });

    const result = resolveRootCauseErrorMessage(
      model,
      [sourceA, sourceB, model],
      "fallback",
    );

    expect(result).toBe("source_a: error A");
  });

  it("skips refs not found in allResources", () => {
    const source = makeResource("Source", "raw_data", {
      reconcileError: "connection refused",
    });
    const model = makeResource("Model", "clean_data", {
      reconcileError: "dependency error",
      refs: [
        { kind: "Source", name: "missing_source" },
        { kind: "Source", name: "raw_data" },
      ],
    });

    const result = resolveRootCauseErrorMessage(
      model,
      [source, model],
      "fallback",
    );

    expect(result).toBe("raw_data: connection refused");
  });
});
