import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";
import {
  consumeViewSearchParam,
  WORKSPACE_VIEW_SEARCH_PARAM,
  workspaces,
} from "./workspace-stores";

describe("workspace-stores", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  describe("WorkspaceLayoutStore", () => {
    it("restores a known view from localStorage", () => {
      localStorage.setItem(
        "/dashboards/known.yaml",
        JSON.stringify({ view: "code" }),
      );
      expect(get(workspaces.get("/dashboards/known.yaml").view)).toBe("code");
    });

    it("ignores an unknown persisted view and falls back to the default", () => {
      // Older builds could persist an explore web view (e.g. "pivot") here.
      localStorage.setItem(
        "/dashboards/stale.yaml",
        JSON.stringify({ view: "pivot" }),
      );
      expect(get(workspaces.get("/dashboards/stale.yaml").view)).toBe("viz");
    });
  });

  describe("consumeViewSearchParam", () => {
    it("returns null when the param is absent", () => {
      const url = new URL("http://localhost/files/metrics/a.yaml?view=pivot");
      expect(consumeViewSearchParam(url, "/metrics/a.yaml")).toBeNull();
      // The explore's own `view` param is left alone.
      expect(url.searchParams.get("view")).toBe("pivot");
    });

    it("stores a known view and strips only its own param", () => {
      const url = new URL(
        `http://localhost/files/metrics/b.yaml?${WORKSPACE_VIEW_SEARCH_PARAM}=explore&view=pivot&tr=P7D`,
      );
      expect(consumeViewSearchParam(url, "/metrics/b.yaml")).toBe(
        "/files/metrics/b.yaml?view=pivot&tr=P7D",
      );
      expect(get(workspaces.get("/metrics/b.yaml").view)).toBe("explore");
    });

    it("strips an unknown view without storing it", () => {
      const url = new URL(
        `http://localhost/files/dashboards/c.yaml?${WORKSPACE_VIEW_SEARCH_PARAM}=pivot`,
      );
      expect(consumeViewSearchParam(url, "/dashboards/c.yaml")).toBe(
        "/files/dashboards/c.yaml",
      );
      expect(get(workspaces.get("/dashboards/c.yaml").view)).toBe("viz");
    });
  });
});
