import { describe, it, expect } from "vitest";
import { coerceResourceKind, ResourceKind } from "./resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

describe("resource-selectors", () => {
  describe("coerceResourceKind", () => {
    it("should return Model kind for models with model dependencies", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
          refs: [
            { kind: ResourceKind.Model, name: "raw_orders" },
          ],
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

    it("should return Source kind for root models with no model dependencies", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "raw_orders",
          },
          refs: [], // No model refs - this is a root model
        },
        model: {
          spec: {
            definedAsSource: false,
          },
          state: {
            resultTable: "raw_orders",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
    });

    it("should return Source kind for models with only connector refs (no model refs)", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "raw_data",
          },
          refs: [
            { kind: ResourceKind.Connector, name: "duckdb" },
          ],
        },
        model: {
          spec: {
            definedAsSource: false,
          },
          state: {
            resultTable: "raw_data",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
    });

    it("should return Source kind for models defined-as-source with matching table name", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "raw_orders",
          },
          refs: [
            { kind: ResourceKind.Model, name: "other_model" },
          ],
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
      // definedAsSource takes precedence even with model refs
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
    });

    it("should return Model kind for models defined-as-source with non-matching table name", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "raw_orders",
          },
          refs: [
            { kind: ResourceKind.Model, name: "other_model" },
          ],
        },
        model: {
          spec: {
            definedAsSource: true,
          },
          state: {
            resultTable: "different_table",
          },
        },
      };
      // Has model deps and definedAsSource doesn't match table name
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should pass through Source kind unchanged", () => {
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
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
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

    it("should return Source for models with undefined refs (treated as no model deps)", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
          // No refs property - should be treated as no model dependencies
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
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
    });
  });
});
