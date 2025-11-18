import { describe, it, expect } from "vitest";
import {
  traverseUpstream,
  traverseDownstream,
  traverseBidirectional,
} from "./graph-traversal";
import type { Edge } from "@xyflow/svelte";

describe("graph-traversal", () => {
  describe("traverseUpstream", () => {
    it("should traverse upstream in a simple linear chain", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["metrics1"]), edges);

      expect(result.visited).toEqual(new Set(["metrics1", "model1", "source1"]));
      expect(result.edgeIds).toEqual(new Set(["e2", "e1"]));
    });

    it("should traverse upstream with branching dependencies", () => {
      // metrics1 depends on model1 and model2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model2", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["metrics1"]), edges);

      expect(result.visited).toEqual(
        new Set(["metrics1", "model1", "model2", "source1", "source2"])
      );
      expect(result.edgeIds).toEqual(new Set(["e3", "e4", "e1", "e2"]));
    });

    it("should only include the starting node when no incoming edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseUpstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(new Set(["source1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty edge list", () => {
      const edges: Edge[] = [];

      const result = traverseUpstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle multiple starting nodes", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
      ];

      const result = traverseUpstream(new Set(["model1", "model2"]), edges);

      expect(result.visited).toEqual(
        new Set(["model1", "model2", "source1", "source2"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should not traverse downstream edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
        { id: "e3", source: "model1", target: "metrics2" },
      ];

      const result = traverseUpstream(new Set(["model1"]), edges);

      // Should only include model1 and source1, not metrics1 or metrics2
      expect(result.visited).toEqual(new Set(["model1", "source1"]));
      expect(result.edgeIds).toEqual(new Set(["e1"]));
    });

    it("should handle diamond-shaped dependency graph", () => {
      // metrics1 <- model1 <- source1
      // metrics1 <- model2 <- source1
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model2", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["metrics1"]), edges);

      expect(result.visited).toEqual(
        new Set(["metrics1", "model1", "model2", "source1"])
      );
      expect(result.edgeIds).toEqual(new Set(["e3", "e4", "e1", "e2"]));
    });

    it("should handle empty starting set", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseUpstream(new Set(), edges);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle nodes that don't exist in the graph", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseUpstream(new Set(["nonexistent"]), edges);

      expect(result.visited).toEqual(new Set(["nonexistent"]));
      expect(result.edgeIds).toEqual(new Set());
    });
  });

  describe("traverseDownstream", () => {
    it("should traverse downstream in a simple linear chain", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(new Set(["source1", "model1", "metrics1"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should traverse downstream with multiple dependents", () => {
      // source1 -> model1 -> metrics1
      // source1 -> model2 -> metrics2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model2", target: "metrics2" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "model2", "metrics1", "metrics2"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should only include the starting node when no outgoing edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseDownstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty edge list", () => {
      const edges: Edge[] = [];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(new Set(["source1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle multiple starting nodes", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
      ];

      const result = traverseDownstream(new Set(["source1", "source2"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "source2", "model1", "model2"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should not traverse upstream edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["model1"]), edges);

      // Should only include model1 and metrics1, not source1
      expect(result.visited).toEqual(new Set(["model1", "metrics1"]));
      expect(result.edgeIds).toEqual(new Set(["e2"]));
    });

    it("should handle diamond-shaped dependency graph", () => {
      // source1 -> model1 -> metrics1
      // source1 -> model2 -> metrics1
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model2", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "model2", "metrics1"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should handle empty starting set", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseDownstream(new Set(), edges);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle nodes that don't exist in the graph", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseDownstream(new Set(["nonexistent"]), edges);

      expect(result.visited).toEqual(new Set(["nonexistent"]));
      expect(result.edgeIds).toEqual(new Set());
    });
  });

  describe("traverseBidirectional", () => {
    it("should traverse both upstream and downstream", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "metrics1"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should find entire connected component", () => {
      // Complex graph:
      // source1 -> model1 -> metrics1
      // source2 -> model1
      // model1 -> metrics2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model1" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model1", target: "metrics2" },
      ];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "source2", "model1", "metrics1", "metrics2"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should handle disconnected components", () => {
      // Two separate chains:
      // source1 -> model1 -> metrics1
      // source2 -> model2 -> metrics2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
        { id: "e3", source: "source2", target: "model2" },
        { id: "e4", source: "model2", target: "metrics2" },
      ];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      // Should only include nodes connected to model1
      expect(result.visited).toEqual(new Set(["source1", "model1", "metrics1"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should handle empty edge list", () => {
      const edges: Edge[] = [];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty starting set", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
      ];

      const result = traverseBidirectional(new Set(), edges);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle multiple starting nodes in different components", () => {
      // Two separate chains
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
        { id: "e3", source: "source2", target: "model2" },
        { id: "e4", source: "model2", target: "metrics2" },
      ];

      const result = traverseBidirectional(
        new Set(["model1", "model2"]),
        edges
      );

      // Should include both connected components
      expect(result.visited).toEqual(
        new Set([
          "source1",
          "model1",
          "metrics1",
          "source2",
          "model2",
          "metrics2",
        ])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should handle complex DAG with multiple paths", () => {
      // source1 -> model1 -> model2 -> metrics1
      // source1 -> model2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "model2" },
        { id: "e4", source: "model2", target: "metrics1" },
      ];

      const result = traverseBidirectional(new Set(["model2"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "model2", "metrics1"])
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should combine upstream and downstream correctly", () => {
      const edges: Edge[] = [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "c" },
        { id: "e3", source: "c", target: "d" },
      ];

      // Start from middle node 'b'
      const result = traverseBidirectional(new Set(["b"]), edges);

      expect(result.visited).toEqual(new Set(["a", "b", "c", "d"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3"]));
    });
  });
});
