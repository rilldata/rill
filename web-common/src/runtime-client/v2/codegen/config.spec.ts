import { describe, it, expect } from "vitest";
import { classifyMethod } from "./config";

describe("classifyMethod", () => {
  describe("QueryService overrides", () => {
    it.each([
      ["export", "mutation"],
      ["exportReport", "mutation"],
      ["query", "mutation"],
      ["queryBatch", "skip"],
    ] as const)("classifies %s as %s", (method, expected) => {
      expect(classifyMethod("QueryService", method)).toBe(expected);
    });
  });

  describe("QueryService defaults to query", () => {
    it.each([
      "metricsViewAggregation",
      "metricsViewTimeSeries",
      "resolveCanvas",
    ])("classifies %s as query", (method) => {
      expect(classifyMethod("QueryService", method)).toBe("query");
    });
  });

  describe("RuntimeService overrides", () => {
    it.each([
      ["createInstance", "skip"],
      ["editInstance", "skip"],
      ["deleteInstance", "skip"],
      ["watchFiles", "skip"],
      ["watchLogs", "skip"],
      ["watchResources", "skip"],
      ["completeStreaming", "skip"],
      ["issueDevJWT", "query"],
      ["analyzeConnectors", "query"],
      ["analyzeVariables", "query"],
      ["queryResolver", "query"],
    ] as const)("classifies %s as %s", (method, expected) => {
      expect(classifyMethod("RuntimeService", method)).toBe(expected);
    });
  });

  describe("RuntimeService prefix-based defaults", () => {
    it.each([
      ["getFile", "query"],
      ["getExplore", "query"],
      ["listFiles", "query"],
      ["listResources", "query"],
      ["ping", "query"],
      ["health", "query"],
      ["instanceHealth", "query"],
    ] as const)("classifies %s as query (prefix match)", (method, expected) => {
      expect(classifyMethod("RuntimeService", method)).toBe(expected);
    });

    it.each([
      "putFile",
      "deleteFile",
      "generateMetricsViewFile",
      "gitCommit",
      "createTrigger",
    ])("classifies %s as mutation (no matching prefix)", (method) => {
      expect(classifyMethod("RuntimeService", method)).toBe("mutation");
    });
  });

  describe("ConnectorService defaults to query", () => {
    it.each(["listBuckets", "listTables", "getTable"])(
      "classifies %s as query",
      (method) => {
        expect(classifyMethod("ConnectorService", method)).toBe("query");
      },
    );
  });

  describe("unknown service defaults to query", () => {
    it("classifies any method on an unknown service as query", () => {
      expect(classifyMethod("UnknownService", "doSomething")).toBe("query");
    });
  });
});
