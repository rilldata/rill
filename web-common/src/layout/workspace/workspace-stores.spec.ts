import { get } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";

// The module keeps a singleton map of stores keyed by file path, so it is
// re-imported for every test to keep them independent.
async function loadModule() {
  return await import("./workspace-stores");
}

describe("workspace-stores", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  describe("WorkspaceLayoutStore", () => {
    it("restores a known view from localStorage", async () => {
      localStorage.setItem("/a.yaml", JSON.stringify({ view: "code" }));
      const { workspaces } = await loadModule();
      expect(get(workspaces.get("/a.yaml").view)).toBe("code");
    });

    it("ignores an unknown persisted view and falls back to the default", async () => {
      // Older builds could persist an explore web view (e.g. "pivot") here.
      localStorage.setItem("/a.yaml", JSON.stringify({ view: "pivot" }));
      const { workspaces } = await loadModule();
      expect(get(workspaces.get("/a.yaml").view)).toBe("viz");
    });
  });

  describe("consumeViewSearchParam", () => {
    it("returns null when the param is absent", async () => {
      const { consumeViewSearchParam } = await loadModule();
      const url = new URL("http://localhost/files/a.yaml?view=pivot");
      expect(consumeViewSearchParam(url, "/a.yaml")).toBeNull();
      // The explore's own `view` param is left alone.
      expect(url.searchParams.get("view")).toBe("pivot");
    });

    it("stores a known view and strips only its own param", async () => {
      const {
        consumeViewSearchParam,
        workspaces,
        WORKSPACE_VIEW_SEARCH_PARAM,
      } = await loadModule();
      const url = new URL(
        `http://localhost/files/a.yaml?${WORKSPACE_VIEW_SEARCH_PARAM}=explore&view=pivot&tr=P7D`,
      );
      expect(consumeViewSearchParam(url, "/a.yaml")).toBe(
        "/files/a.yaml?view=pivot&tr=P7D",
      );
      expect(get(workspaces.get("/a.yaml").view)).toBe("explore");
    });

    it("strips an unknown view without storing it", async () => {
      const {
        consumeViewSearchParam,
        workspaces,
        WORKSPACE_VIEW_SEARCH_PARAM,
      } = await loadModule();
      const url = new URL(
        `http://localhost/files/a.yaml?${WORKSPACE_VIEW_SEARCH_PARAM}=pivot`,
      );
      expect(consumeViewSearchParam(url, "/a.yaml")).toBe("/files/a.yaml");
      expect(get(workspaces.get("/a.yaml").view)).toBe("viz");
    });
  });
});
