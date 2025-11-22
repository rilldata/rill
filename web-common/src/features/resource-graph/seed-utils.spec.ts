import { describe, it, expect } from "vitest";
import {
  normalizeSeed,
  isKindToken,
  tokenForKind,
  tokenForSeedString,
  expandSeedsByKind,
  ALLOWED_FOR_GRAPH,
} from "./seed-utils";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

describe("seed-utils", () => {
  describe("normalizeSeed", () => {
    it("should default name-only seeds to MetricsView", () => {
      const result = normalizeSeed("orders");
      expect(result).toEqual({
        kind: ResourceKind.MetricsView,
        name: "orders",
      });
    });

    it("should parse kind:name format with singular kind", () => {
      const result = normalizeSeed("model:clean_orders");
      expect(result).toEqual({
        kind: ResourceKind.Model,
        name: "clean_orders",
      });
    });

    it("should parse kind:name format with plural kind", () => {
      const result = normalizeSeed("models:clean_orders");
      expect(result).toEqual({
        kind: ResourceKind.Model,
        name: "clean_orders",
      });
    });

    it("should handle plural forms - sources", () => {
      const result = normalizeSeed("sources:raw_data");
      expect(result).toEqual({
        kind: ResourceKind.Source,
        name: "raw_data",
      });
    });

    it("should handle plural forms - metrics", () => {
      const result = normalizeSeed("metrics:revenue");
      expect(result).toEqual({
        kind: ResourceKind.MetricsView,
        name: "revenue",
      });
    });

    it("should handle plural forms - dashboards", () => {
      const result = normalizeSeed("dashboards:sales");
      expect(result).toEqual({
        kind: ResourceKind.Explore,
        name: "sales",
      });
    });

    it("should handle singular alias - metric", () => {
      const result = normalizeSeed("metric:revenue");
      expect(result).toEqual({
        kind: ResourceKind.MetricsView,
        name: "revenue",
      });
    });

    it("should handle singular alias - dashboard", () => {
      const result = normalizeSeed("dashboard:sales");
      expect(result).toEqual({
        kind: ResourceKind.Explore,
        name: "sales",
      });
    });

    it("should handle fully qualified kind", () => {
      const result = normalizeSeed("rill.runtime.v1.Model:orders");
      expect(result).toEqual({
        kind: "rill.runtime.v1.Model",
        name: "orders",
      });
    });

    it("should preserve case in names", () => {
      const result = normalizeSeed("model:OrdersByRegion");
      expect(result).toEqual({
        kind: ResourceKind.Model,
        name: "OrdersByRegion",
      });
    });

    it("should handle names with special characters", () => {
      const result = normalizeSeed("model:orders_2024-v2");
      expect(result).toEqual({
        kind: ResourceKind.Model,
        name: "orders_2024-v2",
      });
    });

    it("should handle names with colons", () => {
      const result = normalizeSeed("model:table:column");
      expect(result).toEqual({
        kind: ResourceKind.Model,
        name: "table:column",
      });
    });

    it("should return as-is for unknown kind aliases", () => {
      const result = normalizeSeed("unknown:test");
      expect(result).toBe("unknown:test");
    });

    it("should handle metricsview as alias", () => {
      const result = normalizeSeed("metricsview:revenue");
      expect(result).toEqual({
        kind: ResourceKind.MetricsView,
        name: "revenue",
      });
    });

    it("should handle explore as alias", () => {
      const result = normalizeSeed("explore:dashboard");
      expect(result).toEqual({
        kind: ResourceKind.Explore,
        name: "dashboard",
      });
    });
  });

  describe("isKindToken", () => {
    it("should recognize 'metrics' as MetricsView token", () => {
      expect(isKindToken("metrics")).toBe(ResourceKind.MetricsView);
    });

    it("should recognize 'models' as Model token", () => {
      expect(isKindToken("models")).toBe(ResourceKind.Model);
    });

    it("should recognize 'sources' as Source token", () => {
      expect(isKindToken("sources")).toBe(ResourceKind.Source);
    });

    it("should recognize 'dashboards' as Explore token", () => {
      expect(isKindToken("dashboards")).toBe(ResourceKind.Explore);
    });

    it("should recognize singular 'model' as Model token", () => {
      expect(isKindToken("model")).toBe(ResourceKind.Model);
    });

    it("should recognize singular 'source' as Source token", () => {
      expect(isKindToken("source")).toBe(ResourceKind.Source);
    });

    it("should recognize singular 'metric' as MetricsView token", () => {
      expect(isKindToken("metric")).toBe(ResourceKind.MetricsView);
    });

    it("should recognize singular 'dashboard' as Explore token", () => {
      expect(isKindToken("dashboard")).toBe(ResourceKind.Explore);
    });

    it("should recognize 'metricsview' as MetricsView token", () => {
      expect(isKindToken("metricsview")).toBe(ResourceKind.MetricsView);
    });

    it("should recognize 'explore' as Explore token", () => {
      expect(isKindToken("explore")).toBe(ResourceKind.Explore);
    });

    it("should return undefined for non-kind tokens", () => {
      expect(isKindToken("orders")).toBeUndefined();
    });

    it("should return undefined for empty string", () => {
      expect(isKindToken("")).toBeUndefined();
    });

    it("should be case-insensitive", () => {
      expect(isKindToken("MODELS")).toBe(ResourceKind.Model);
      expect(isKindToken("Models")).toBe(ResourceKind.Model);
    });
  });

  describe("tokenForKind", () => {
    it("should return 'sources' for Source kind", () => {
      expect(tokenForKind(ResourceKind.Source)).toBe("sources");
    });

    it("should return 'models' for Model kind", () => {
      expect(tokenForKind(ResourceKind.Model)).toBe("models");
    });

    it("should return 'metrics' for MetricsView kind", () => {
      expect(tokenForKind(ResourceKind.MetricsView)).toBe("metrics");
    });

    it("should return 'dashboards' for Explore kind", () => {
      expect(tokenForKind(ResourceKind.Explore)).toBe("dashboards");
    });

    it("should handle fully qualified kind strings", () => {
      expect(tokenForKind("rill.runtime.v1.Source")).toBe("sources");
      expect(tokenForKind("rill.runtime.v1.Model")).toBe("models");
      expect(tokenForKind("rill.runtime.v1.MetricsView")).toBe("metrics");
      expect(tokenForKind("rill.runtime.v1.Explore")).toBe("dashboards");
    });

    it("should return null for undefined kind", () => {
      expect(tokenForKind(undefined)).toBeNull();
    });

    it("should return null for null kind", () => {
      expect(tokenForKind(null)).toBeNull();
    });

    it("should return null for empty string", () => {
      expect(tokenForKind("")).toBeNull();
    });

    it("should return null for unknown kind", () => {
      expect(tokenForKind("unknown")).toBeNull();
    });

    it("should be case-insensitive", () => {
      expect(tokenForKind("SOURCE")).toBe("sources");
      expect(tokenForKind("Model")).toBe("models");
    });
  });

  describe("tokenForSeedString", () => {
    it("should return 'models' for model seed", () => {
      expect(tokenForSeedString("model:orders")).toBe("models");
    });

    it("should return 'sources' for source seed", () => {
      expect(tokenForSeedString("source:raw_data")).toBe("sources");
    });

    it("should return 'metrics' for metrics seed", () => {
      expect(tokenForSeedString("metrics:revenue")).toBe("metrics");
    });

    it("should return 'dashboards' for dashboard seed", () => {
      expect(tokenForSeedString("dashboard:sales")).toBe("dashboards");
    });

    it("should return 'metrics' for kind token 'metrics'", () => {
      expect(tokenForSeedString("metrics")).toBe("metrics");
    });

    it("should return 'models' for kind token 'models'", () => {
      expect(tokenForSeedString("models")).toBe("models");
    });

    it("should return 'metrics' for name-only seed (defaults to metrics)", () => {
      expect(tokenForSeedString("orders")).toBe("metrics");
    });

    it("should handle plural forms in seeds", () => {
      expect(tokenForSeedString("models:clean_orders")).toBe("models");
      expect(tokenForSeedString("sources:raw_data")).toBe("sources");
    });

    it("should return null for undefined seed", () => {
      expect(tokenForSeedString(undefined)).toBeNull();
    });

    it("should return null for null seed", () => {
      expect(tokenForSeedString(null)).toBeNull();
    });

    it("should return null for empty string", () => {
      expect(tokenForSeedString("")).toBeNull();
    });

    it("should return null for whitespace-only string", () => {
      expect(tokenForSeedString("   ")).toBeNull();
    });

    it("should handle fully qualified kinds in seeds", () => {
      expect(tokenForSeedString("rill.runtime.v1.Model:orders")).toBe("models");
    });

    it("should be case-insensitive", () => {
      expect(tokenForSeedString("MODEL:orders")).toBe("models");
      expect(tokenForSeedString("METRICS")).toBe("metrics");
    });
  });

  describe("expandSeedsByKind", () => {
    const mockCoerceKind = (res: V1Resource) =>
      res.meta?.name?.kind as ResourceKind;

    it("should keep explicit seeds unchanged", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["model:orders"],
        resources,
        mockCoerceKind,
      );

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should expand 'models' kind token to all Model resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "customers" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "raw_data" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["models"], resources, mockCoerceKind);

      expect(result).toHaveLength(2);
      expect(result).toContainEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
      expect(result).toContainEqual({
        kind: ResourceKind.Model,
        name: "customers",
      });
    });

    it("should expand 'sources' kind token to all Source resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "raw_orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "raw_users" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["sources"], resources, mockCoerceKind);

      expect(result).toHaveLength(2);
      expect(result).toContainEqual({
        kind: ResourceKind.Source,
        name: "raw_orders",
      });
      expect(result).toContainEqual({
        kind: ResourceKind.Source,
        name: "raw_users",
      });
    });

    it("should expand 'metrics' kind token to all MetricsView resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "revenue" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "sales" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["metrics"], resources, mockCoerceKind);

      expect(result).toHaveLength(2);
      expect(result).toContainEqual({
        kind: ResourceKind.MetricsView,
        name: "revenue",
      });
      expect(result).toContainEqual({
        kind: ResourceKind.MetricsView,
        name: "sales",
      });
    });

    it("should expand 'dashboards' kind token to all Explore resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Explore, name: "main_dashboard" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Explore, name: "sales_dashboard" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["dashboards"],
        resources,
        mockCoerceKind,
      );

      expect(result).toHaveLength(2);
    });

    it("should handle mix of explicit seeds and kind tokens", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "customers" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Source, name: "raw_data" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["model:orders", "sources"],
        resources,
        mockCoerceKind,
      );

      expect(result).toHaveLength(2);
      expect(result).toContainEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
      expect(result).toContainEqual({
        kind: ResourceKind.Source,
        name: "raw_data",
      });
    });

    it("should deduplicate seeds", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["model:orders", "models", "model:orders"],
        resources,
        mockCoerceKind,
      );

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should filter out hidden resources", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "hidden_model" },
            hidden: true,
          },
        },
      ];

      const result = expandSeedsByKind(["models"], resources, mockCoerceKind);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should filter out resources not allowed for graph", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: {
              kind: "rill.runtime.v1.Component" as ResourceKind,
              name: "button",
            },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["models"], resources, mockCoerceKind);

      // Only Model should be included (Component is not in ALLOWED_FOR_GRAPH)
      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should handle empty seed array", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind([], resources, mockCoerceKind);

      expect(result).toEqual([]);
    });

    it("should handle undefined seed array", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(undefined, resources, mockCoerceKind);

      expect(result).toEqual([]);
    });

    it("should handle empty resources array", () => {
      const result = expandSeedsByKind(["models"], [], mockCoerceKind);

      expect(result).toEqual([]);
    });

    it("should handle name-only seeds (default to MetricsView)", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.MetricsView, name: "revenue" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["revenue"], resources, mockCoerceKind);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.MetricsView,
        name: "revenue",
      });
    });

    it("should skip resources with missing name", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["models"], resources, mockCoerceKind);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should skip resources with missing kind", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { name: "no_kind" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(["models"], resources, mockCoerceKind);

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });

    it("should use coerceKind function for kind determination", () => {
      const customCoerceKind = (res: V1Resource) => {
        // Simulate coercing models to sources
        const kind = res.meta?.name?.kind as ResourceKind;
        if (kind === ResourceKind.Model && res.meta?.name?.name === "special") {
          return ResourceKind.Source;
        }
        return kind;
      };

      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "special" },
            hidden: false,
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "normal" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["sources"],
        resources,
        customCoerceKind,
      );

      // Should find the "special" model that's coerced to Source
      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "special",
      });
    });

    it("should handle null or falsy seeds in array", () => {
      const resources: V1Resource[] = [
        {
          meta: {
            name: { kind: ResourceKind.Model, name: "orders" },
            hidden: false,
          },
        },
      ];

      const result = expandSeedsByKind(
        ["model:orders", "", null, undefined] as string[],
        resources,
        mockCoerceKind,
      );

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        kind: ResourceKind.Model,
        name: "orders",
      });
    });
  });

  describe("ALLOWED_FOR_GRAPH", () => {
    it("should include Source, Model, MetricsView, and Explore", () => {
      expect(ALLOWED_FOR_GRAPH.has(ResourceKind.Source)).toBe(true);
      expect(ALLOWED_FOR_GRAPH.has(ResourceKind.Model)).toBe(true);
      expect(ALLOWED_FOR_GRAPH.has(ResourceKind.MetricsView)).toBe(true);
      expect(ALLOWED_FOR_GRAPH.has(ResourceKind.Explore)).toBe(true);
    });

    it("should have exactly 4 kinds", () => {
      expect(ALLOWED_FOR_GRAPH.size).toBe(4);
    });
  });
});
