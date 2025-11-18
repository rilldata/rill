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

    it("should return Source kind for models defined-as-source with matching table name", () => {
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
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Source);
    });

    it("should return Model kind for models defined-as-source with non-matching table name", () => {
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
            resultTable: "different_table",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when definedAsSource is undefined", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {},
          state: {
            resultTable: "orders",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when definedAsSource is null", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: null,
          },
          state: {
            resultTable: "orders",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when resultTable is undefined", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: true,
          },
          state: {},
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when resultTable is empty string", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: true,
          },
          state: {
            resultTable: "",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when name is undefined but other conditions match", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
          },
        },
        model: {
          spec: {
            definedAsSource: true,
          },
          state: {
            resultTable: "some_table",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when model.state is undefined", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          spec: {
            definedAsSource: true,
          },
        },
      };
      expect(coerceResourceKind(resource)).toBe(ResourceKind.Model);
    });

    it("should return Model kind when model.spec is undefined", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
        model: {
          state: {
            resultTable: "orders",
          },
        },
      };
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

    it("should return undefined when resource has no meta", () => {
      const resource: V1Resource = {};
      expect(coerceResourceKind(resource)).toBeUndefined();
    });

    it("should return undefined when resource meta has no name", () => {
      const resource: V1Resource = {
        meta: {},
      };
      expect(coerceResourceKind(resource)).toBeUndefined();
    });

    it("should return undefined when resource meta.name has no kind", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            name: "orders",
          },
        },
      };
      expect(coerceResourceKind(resource)).toBeUndefined();
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
