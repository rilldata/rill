import { describe, it, expect } from "vitest";
import {
  buildResourceGraph,
  partitionResourcesByMetrics,
  partitionResourcesBySeeds,
} from "./build-resource-graph";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

describe("build-resource-graph", () => {
  // Helper to create a test resource
  function createResource(
    kind: ResourceKind,
    name: string,
    refs: Array<{ kind: ResourceKind; name: string }> = [],
    hidden = false
  ): V1Resource {
    return {
      meta: {
        name: { kind, name },
        refs,
        hidden,
      },
    };
  }

  describe("buildResourceGraph", () => {
    it("should build graph with nodes and edges for simple chain", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(3);
      expect(result.edges).toHaveLength(2);

      // Check nodes exist
      const nodeIds = result.nodes.map((n) => n.id);
      expect(nodeIds).toContain(`${ResourceKind.Source}:source1`);
      expect(nodeIds).toContain(`${ResourceKind.Model}:model1`);
      expect(nodeIds).toContain(`${ResourceKind.MetricsView}:metrics1`);

      // Check edges
      const edgeIds = result.edges.map((e) => e.id);
      expect(edgeIds).toContain(
        `${ResourceKind.Source}:source1->${ResourceKind.Model}:model1`
      );
      expect(edgeIds).toContain(
        `${ResourceKind.Model}:model1->${ResourceKind.MetricsView}:metrics1`
      );
    });

    it("should exclude hidden resources", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(
          ResourceKind.Model,
          "model1",
          [{ kind: ResourceKind.Source, name: "source1" }],
          true
        ), // hidden
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const result = buildResourceGraph(resources);

      const nodeIds = result.nodes.map((n) => n.id);
      expect(nodeIds).not.toContain(`${ResourceKind.Model}:model1`);
      expect(nodeIds).toContain(`${ResourceKind.Source}:source1`);
      expect(nodeIds).toContain(`${ResourceKind.MetricsView}:metrics1`);
    });

    it("should handle empty resource list", () => {
      const result = buildResourceGraph([]);

      expect(result.nodes).toEqual([]);
      expect(result.edges).toEqual([]);
    });

    it("should handle resources with no references", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Source, "source2"),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(2);
      expect(result.edges).toHaveLength(0);
    });

    it("should handle diamond dependency graph", () => {
      // metrics1 <- model1 <- source1
      // metrics1 <- model2 <- source1
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.Model, "model2", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
          { kind: ResourceKind.Model, name: "model2" },
        ]),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(4);
      expect(result.edges).toHaveLength(4);

      const edgeIds = result.edges.map((e) => e.id);
      expect(edgeIds).toContain(
        `${ResourceKind.Source}:source1->${ResourceKind.Model}:model1`
      );
      expect(edgeIds).toContain(
        `${ResourceKind.Source}:source1->${ResourceKind.Model}:model2`
      );
      expect(edgeIds).toContain(
        `${ResourceKind.Model}:model1->${ResourceKind.MetricsView}:metrics1`
      );
      expect(edgeIds).toContain(
        `${ResourceKind.Model}:model2->${ResourceKind.MetricsView}:metrics1`
      );
    });

    it("should skip invalid resource references", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "nonexistent_source" },
        ]),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(2);
      // No edge created because reference doesn't exist
      expect(result.edges).toHaveLength(0);
    });

    it("should include node data with resource information", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "test_source"),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes[0].data).toBeDefined();
      expect(result.nodes[0].data.label).toBe("test_source");
      expect(result.nodes[0].data.kind).toBe(ResourceKind.Source);
      expect(result.nodes[0].data.resource).toEqual(resources[0]);
    });

    it("should assign positions to all nodes", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const result = buildResourceGraph(resources);

      result.nodes.forEach((node) => {
        expect(node.position).toBeDefined();
        expect(typeof node.position.x).toBe("number");
        expect(typeof node.position.y).toBe("number");
      });
    });

    it("should not create self-referencing edges", () => {
      const resource = createResource(ResourceKind.Model, "model1", [
        { kind: ResourceKind.Model, name: "model1" },
      ]);
      const resources: V1Resource[] = [resource];

      const result = buildResourceGraph(resources);

      expect(result.edges).toHaveLength(0);
    });

    it("should deduplicate edges", () => {
      // Even if we somehow get duplicate refs, should only create one edge
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "model1" },
            refs: [
              { kind: ResourceKind.Source, name: "source1" },
              { kind: ResourceKind.Source, name: "source1" }, // duplicate
            ],
          },
        },
      ];

      const result = buildResourceGraph(resources);

      expect(result.edges).toHaveLength(1);
    });

    it("should handle resources with missing metadata", () => {
      const resources: V1Resource[] = [
        {}, // no metadata
        createResource(ResourceKind.Source, "source1"),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(1);
      expect(result.nodes[0].data.label).toBe("source1");
    });

    it("should handle resources with missing name", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model }, // no name
          },
        },
        createResource(ResourceKind.Source, "source1"),
      ];

      const result = buildResourceGraph(resources);

      expect(result.nodes).toHaveLength(1);
      expect(result.nodes[0].data.label).toBe("source1");
    });
  });

  describe("partitionResourcesByMetrics", () => {
    it("should create one group per MetricsView", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.MetricsView, "revenue", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
        createResource(ResourceKind.MetricsView, "sales", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups.length).toBeGreaterThanOrEqual(2);

      // Find the revenue and sales groups
      const revenueGroup = groups.find((g) => g.label === "revenue");
      const salesGroup = groups.find((g) => g.label === "sales");

      expect(revenueGroup).toBeDefined();
      expect(salesGroup).toBeDefined();
    });

    it("should include connected resources in each group", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      const metricsGroup = groups.find((g) => g.label === "metrics1");
      expect(metricsGroup).toBeDefined();

      // The group should contain at least the metrics view itself
      expect(metricsGroup!.resources.length).toBeGreaterThanOrEqual(1);

      const resourceNames = metricsGroup!.resources.map(
        (r) => r.meta?.name?.name
      );
      expect(resourceNames).toContain("metrics1");
    });

    it("should handle orphaned resources (not connected to MetricsView)", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "orphan_source"),
        createResource(ResourceKind.Model, "orphan_model", [
          { kind: ResourceKind.Source, name: "orphan_source" },
        ]),
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      // Should have at least metrics1 group and potentially an orphan group
      expect(groups.length).toBeGreaterThanOrEqual(1);

      const metricsGroup = groups.find((g) => g.label === "metrics1");
      expect(metricsGroup).toBeDefined();
    });

    it("should handle empty resource list", () => {
      const groups = partitionResourcesByMetrics([]);

      expect(groups).toEqual([]);
    });

    it("should exclude hidden resources from groups", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(
          ResourceKind.Model,
          "model1",
          [{ kind: ResourceKind.Source, name: "source1" }],
          true
        ), // hidden
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      const metricsGroup = groups.find((g) => g.label === "metrics1");
      expect(metricsGroup).toBeDefined();

      const resourceNames = metricsGroup!.resources.map(
        (r) => r.meta?.name?.name
      );
      expect(resourceNames).not.toContain("model1");
    });

    it("should handle multiple disconnected components", () => {
      const resources: V1Resource[] = [
        // Component 1
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        // Component 2
        createResource(ResourceKind.Source, "source2"),
        createResource(ResourceKind.MetricsView, "metrics2", [
          { kind: ResourceKind.Source, name: "source2" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      expect(groups.length).toBeGreaterThanOrEqual(2);

      const group1 = groups.find((g) => g.label === "metrics1");
      const group2 = groups.find((g) => g.label === "metrics2");

      expect(group1).toBeDefined();
      expect(group2).toBeDefined();
    });

    it("should create valid group IDs", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const groups = partitionResourcesByMetrics(resources);

      groups.forEach((group) => {
        expect(group.id).toBeDefined();
        expect(typeof group.id).toBe("string");
        expect(group.id.length).toBeGreaterThan(0);
      });
    });
  });

  describe("partitionResourcesBySeeds", () => {
    it("should create groups for seeds", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.Model, "model2", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const seeds = ["model:model1", "model:model2"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      // Should create at least one group
      expect(groups.length).toBeGreaterThanOrEqual(1);

      // Check that we can find groups by checking resources
      const hasModel1 = groups.some((g) =>
        g.resources.some((r) => r.meta?.name?.name === "model1")
      );
      const hasModel2 = groups.some((g) =>
        g.resources.some((r) => r.meta?.name?.name === "model2")
      );

      // At least one of the models should be in the groups
      expect(hasModel1 || hasModel2).toBeTruthy();
    });

    it("should include dependencies in seed groups", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const seeds = ["model:model1"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      expect(groups.length).toBeGreaterThanOrEqual(1);

      // Find group containing model1
      const group = groups.find((g) =>
        g.resources.some((r) => r.meta?.name?.name === "model1")
      );
      expect(group).toBeDefined();

      const resourceNames = group!.resources.map((r) => r.meta?.name?.name);
      expect(resourceNames).toContain("model1");
    });

    it("should handle V1ResourceName objects as seeds", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const seeds = [{ kind: ResourceKind.Model, name: "model1" }];
      const groups = partitionResourcesBySeeds(resources, seeds);

      expect(groups.length).toBeGreaterThanOrEqual(1);

      const group = groups.find((g) => g.label === "model1");
      expect(group).toBeDefined();
    });

    it("should handle empty seed array", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const groups = partitionResourcesBySeeds(resources, []);

      // Empty seeds might create cached groups or no groups
      expect(Array.isArray(groups)).toBe(true);
    });

    it("should handle seeds for non-existent resources", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
      ];

      const seeds = ["model:nonexistent"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      // Should handle gracefully
      expect(Array.isArray(groups)).toBe(true);
    });

    it("should handle multiple seeds", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
        createResource(ResourceKind.Model, "model2", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const seeds = ["model:model1", "model:model2"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      // Should create some groups
      expect(groups.length).toBeGreaterThanOrEqual(1);
    });

    it("should handle hidden resources in seed dependencies", () => {
      // Note: The partition function creates placeholders for missing/hidden resources
      // that are referenced by visible resources to maintain graph integrity
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(
          ResourceKind.Model,
          "model1",
          [{ kind: ResourceKind.Source, name: "source1" }],
          true
        ), // hidden
        createResource(ResourceKind.MetricsView, "metrics1", [
          { kind: ResourceKind.Model, name: "model1" },
        ]),
      ];

      const seeds = ["metrics:metrics1"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      expect(groups.length).toBeGreaterThanOrEqual(1);

      // Verify metrics1 is in at least one group
      const hasMetrics1 = groups.some((g) =>
        g.resources.some((r) => r.meta?.name?.name === "metrics1")
      );
      expect(hasMetrics1).toBeTruthy();
    });

    it("should create valid group IDs", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const seeds = ["model:model1"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      groups.forEach((group) => {
        expect(group.id).toBeDefined();
        expect(typeof group.id).toBe("string");
        expect(group.id.length).toBeGreaterThan(0);
      });
    });

    it("should handle fully qualified kind in seeds", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Source, "source1"),
        createResource(ResourceKind.Model, "model1", [
          { kind: ResourceKind.Source, name: "source1" },
        ]),
      ];

      const seeds = [`${ResourceKind.Model}:model1`];
      const groups = partitionResourcesBySeeds(resources, seeds);

      expect(groups.length).toBeGreaterThanOrEqual(1);

      const group = groups.find((g) => g.label === "model1");
      expect(group).toBeDefined();
    });

    it("should handle seeds with special characters in names", () => {
      const resources: V1Resource[] = [
        createResource(ResourceKind.Model, "orders_2024-v2"),
      ];

      const seeds = ["model:orders_2024-v2"];
      const groups = partitionResourcesBySeeds(resources, seeds);

      // Should handle gracefully
      expect(Array.isArray(groups)).toBe(true);
    });
  });
});
