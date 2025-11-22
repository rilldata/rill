import { describe, it, expect } from "vitest";
import { ResourceId, ResourceIdError } from "./resource-id";
import type {
  V1ResourceName,
  V1ResourceMeta,
} from "@rilldata/web-common/runtime-client";

describe("ResourceId", () => {
  describe("create", () => {
    it("should create valid ResourceId", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.kind).toBe("rill.runtime.v1.Model");
      expect(id.name).toBe("orders");
    });

    it("should throw on empty kind", () => {
      expect(() => ResourceId.create("", "orders")).toThrow(ResourceIdError);
      expect(() => ResourceId.create("   ", "orders")).toThrow(ResourceIdError);
    });

    it("should throw on empty name", () => {
      expect(() => ResourceId.create("rill.runtime.v1.Model", "")).toThrow(
        ResourceIdError,
      );
      expect(() => ResourceId.create("rill.runtime.v1.Model", "   ")).toThrow(
        ResourceIdError,
      );
    });

    it("should throw on invalid characters in kind", () => {
      expect(() => ResourceId.create("model<script>", "orders")).toThrow(
        ResourceIdError,
      );
      expect(() => ResourceId.create("model|test", "orders")).toThrow(
        ResourceIdError,
      );
    });

    it("should throw on invalid characters in name", () => {
      expect(() =>
        ResourceId.create("rill.runtime.v1.Model", "<script>alert</script>"),
      ).toThrow(ResourceIdError);
    });

    it("should throw on separator in kind", () => {
      expect(() => ResourceId.create("model:test", "orders")).toThrow(
        ResourceIdError,
      );
    });

    it("should allow separator in name (for complex naming)", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "table:column");
      expect(id.name).toBe("table:column");
    });

    it("should throw on excessive length", () => {
      const longString = "a".repeat(300);
      expect(() => ResourceId.create(longString, "orders")).toThrow(
        ResourceIdError,
      );
      expect(() => ResourceId.create("model", longString)).toThrow(
        ResourceIdError,
      );
    });
  });

  describe("tryCreate", () => {
    it("should return ResourceId on success", () => {
      const id = ResourceId.tryCreate("rill.runtime.v1.Model", "orders");
      expect(id).not.toBeNull();
      expect(id?.kind).toBe("rill.runtime.v1.Model");
    });

    it("should return null on failure", () => {
      const id = ResourceId.tryCreate("", "orders");
      expect(id).toBeNull();
    });
  });

  describe("parse", () => {
    it("should parse valid ID string", () => {
      const id = ResourceId.parse("rill.runtime.v1.Model:orders");
      expect(id.kind).toBe("rill.runtime.v1.Model");
      expect(id.name).toBe("orders");
    });

    it("should handle name with colons (split on first colon)", () => {
      const id = ResourceId.parse("rill.runtime.v1.Model:table:column");
      expect(id.kind).toBe("rill.runtime.v1.Model");
      expect(id.name).toBe("table:column");
    });

    it("should throw on missing separator", () => {
      expect(() => ResourceId.parse("orders")).toThrow(ResourceIdError);
    });

    it("should throw on separator at start", () => {
      expect(() => ResourceId.parse(":orders")).toThrow(ResourceIdError);
    });

    it("should throw on separator at end", () => {
      expect(() => ResourceId.parse("model:")).toThrow(ResourceIdError);
    });

    it("should throw on empty string", () => {
      expect(() => ResourceId.parse("")).toThrow(ResourceIdError);
    });
  });

  describe("tryParse", () => {
    it("should return ResourceId on success", () => {
      const id = ResourceId.tryParse("rill.runtime.v1.Model:orders");
      expect(id).not.toBeNull();
      expect(id?.kind).toBe("rill.runtime.v1.Model");
    });

    it("should return null on failure", () => {
      const id = ResourceId.tryParse("invalid");
      expect(id).toBeNull();
    });
  });

  describe("fromResourceName", () => {
    it("should create from valid V1ResourceName", () => {
      const resourceName: V1ResourceName = {
        kind: "rill.runtime.v1.Model",
        name: "orders",
      };
      const id = ResourceId.fromResourceName(resourceName);
      expect(id).not.toBeNull();
      expect(id?.kind).toBe("rill.runtime.v1.Model");
      expect(id?.name).toBe("orders");
    });

    it("should return null for missing kind", () => {
      const resourceName: V1ResourceName = {
        kind: "",
        name: "orders",
      };
      const id = ResourceId.fromResourceName(resourceName);
      expect(id).toBeNull();
    });

    it("should return null for missing name", () => {
      const resourceName: V1ResourceName = {
        kind: "rill.runtime.v1.Model",
        name: "",
      };
      const id = ResourceId.fromResourceName(resourceName);
      expect(id).toBeNull();
    });

    it("should return null for null input", () => {
      const id = ResourceId.fromResourceName(null);
      expect(id).toBeNull();
    });
  });

  describe("fromMeta", () => {
    it("should create from valid metadata", () => {
      const meta: V1ResourceMeta = {
        name: {
          kind: "rill.runtime.v1.Model",
          name: "orders",
        },
      };
      const id = ResourceId.fromMeta(meta);
      expect(id).not.toBeNull();
      expect(id?.kind).toBe("rill.runtime.v1.Model");
    });

    it("should return null for null meta", () => {
      const id = ResourceId.fromMeta(null);
      expect(id).toBeNull();
    });

    it("should return null for meta without name", () => {
      const meta: V1ResourceMeta = {};
      const id = ResourceId.fromMeta(meta);
      expect(id).toBeNull();
    });
  });

  describe("toString", () => {
    it("should convert to string format", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.toString()).toBe("rill.runtime.v1.Model:orders");
    });

    it("should handle complex names", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "user_orders_2024");
      expect(id.toString()).toBe("rill.runtime.v1.Model:user_orders_2024");
    });
  });

  describe("toResourceName", () => {
    it("should convert to V1ResourceName", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      const resourceName = id.toResourceName();
      expect(resourceName.kind).toBe("rill.runtime.v1.Model");
      expect(resourceName.name).toBe("orders");
    });
  });

  describe("equals", () => {
    it("should equal another ResourceId with same kind and name", () => {
      const id1 = ResourceId.create("rill.runtime.v1.Model", "orders");
      const id2 = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id1.equals(id2)).toBe(true);
    });

    it("should not equal ResourceId with different kind", () => {
      const id1 = ResourceId.create("rill.runtime.v1.Model", "orders");
      const id2 = ResourceId.create("rill.runtime.v1.Source", "orders");
      expect(id1.equals(id2)).toBe(false);
    });

    it("should not equal ResourceId with different name", () => {
      const id1 = ResourceId.create("rill.runtime.v1.Model", "orders");
      const id2 = ResourceId.create("rill.runtime.v1.Model", "users");
      expect(id1.equals(id2)).toBe(false);
    });

    it("should equal string representation", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.equals("rill.runtime.v1.Model:orders")).toBe(true);
    });

    it("should not equal invalid string", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.equals("invalid")).toBe(false);
    });
  });

  describe("getCacheKey", () => {
    it("should generate cache key with default namespace", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.getCacheKey()).toBe("global:rill.runtime.v1.Model:orders");
    });

    it("should generate cache key with custom namespace", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.getCacheKey("dashboard")).toBe(
        "dashboard:rill.runtime.v1.Model:orders",
      );
    });
  });

  describe("isKind", () => {
    it("should return true for matching kind", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.isKind("rill.runtime.v1.Model")).toBe(true);
    });

    it("should return false for non-matching kind", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.isKind("rill.runtime.v1.Source")).toBe(false);
    });
  });

  describe("kindIncludes", () => {
    it("should return true when kind contains substring", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.kindIncludes("Model")).toBe(true);
    });

    it("should be case-insensitive", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.kindIncludes("model")).toBe(true);
      expect(id.kindIncludes("MODEL")).toBe(true);
    });

    it("should return false when kind does not contain substring", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "orders");
      expect(id.kindIncludes("Source")).toBe(false);
    });
  });

  describe("sanitize", () => {
    it("should remove invalid characters", () => {
      expect(ResourceId.sanitize("test<script>")).toBe("test_script_");
      expect(ResourceId.sanitize("test|name")).toBe("test_name");
    });

    it("should trim whitespace", () => {
      expect(ResourceId.sanitize("  test  ")).toBe("test");
    });

    it("should handle empty string", () => {
      expect(ResourceId.sanitize("")).toBe("");
    });
  });

  describe("round-trip conversion", () => {
    it("should preserve data through toString and parse", () => {
      const original = ResourceId.create("rill.runtime.v1.Model", "orders");
      const str = original.toString();
      const parsed = ResourceId.parse(str);

      expect(parsed.equals(original)).toBe(true);
    });

    it("should preserve data through toResourceName and fromResourceName", () => {
      const original = ResourceId.create(
        "rill.runtime.v1.MetricsView",
        "revenue",
      );
      const resourceName = original.toResourceName();
      const recreated = ResourceId.fromResourceName(resourceName);

      expect(recreated?.equals(original)).toBe(true);
    });
  });

  describe("edge cases", () => {
    it("should handle special characters in name", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "user_orders-2024");
      expect(id.name).toBe("user_orders-2024");
    });

    it("should handle very short names", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "a");
      expect(id.name).toBe("a");
    });

    it("should handle names with numbers", () => {
      const id = ResourceId.create("rill.runtime.v1.Model", "table123");
      expect(id.name).toBe("table123");
    });

    it("should handle fully qualified kinds", () => {
      const id = ResourceId.create(
        "rill.runtime.v1.MetricsView",
        "dashboard_metrics",
      );
      expect(id.kind).toBe("rill.runtime.v1.MetricsView");
    });
  });
});
