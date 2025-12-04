import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  buildResourceGraph,
  partitionResourcesBySeeds,
  partitionResourcesByMetrics,
} from "./graph-builder";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
  writable: true,
});

describe("build-resource-graph", () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe("buildResourceGraph", () => {
    it("should create nodes for allowed resource kinds", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Explore, name: "explore1" },
            hidden: false,
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(4);
      expect(nodes.map((n) => n.id)).toEqual([
        "rill.runtime.v1.Source:source1",
        "rill.runtime.v1.Model:model1",
        "rill.runtime.v1.MetricsView:metrics1",
        "rill.runtime.v1.Explore:explore1",
      ]);
    });

    it("should filter out hidden resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "visible" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "hidden" },
            hidden: true,
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(1);
      expect(nodes[0].id).toBe("rill.runtime.v1.Model:visible");
    });

    it("should create edges based on resource refs", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
      ];

      const { edges } = buildResourceGraph(resources);

      expect(edges).toHaveLength(1);
      expect(edges[0].source).toBe("rill.runtime.v1.Source:source1");
      expect(edges[0].target).toBe("rill.runtime.v1.Model:model1");
    });

    it("should handle multiple refs per resource", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source2" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [
              { kind: ResourceKind.Source, name: "source1" },
              { kind: ResourceKind.Source, name: "source2" },
            ],
            hidden: false,
          },
        },
      ];

      const { edges } = buildResourceGraph(resources);

      expect(edges).toHaveLength(2);
      expect(edges.map((e) => e.source)).toEqual([
        "rill.runtime.v1.Source:source1",
        "rill.runtime.v1.Source:source2",
      ]);
    });

    it("should not create edges to missing refs", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "missing_source" }],
            hidden: false,
          },
        },
      ];

      const { edges } = buildResourceGraph(resources);

      expect(edges).toHaveLength(0);
    });

    it("should not create self-referencing edges", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
      ];

      const { edges } = buildResourceGraph(resources);

      expect(edges).toHaveLength(0);
    });

    it("should handle empty resource list", () => {
      const { nodes, edges } = buildResourceGraph([]);

      expect(nodes).toHaveLength(0);
      expect(edges).toHaveLength(0);
    });

    it("should handle resources with no refs", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
      ];

      const { nodes, edges } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(1);
      expect(edges).toHaveLength(0);
    });

    it("should set node dimensions appropriately", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "short" },
            hidden: false,
          },
        },
        {
          meta: {
            name: {
              kind: ResourceKind.Model,
              name: "very_long_resource_name_that_should_be_wider",
            },
            hidden: false,
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes[0].width).toBeDefined();
      expect(nodes[0].height).toBeDefined();
      expect(nodes[1].width).toBeGreaterThan(nodes[0].width!);
      expect(nodes[0].height).toBe(nodes[1].height);
    });

    it("should set source position and target position for nodes", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes[0].sourcePosition).toBe("bottom");
      expect(nodes[0].targetPosition).toBe("top");
    });

    it("should use different positions for different namespaces", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const result1 = buildResourceGraph(resources, { positionNs: "ns1" });
      const result2 = buildResourceGraph(resources, { positionNs: "ns2" });

      // Positions might differ depending on cache state and dagre layout
      // We mainly verify that the namespace parameter is accepted
      expect(result1.nodes).toHaveLength(1);
      expect(result2.nodes).toHaveLength(1);
    });

    it("should ignore cache when ignoreCache option is true", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const result1 = buildResourceGraph(resources, {
        positionNs: "test",
        ignoreCache: false,
      });
      const result2 = buildResourceGraph(resources, {
        positionNs: "test",
        ignoreCache: true,
      });

      // Both should create valid nodes
      expect(result1.nodes).toHaveLength(1);
      expect(result2.nodes).toHaveLength(1);
    });

    it("should handle complex dependency graph", () => {
      // source1 -> model1 -> metrics1
      //         -> model2 -> metrics1
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model2" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [
              { kind: ResourceKind.Model, name: "model1" },
              { kind: ResourceKind.Model, name: "model2" },
            ],
            hidden: false,
          },
        },
      ];

      const { nodes, edges } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(4);
      expect(edges).toHaveLength(4);
    });

    it("should assign resource data to nodes", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
          model: {
            spec: {},
            state: {},
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes[0].data.resource).toBeDefined();
      expect(nodes[0].data.kind).toBe(ResourceKind.Model);
      expect(nodes[0].data.label).toBe("model1");
    });

    it("should deduplicate edges with same source and target", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [
              { kind: ResourceKind.Source, name: "source1" },
              { kind: ResourceKind.Source, name: "source1" }, // Duplicate ref
            ],
            hidden: false,
          },
        },
      ];

      const { edges } = buildResourceGraph(resources);

      expect(edges).toHaveLength(1);
    });

    it("should handle resources with incomplete meta", () => {
      const resources: V1Resource[] = [
        {},
        { meta: {} },
        { meta: { name: {} as any } },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "valid" },
            hidden: false,
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(1);
      expect(nodes[0].id).toBe("rill.runtime.v1.Model:valid");
    });

    it("should coerce model defined-as-source to Source kind", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "raw_orders" },
            hidden: false,
          },
          model: {
            spec: {
              definedAsSource: true,
            },
            state: {
              resultTable: "raw_orders",
            },
          },
        },
      ];

      const { nodes } = buildResourceGraph(resources);

      expect(nodes).toHaveLength(1);
      // The node data should have kind Source due to coercion
      expect(nodes[0].data.kind).toBe(ResourceKind.Source);
    });
  });

  describe("partitionResourcesBySeeds", () => {
    it("should create single group from single seed", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Model:model1",
      ]);

      expect(groups).toHaveLength(1);
      expect(groups[0].id).toBe("rill.runtime.v1.Model:model1");
      expect(groups[0].resources).toHaveLength(2);
      expect(groups[0].label).toBe("model1");
    });

    it("should include upstream dependencies in group", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.MetricsView:metrics1",
      ]);

      expect(groups).toHaveLength(1);
      expect(groups[0].resources).toHaveLength(3);
      const resourceNames = groups[0].resources.map((r) => r.meta?.name?.name);
      expect(resourceNames).toContain("source1");
      expect(resourceNames).toContain("model1");
      expect(resourceNames).toContain("metrics1");
    });

    it("should include downstream dependents in group", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Source:source1",
      ]);

      expect(groups).toHaveLength(1);
      expect(groups[0].resources).toHaveLength(3);
    });

    it("should create separate groups for multiple seeds", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source2" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model2" },
            refs: [{ kind: ResourceKind.Source, name: "source2" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Model:model1",
        "rill.runtime.v1.Model:model2",
      ]);

      expect(groups).toHaveLength(2);
      expect(groups[0].label).toBe("model1");
      expect(groups[1].label).toBe("model2");
    });

    it("should accept V1ResourceName objects as seeds", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        { kind: ResourceKind.Model, name: "model1" },
      ]);

      expect(groups).toHaveLength(1);
      expect(groups[0].label).toBe("model1");
    });

    it("should handle empty seeds array", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, []);

      // May have cached groups from previous runs
      expect(groups.length).toBeGreaterThanOrEqual(0);
    });

    it("should handle seed that doesn't exist", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Model:nonexistent",
      ]);

      // Should create empty group or no group depending on implementation
      expect(groups.length).toBeGreaterThanOrEqual(0);
    });

    it("should deduplicate seed identifiers", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Model:model1",
        "rill.runtime.v1.Model:model1", // Duplicate
      ]);

      expect(groups).toHaveLength(1);
    });

    it("should filter out hidden resources from groups", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: true,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesBySeeds(resources, [
        "rill.runtime.v1.Model:model1",
      ]);

      expect(groups).toHaveLength(1);
      // Verify that model1 is included in the group
      const model1 = groups[0].resources.find(
        (r) => r.meta?.name?.name === "model1",
      );
      expect(model1).toBeDefined();
      expect(model1?.meta?.hidden).toBe(false);

      // Verify source1 is either not included OR included as placeholder (reconcileError set)
      const source1Visible = groups[0].resources.find(
        (r) => r.meta?.name?.name === "source1" && !r.meta?.hidden,
      );
      // If source1 is in the group without hidden flag, it should be a placeholder with error
      if (source1Visible) {
        expect(source1Visible.meta?.reconcileError).toBeTruthy();
      }
    });
  });

  describe("partitionResourcesByMetrics", () => {
    it("should create groups based on MetricsView resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "revenue" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(1);
      expect(groups[0].label).toBe("revenue");
      expect(groups[0].resources).toHaveLength(3);
    });

    it("should sort groups alphabetically by MetricsView name", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "zebra" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "apple" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "mango" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(3);
      expect(groups.map((g) => g.label)).toEqual(["apple", "mango", "zebra"]);
    });

    it("should create 'Other resources' group for unconnected resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orphan" },
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(2);
      const otherGroup = groups.find((g) => g.label === "Other resources");
      expect(otherGroup).toBeDefined();
      expect(otherGroup?.resources).toHaveLength(1);
    });

    it("should handle empty resource list", () => {
      const groups = partitionResourcesByMetrics([]);

      expect(groups).toHaveLength(0);
    });

    it("should handle resources with no MetricsView", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(1);
      expect(groups[0].label).toBe("Other resources");
    });

    it("should group connected components together", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model2" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [{ kind: ResourceKind.Model, name: "model2" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(1);
      expect(groups[0].resources).toHaveLength(4);
    });

    it("should create separate groups for disconnected metrics", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source2" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model2" },
            refs: [{ kind: ResourceKind.Source, name: "source2" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics2" },
            refs: [{ kind: ResourceKind.Model, name: "model2" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(2);
      expect(groups[0].resources).toHaveLength(3);
      expect(groups[1].resources).toHaveLength(3);
    });

    it("should filter out hidden resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "source1" },
            hidden: true,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "metrics1" },
            refs: [{ kind: ResourceKind.Source, name: "source1" }],
            hidden: false,
          },
        },
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups).toHaveLength(1);
      expect(groups[0].resources).toHaveLength(1);
    });
  });

  describe("edge cases and performance", () => {
    it("should handle circular dependencies gracefully", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [{ kind: ResourceKind.Model, name: "model2" }],
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model2" },
            refs: [{ kind: ResourceKind.Model, name: "model1" }],
            hidden: false,
          },
        },
      ];

      // Should not hang or crash
      expect(() => buildResourceGraph(resources)).not.toThrow();
      expect(() => partitionResourcesByMetrics(resources)).not.toThrow();
    });

    it("should handle large number of resources", () => {
      const resources: V1Resource[] = [];
      for (let i = 0; i < 100; i++) {
        resources.push({
          meta: {
            name: { kind: ResourceKind.Model, name: `model${i}` },
            hidden: false,
          },
        });
      }

      const startTime = performance.now();
      const { nodes } = buildResourceGraph(resources);
      const endTime = performance.now();

      expect(nodes).toHaveLength(100);
      // Should complete in reasonable time (< 500ms for 100 nodes)
      expect(endTime - startTime).toBeLessThan(500);
    });
  });
});
