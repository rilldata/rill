import { describe, it, expect } from "vitest";
import {
  createResourceId,
  parseResourceId,
  resourceNameToId,
} from "./resource-utils";
import type { V1ResourceMeta, V1ResourceName } from "@rilldata/web-common/runtime-client";

describe("resource-utils", () => {
  describe("resourceNameToId", () => {
    it("should create valid ID from resource name", () => {
      const resourceName: V1ResourceName = {
        kind: "rill.runtime.v1.Model",
        name: "orders",
      };
      expect(resourceNameToId(resourceName)).toBe("rill.runtime.v1.Model:orders");
    });

    it("should handle different resource kinds", () => {
      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Source", name: "raw_data" })
      ).toBe("rill.runtime.v1.Source:raw_data");

      expect(
        resourceNameToId({ kind: "rill.runtime.v1.MetricsView", name: "revenue" })
      ).toBe("rill.runtime.v1.MetricsView:revenue");

      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Explore", name: "dashboard" })
      ).toBe("rill.runtime.v1.Explore:dashboard");
    });

    it("should handle names with special characters", () => {
      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Model", name: "user_orders_2024" })
      ).toBe("rill.runtime.v1.Model:user_orders_2024");

      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Model", name: "orders-v2" })
      ).toBe("rill.runtime.v1.Model:orders-v2");
    });

    it("should return undefined for null input", () => {
      expect(resourceNameToId(null)).toBeUndefined();
    });

    it("should return undefined for undefined input", () => {
      expect(resourceNameToId(undefined)).toBeUndefined();
    });

    it("should return undefined when kind is missing", () => {
      expect(resourceNameToId({ kind: "", name: "orders" })).toBeUndefined();
      expect(resourceNameToId({ name: "orders" } as V1ResourceName)).toBeUndefined();
    });

    it("should return undefined when name is missing", () => {
      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Model", name: "" })
      ).toBeUndefined();
      expect(
        resourceNameToId({ kind: "rill.runtime.v1.Model" } as V1ResourceName)
      ).toBeUndefined();
    });

    it("should return undefined when both kind and name are missing", () => {
      expect(resourceNameToId({} as V1ResourceName)).toBeUndefined();
    });
  });

  describe("createResourceId", () => {
    it("should create valid ID from resource metadata", () => {
      const meta: V1ResourceMeta = {
        name: {
          kind: "rill.runtime.v1.Model",
          name: "orders",
        },
      };
      expect(createResourceId(meta)).toBe("rill.runtime.v1.Model:orders");
    });

    it("should handle metadata with refs and other properties", () => {
      const meta: V1ResourceMeta = {
        name: {
          kind: "rill.runtime.v1.Model",
          name: "clean_orders",
        },
        refs: [
          { kind: "rill.runtime.v1.Source", name: "raw_orders" },
        ],
        reconcileError: "",
        hidden: false,
      };
      expect(createResourceId(meta)).toBe("rill.runtime.v1.Model:clean_orders");
    });

    it("should return undefined for undefined metadata", () => {
      expect(createResourceId(undefined)).toBeUndefined();
    });

    it("should return undefined for empty metadata", () => {
      expect(createResourceId({})).toBeUndefined();
    });

    it("should return undefined when metadata.name is missing", () => {
      const meta: V1ResourceMeta = {
        refs: [],
      };
      expect(createResourceId(meta)).toBeUndefined();
    });

    it("should return undefined when metadata.name.kind is missing", () => {
      const meta: V1ResourceMeta = {
        name: { name: "orders" } as V1ResourceName,
      };
      expect(createResourceId(meta)).toBeUndefined();
    });

    it("should return undefined when metadata.name.name is missing", () => {
      const meta: V1ResourceMeta = {
        name: { kind: "rill.runtime.v1.Model" } as V1ResourceName,
      };
      expect(createResourceId(meta)).toBeUndefined();
    });
  });

  describe("parseResourceId", () => {
    it("should parse valid resource ID", () => {
      const result = parseResourceId("rill.runtime.v1.Model:orders");
      expect(result).toEqual({
        kind: "rill.runtime.v1.Model",
        name: "orders",
      });
    });

    it("should parse different resource kinds", () => {
      expect(parseResourceId("rill.runtime.v1.Source:raw_data")).toEqual({
        kind: "rill.runtime.v1.Source",
        name: "raw_data",
      });

      expect(parseResourceId("rill.runtime.v1.MetricsView:revenue")).toEqual({
        kind: "rill.runtime.v1.MetricsView",
        name: "revenue",
      });

      expect(parseResourceId("rill.runtime.v1.Explore:dashboard")).toEqual({
        kind: "rill.runtime.v1.Explore",
        name: "dashboard",
      });
    });

    it("should handle names with special characters", () => {
      expect(parseResourceId("rill.runtime.v1.Model:user_orders_2024")).toEqual({
        kind: "rill.runtime.v1.Model",
        name: "user_orders_2024",
      });

      expect(parseResourceId("rill.runtime.v1.Model:orders-v2")).toEqual({
        kind: "rill.runtime.v1.Model",
        name: "orders-v2",
      });
    });

    it("should handle kind with multiple colons in name", () => {
      // Edge case: name contains colons (should split on first colon only)
      expect(parseResourceId("rill.runtime.v1.Model:table:column")).toEqual({
        kind: "rill.runtime.v1.Model",
        name: "table:column",
      });
    });

    it("should return null for empty string", () => {
      expect(parseResourceId("")).toBeNull();
    });

    it("should return null when no colon present", () => {
      expect(parseResourceId("orders")).toBeNull();
    });

    it("should return null when colon is at the start", () => {
      expect(parseResourceId(":orders")).toBeNull();
    });

    it("should return null when colon is at the end", () => {
      expect(parseResourceId("rill.runtime.v1.Model:")).toBeNull();
    });

    it("should return null when kind is empty", () => {
      expect(parseResourceId(":orders")).toBeNull();
    });

    it("should return null when name is empty", () => {
      expect(parseResourceId("rill.runtime.v1.Model:")).toBeNull();
    });

    it("should return null when only colon is present", () => {
      expect(parseResourceId(":")).toBeNull();
    });
  });

  describe("round-trip conversion", () => {
    it("should preserve data through resourceNameToId and parseResourceId", () => {
      const resourceName: V1ResourceName = {
        kind: "rill.runtime.v1.Model",
        name: "orders",
      };

      const id = resourceNameToId(resourceName);
      const parsed = parseResourceId(id!);

      expect(parsed).toEqual(resourceName);
    });

    it("should handle complex resource names", () => {
      const resourceName: V1ResourceName = {
        kind: "rill.runtime.v1.MetricsView",
        name: "revenue_by_region_2024",
      };

      const id = resourceNameToId(resourceName);
      const parsed = parseResourceId(id!);

      expect(parsed).toEqual(resourceName);
    });

    it("should preserve data through createResourceId and parseResourceId", () => {
      const meta: V1ResourceMeta = {
        name: {
          kind: "rill.runtime.v1.Source",
          name: "raw_events",
        },
      };

      const id = createResourceId(meta);
      const parsed = parseResourceId(id!);

      expect(parsed).toEqual(meta.name);
    });
  });
});
