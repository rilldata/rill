import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  navigateToResourceGraph,
  createGraphNavigationHandler,
  buildGraphUrl,
} from "./navigation-utils";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

// Mock $app/navigation
vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

describe("navigation-utils", () => {
  let gotoMock: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    const { goto } = await import("$app/navigation");
    gotoMock = goto as ReturnType<typeof vi.fn>;
    gotoMock.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("navigateToResourceGraph", () => {
    it("should navigate to graph page with single seed", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "orders");

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders"
      );
    });

    it("should navigate with multiple seeds", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "orders", [
        "rill.runtime.v1.Source:users",
      ]);

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders&seed=rill.runtime.v1.Source%3Ausers"
      );
    });

    it("should handle names with special characters", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "my-model_v2");

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Amy-model_v2"
      );
    });

    it("should handle names with spaces", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "my model");

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Amy%20model"
      );
    });

    it("should not navigate when kind is empty", () => {
      navigateToResourceGraph("", "orders");

      expect(gotoMock).not.toHaveBeenCalled();
    });

    it("should not navigate when name is empty", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "");

      expect(gotoMock).not.toHaveBeenCalled();
    });

    it("should handle empty additional seeds array", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "orders", []);

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders"
      );
    });

    it("should handle undefined additional seeds", () => {
      navigateToResourceGraph("rill.runtime.v1.Model", "orders", undefined);

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders"
      );
    });

    it("should navigate for different resource kinds", () => {
      navigateToResourceGraph("rill.runtime.v1.Source", "raw_data");
      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Source%3Araw_data"
      );

      gotoMock.mockClear();

      navigateToResourceGraph("rill.runtime.v1.MetricsView", "revenue");
      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.MetricsView%3Arevenue"
      );

      gotoMock.mockClear();

      navigateToResourceGraph("rill.runtime.v1.Explore", "dashboard");
      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Explore%3Adashboard"
      );
    });
  });

  describe("buildGraphUrl", () => {
    it("should build URL with single seed", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "orders" },
      ]);

      expect(url).toBe("/graph?seed=rill.runtime.v1.Model%3Aorders");
    });

    it("should build URL with multiple seeds", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "orders" },
        { kind: "rill.runtime.v1.Source", name: "users" },
      ]);

      expect(url).toBe(
        "/graph?seed=rill.runtime.v1.Model%3Aorders&seed=rill.runtime.v1.Source%3Ausers"
      );
    });

    it("should handle empty seeds array", () => {
      const url = buildGraphUrl([]);

      expect(url).toBe("/graph?");
    });

    it("should filter out invalid seeds with empty kind", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "orders" },
        { kind: "", name: "invalid" },
      ]);

      expect(url).toBe("/graph?seed=rill.runtime.v1.Model%3Aorders");
    });

    it("should filter out invalid seeds with empty name", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "orders" },
        { kind: "rill.runtime.v1.Model", name: "" },
      ]);

      expect(url).toBe("/graph?seed=rill.runtime.v1.Model%3Aorders");
    });

    it("should properly encode special characters", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "my-model_v2" },
      ]);

      expect(url).toBe("/graph?seed=rill.runtime.v1.Model%3Amy-model_v2");
    });

    it("should handle names with spaces", () => {
      const url = buildGraphUrl([
        { kind: "rill.runtime.v1.Model", name: "my model" },
      ]);

      expect(url).toBe("/graph?seed=rill.runtime.v1.Model%3Amy%20model");
    });
  });

  describe("createGraphNavigationHandler", () => {
    let consoleWarnSpy: ReturnType<typeof vi.spyOn>;
    let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
      consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    });

    afterEach(() => {
      consoleWarnSpy.mockRestore();
      consoleErrorSpy.mockRestore();
    });

    it("should create handler that navigates with valid resource", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
      };

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => resource
      );

      handler();

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders"
      );
    });

    it("should warn when resource name is missing", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
          },
        },
      };

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => resource
      );

      handler();

      expect(gotoMock).not.toHaveBeenCalled();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        "[TestComponent] Cannot navigate to graph: resource name is missing"
      );
    });

    it("should warn when resource is undefined", () => {
      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => undefined
      );

      handler();

      expect(gotoMock).not.toHaveBeenCalled();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        "[TestComponent] Cannot navigate to graph: resource name is missing"
      );
    });

    it("should warn when resource meta is undefined", () => {
      const resource: V1Resource = {};

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => resource
      );

      handler();

      expect(gotoMock).not.toHaveBeenCalled();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        "[TestComponent] Cannot navigate to graph: resource name is missing"
      );
    });

    it("should warn when resource meta.name is undefined", () => {
      const resource: V1Resource = {
        meta: {},
      };

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => resource
      );

      handler();

      expect(gotoMock).not.toHaveBeenCalled();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        "[TestComponent] Cannot navigate to graph: resource name is missing"
      );
    });

    it("should handle errors thrown by getResource", () => {
      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => {
          throw new Error("Resource fetch failed");
        }
      );

      handler();

      expect(gotoMock).not.toHaveBeenCalled();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "[TestComponent] Failed to navigate to graph:",
        expect.any(Error)
      );
    });

    it("should use component name in error messages", () => {
      const handler = createGraphNavigationHandler(
        "ModelMenuItems",
        "rill.runtime.v1.Model",
        () => undefined
      );

      handler();

      expect(consoleWarnSpy).toHaveBeenCalledWith(
        "[ModelMenuItems] Cannot navigate to graph: resource name is missing"
      );
    });

    it("should work with different resource kinds", () => {
      const sourceResource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Source,
            name: "raw_data",
          },
        },
      };

      const handler = createGraphNavigationHandler(
        "SourceMenuItems",
        "rill.runtime.v1.Source",
        () => sourceResource
      );

      handler();

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Source%3Araw_data"
      );
    });

    it("should be reusable - can be called multiple times", () => {
      let currentResource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "orders",
          },
        },
      };

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => currentResource
      );

      handler();
      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Aorders"
      );

      gotoMock.mockClear();

      // Change the resource and call again
      currentResource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "users",
          },
        },
      };

      handler();
      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Ausers"
      );
    });

    it("should handle resource names with special characters", () => {
      const resource: V1Resource = {
        meta: {
          name: {
            kind: ResourceKind.Model,
            name: "my-model_v2",
          },
        },
      };

      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => resource
      );

      handler();

      expect(gotoMock).toHaveBeenCalledWith(
        "/graph?seed=rill.runtime.v1.Model%3Amy-model_v2"
      );
    });
  });

  describe("error handling edge cases", () => {
    let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
      consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    });

    afterEach(() => {
      consoleErrorSpy.mockRestore();
    });

    it("should handle null resource gracefully", () => {
      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => null as any
      );

      expect(() => handler()).not.toThrow();
      expect(gotoMock).not.toHaveBeenCalled();
    });

    it("should handle getResource returning non-object", () => {
      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => "not a resource" as any
      );

      expect(() => handler()).not.toThrow();
      expect(gotoMock).not.toHaveBeenCalled();
    });

    it("should handle getResource throwing non-Error object", () => {
      const handler = createGraphNavigationHandler(
        "TestComponent",
        "rill.runtime.v1.Model",
        () => {
          throw "string error";
        }
      );

      expect(() => handler()).not.toThrow();
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "[TestComponent] Failed to navigate to graph:",
        "string error"
      );
    });
  });
});
