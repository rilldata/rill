import { describe, it, expect } from "vitest";
import {
  traverseUpstream,
  traverseDownstream,
  traverseBidirectional,
} from "./graph-traversal";
import type { Edge } from "@xyflow/svelte";

describe("graph-traversal", () => {
  describe("traverseUpstream", () => {
    it("should find upstream sources from a single node", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model1" },
        { id: "e3", source: "model1", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1", "source1", "source2"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should traverse multiple levels upstream", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "model2" },
        { id: "e3", source: "model2", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["metrics1"]), edges);

      expect(result.visited).toEqual(
        new Set(["metrics1", "model2", "model1", "source1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e3", "e2", "e1"]));
    });

    it("should handle multiple starting nodes", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
      ];

      const result = traverseUpstream(new Set(["model1", "model2"]), edges);

      expect(result.visited).toEqual(
        new Set(["model1", "model2", "source1", "source2"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should handle diamond dependencies", () => {
      // source1 -> model1 -> model3
      //         -> model2 -> model3
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "model3" },
        { id: "e4", source: "model2", target: "model3" },
      ];

      const result = traverseUpstream(new Set(["model3"]), edges);

      expect(result.visited).toEqual(
        new Set(["model3", "model1", "model2", "source1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e3", "e4", "e1", "e2"]));
    });

    it("should return only selected node when it has no upstream dependencies", () => {
      const edges: Edge[] = [{ id: "e1", source: "source1", target: "model1" }];

      const result = traverseUpstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(new Set(["source1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty edge list", () => {
      const result = traverseUpstream(new Set(["model1"]), []);

      expect(result.visited).toEqual(new Set(["model1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty selected nodes", () => {
      const edges: Edge[] = [{ id: "e1", source: "source1", target: "model1" }];

      const result = traverseUpstream(new Set([]), edges);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should not traverse downstream edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseUpstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1", "source1"]));
      expect(result.edgeIds).toEqual(new Set(["e1"]));
      expect(result.visited.has("metrics1")).toBe(false);
    });

    it("should handle cycles gracefully without infinite loop", () => {
      // This shouldn't happen in a DAG, but test defensive coding
      const edges: Edge[] = [
        { id: "e1", source: "node1", target: "node2" },
        { id: "e2", source: "node2", target: "node3" },
        { id: "e3", source: "node3", target: "node1" }, // Creates cycle
      ];

      const result = traverseUpstream(new Set(["node1"]), edges);

      expect(result.visited).toEqual(new Set(["node1", "node3", "node2"]));
      expect(result.edgeIds).toEqual(new Set(["e3", "e2", "e1"]));
    });
  });

  describe("traverseDownstream", () => {
    it("should find downstream dependents from a single node", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source1", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      // Should traverse all downstream nodes recursively
      expect(result.visited).toEqual(
        new Set(["source1", "model1", "model2", "metrics1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3"]));
    });

    it("should traverse multiple levels downstream", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "model2" },
        { id: "e3", source: "model2", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "model2", "metrics1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3"]));
    });

    it("should handle multiple starting nodes", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
      ];

      const result = traverseDownstream(new Set(["source1", "source2"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "source2", "model1", "model2"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should handle fan-out dependencies", () => {
      // source1 -> model1 -> metrics1
      //                   -> metrics2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
        { id: "e3", source: "model1", target: "metrics2" },
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "metrics1", "metrics2"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3"]));
    });

    it("should return only selected node when it has no downstream dependents", () => {
      const edges: Edge[] = [{ id: "e1", source: "source1", target: "model1" }];

      const result = traverseDownstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty edge list", () => {
      const result = traverseDownstream(new Set(["source1"]), []);

      expect(result.visited).toEqual(new Set(["source1"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty selected nodes", () => {
      const edges: Edge[] = [{ id: "e1", source: "source1", target: "model1" }];

      const result = traverseDownstream(new Set([]), edges);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should not traverse upstream edges", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseDownstream(new Set(["model1"]), edges);

      expect(result.visited).toEqual(new Set(["model1", "metrics1"]));
      expect(result.edgeIds).toEqual(new Set(["e2"]));
      expect(result.visited.has("source1")).toBe(false);
    });

    it("should handle cycles gracefully without infinite loop", () => {
      const edges: Edge[] = [
        { id: "e1", source: "node1", target: "node2" },
        { id: "e2", source: "node2", target: "node3" },
        { id: "e3", source: "node3", target: "node1" },
      ];

      const result = traverseDownstream(new Set(["node1"]), edges);

      expect(result.visited).toEqual(new Set(["node1", "node2", "node3"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3"]));
    });
  });

  describe("traverseBidirectional", () => {
    it("should find all connected nodes in both directions", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "metrics1" },
      ];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      expect(result.visited).toEqual(
        new Set(["model1", "source1", "metrics1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2"]));
    });

    it("should combine upstream and downstream for complex graph", () => {
      // source1 -> model1 -> model2 -> metrics1
      //                   -> model3 -> metrics2
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "model1", target: "model2" },
        { id: "e3", source: "model1", target: "model3" },
        { id: "e4", source: "model2", target: "metrics1" },
        { id: "e5", source: "model3", target: "metrics2" },
      ];

      const result = traverseBidirectional(new Set(["model1"]), edges);

      expect(result.visited).toEqual(
        new Set([
          "source1",
          "model1",
          "model2",
          "model3",
          "metrics1",
          "metrics2",
        ]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4", "e5"]));
    });

    it("should handle isolated node", () => {
      const edges: Edge[] = [{ id: "e1", source: "source1", target: "model1" }];

      const result = traverseBidirectional(new Set(["isolated"]), edges);

      expect(result.visited).toEqual(new Set(["isolated"]));
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should handle empty inputs", () => {
      const result = traverseBidirectional(new Set([]), []);

      expect(result.visited).toEqual(new Set());
      expect(result.edgeIds).toEqual(new Set());
    });

    it("should find entire connected component", () => {
      const edges: Edge[] = [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "c" },
        { id: "e3", source: "c", target: "d" },
        { id: "e4", source: "d", target: "e" },
      ];

      // Starting from middle node should find all connected nodes
      const result = traverseBidirectional(new Set(["c"]), edges);

      expect(result.visited).toEqual(new Set(["a", "b", "c", "d", "e"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });

    it("should handle multiple starting nodes in connected component", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { id: "e2", source: "source2", target: "model2" },
        { id: "e3", source: "model1", target: "metrics1" },
        { id: "e4", source: "model2", target: "metrics1" },
      ];

      const result = traverseBidirectional(
        new Set(["model1", "model2"]),
        edges,
      );

      expect(result.visited).toEqual(
        new Set(["source1", "source2", "model1", "model2", "metrics1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });
  });

  describe("performance and edge cases", () => {
    it("should handle large graphs efficiently", () => {
      // Create a chain of 1000 nodes
      const edges: Edge[] = [];
      for (let i = 0; i < 999; i++) {
        edges.push({
          id: `e${i}`,
          source: `node${i}`,
          target: `node${i + 1}`,
        });
      }

      const startTime = performance.now();
      const result = traverseDownstream(new Set(["node0"]), edges);
      const endTime = performance.now();

      expect(result.visited.size).toBe(1000);
      expect(result.edgeIds.size).toBe(999);
      // Should complete in reasonable time (< 100ms for 1000 nodes)
      expect(endTime - startTime).toBeLessThan(100);
    });

    it("should handle edges with missing IDs", () => {
      const edges: Edge[] = [
        { id: "e1", source: "source1", target: "model1" },
        { source: "model1", target: "metrics1" } as Edge, // Missing id
      ];

      const result = traverseDownstream(new Set(["source1"]), edges);

      expect(result.visited).toEqual(
        new Set(["source1", "model1", "metrics1"]),
      );
      expect(result.edgeIds).toEqual(new Set(["e1", undefined]));
    });

    it("should deduplicate nodes visited from multiple paths", () => {
      // Create multiple paths to same node
      const edges: Edge[] = [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "a", target: "c" },
        { id: "e3", source: "b", target: "d" },
        { id: "e4", source: "c", target: "d" },
      ];

      const result = traverseDownstream(new Set(["a"]), edges);

      // Node 'd' is reachable via two paths but should only appear once
      expect(result.visited).toEqual(new Set(["a", "b", "c", "d"]));
      expect(result.edgeIds).toEqual(new Set(["e1", "e2", "e3", "e4"]));
    });
  });
});
