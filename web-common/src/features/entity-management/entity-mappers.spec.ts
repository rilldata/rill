import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { describe, it, expect } from "vitest";

describe("entity-mappers", () => {
  describe("getNameFromFile", () => {
    it("happy path", () => {
      expect(getNameFromFile("data/adbids.csv")).toBe("adbids");
    });

    it("absolute path", () => {
      expect(getNameFromFile("/data/adbids.csv")).toBe("adbids");
    });

    it("multiple paths", () => {
      expect(getNameFromFile("/path/to/data/adbids.csv")).toBe("adbids");
    });

    it("multiple extensions", () => {
      expect(getNameFromFile("/path/to/data/adbids.csv.tgz")).toBe("adbids");
    });

    it("keeps dots in YAML resource names", () => {
      expect(getNameFromFile("/dashboards/dashboard.canvas.yaml")).toBe(
        "dashboard.canvas",
      );
    });

    it("keeps dots in YML resource names", () => {
      expect(getNameFromFile("/dashboards/dashboard.canvas.yml")).toBe(
        "dashboard.canvas",
      );
    });

    it("keeps dots in SQL resource names", () => {
      expect(getNameFromFile("/models/orders.latest.sql")).toBe(
        "orders.latest",
      );
    });

    it("no folder", () => {
      expect(getNameFromFile("adbids.csv")).toBe("adbids");
    });

    it("no extension", () => {
      expect(getNameFromFile("data/adbids")).toBe("adbids");
    });

    it("no folder and extension", () => {
      expect(getNameFromFile("adbids")).toBe("adbids");
    });
  });
});
