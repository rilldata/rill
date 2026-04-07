import { describe, expect, it } from "vitest";
import {
  getHomeRoute,
  getExploreRoute,
  getCanvasRoute,
  getFileRoute,
} from "./preview-route-utils";

describe("preview-route-utils", () => {
  describe("getHomeRoute", () => {
    it("returns /dashboards in preview mode", () => {
      expect(getHomeRoute(true)).toBe("/dashboards");
    });

    it("returns / in developer mode", () => {
      expect(getHomeRoute(false)).toBe("/");
    });
  });

  describe("getExploreRoute", () => {
    it("returns /explore/{name} in preview mode", () => {
      expect(
        getExploreRoute(true, "my_explore", "/dashboards/my_explore.yaml"),
      ).toBe("/explore/my_explore");
    });

    it("returns /files/{path} in developer mode", () => {
      expect(
        getExploreRoute(false, "my_explore", "/dashboards/my_explore.yaml"),
      ).toBe("/files/dashboards/my_explore.yaml");
    });
  });

  describe("getCanvasRoute", () => {
    it("returns /canvas/{name} in preview mode", () => {
      expect(
        getCanvasRoute(true, "my_canvas", "/dashboards/my_canvas.yaml"),
      ).toBe("/canvas/my_canvas");
    });

    it("returns /files/{path} in developer mode", () => {
      expect(
        getCanvasRoute(false, "my_canvas", "/dashboards/my_canvas.yaml"),
      ).toBe("/files/dashboards/my_canvas.yaml");
    });
  });

  describe("getFileRoute", () => {
    it("returns /dashboards in preview mode", () => {
      expect(getFileRoute(true, "/metrics/my_metrics.yaml")).toBe(
        "/dashboards",
      );
    });

    it("returns /files/{path} in developer mode", () => {
      expect(getFileRoute(false, "/metrics/my_metrics.yaml")).toBe(
        "/files/metrics/my_metrics.yaml",
      );
    });
  });
});
