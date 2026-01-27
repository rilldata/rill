import { describe, it, expect } from "vitest";
import { coerceResourceKind, ResourceKind } from "./resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

describe("resource-selectors", () => {
  describe("coerceResourceKind", () => {
    it("should return Model kind for normal models", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: false,
          },
          state: {
            resultTable: "orders_transformed",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind for models defined-as-source (Source is deprecated)", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "raw_orders",
          },
        },
        model: {
          spec: {
            definedAsSource: true,
          },
          state: {
            resultTable: "raw_orders",
          },
        },
      };
      // Models defined-as-source are now treated as Models (Source is deprecated)
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should normalize Source to Model (Source is deprecated)", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Source,
            name: "raw_data",
          },
        },
        source: {
          spec: {
            sourceConnector: "duckdb",
          },
        },
      };
      // Source is normalized to Model (Source is deprecated)
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should pass through MetricsView kind unchanged", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.MetricsView,
            name: "sales_metrics",
          },
        },
        metricsView: {
          spec: {},
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.MetricsView);
    });

    it("should pass through Explore kind unchanged", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Explore,
            name: "dashboard",
          },
        },
        explore: {
          spec: {},
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Explore);
    });

    it("should handle case where definedAsSource is false explicitly", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: false,
          },
          state: {
            resultTable: "orders",
          },
        },
      };
      // Even though table name matches, definedAsSource is false
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });
  });
});
