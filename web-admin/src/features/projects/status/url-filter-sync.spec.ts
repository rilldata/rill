import { describe, expect, it, vi, beforeEach } from "vitest";
import {
  parseArrayParam,
  parseStringParam,
  parseEnumParam,
  createUrlFilterSync,
} from "./url-filter-sync";

// vi.hoisted runs before vi.mock factories and before imports resolve
const { gotoMock, pageStore } = vi.hoisted(() => {
  type PageData = { url: URL };
  let current: PageData = { url: new URL("http://localhost/status/tables") };
  const subs = new Set<(v: PageData) => void>();
  return {
    gotoMock: vi.fn(),
    pageStore: {
      subscribe(fn: (v: PageData) => void) {
        subs.add(fn);
        fn(current);
        return () => subs.delete(fn);
      },
      set(v: PageData) {
        current = v;
        for (const fn of subs) fn(v);
      },
    },
  };
});

vi.mock("$app/navigation", () => ({
  goto: (...args: unknown[]) => gotoMock(...args),
}));

vi.mock("$app/stores", () => ({
  page: pageStore,
}));

describe("url-filter-sync", () => {
  describe("parseArrayParam", () => {
    it("returns empty array for null", () => {
      expect(parseArrayParam(null)).toEqual([]);
    });

    it("returns empty array for empty string", () => {
      expect(parseArrayParam("")).toEqual([]);
    });

    it("splits comma-separated values", () => {
      expect(parseArrayParam("a,b,c")).toEqual(["a", "b", "c"]);
    });

    it("filters out empty segments", () => {
      expect(parseArrayParam("a,,b")).toEqual(["a", "b"]);
    });

    it("returns single-element array for single value", () => {
      expect(parseArrayParam("error")).toEqual(["error"]);
    });
  });

  describe("parseStringParam", () => {
    it("returns empty string for null", () => {
      expect(parseStringParam(null)).toBe("");
    });

    it("returns the value for a non-null string", () => {
      expect(parseStringParam("hello")).toBe("hello");
    });

    it("returns empty string as-is", () => {
      expect(parseStringParam("")).toBe("");
    });
  });

  describe("parseEnumParam", () => {
    const allowed = ["all", "table", "view"] as const;

    it("returns default for null", () => {
      expect(parseEnumParam(null, allowed, "all")).toBe("all");
    });

    it("returns default for invalid value", () => {
      expect(parseEnumParam("invalid", allowed, "all")).toBe("all");
    });

    it("returns the value when it matches an allowed member", () => {
      expect(parseEnumParam("table", allowed, "all")).toBe("table");
      expect(parseEnumParam("view", allowed, "all")).toBe("view");
    });

    it("returns the default value itself when passed as raw", () => {
      expect(parseEnumParam("all", allowed, "all")).toBe("all");
    });
  });

  describe("createUrlFilterSync", () => {
    beforeEach(() => {
      gotoMock.mockReset();
      pageStore.set({ url: new URL("http://localhost/status/tables") });
    });

    it("init sets the baseline search string", () => {
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);
      const url = new URL("http://localhost/status/tables?q=test");

      sync.init(url);

      expect(sync.hasExternalNavigation(url)).toBe(false);
    });

    it("hasExternalNavigation returns true when URL search differs", () => {
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);
      sync.init(new URL("http://localhost/status/tables"));

      const changed = new URL("http://localhost/status/tables?q=test");
      expect(sync.hasExternalNavigation(changed)).toBe(true);
    });

    it("markSynced updates the baseline so hasExternalNavigation returns false", () => {
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);
      sync.init(new URL("http://localhost/status/tables"));

      const changed = new URL("http://localhost/status/tables?q=test");
      expect(sync.hasExternalNavigation(changed)).toBe(true);

      sync.markSynced(changed);
      expect(sync.hasExternalNavigation(changed)).toBe(false);
    });

    it("syncToUrl sets string params and calls goto", () => {
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);
      sync.init(new URL("http://localhost/status/tables"));

      sync.syncToUrl({ q: "hello" });

      expect(gotoMock).toHaveBeenCalledWith("/status/tables?q=hello", {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    });

    it("syncToUrl removes string param when empty", () => {
      pageStore.set({
        url: new URL("http://localhost/status/tables?q=old"),
      });
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);

      sync.syncToUrl({ q: "" });

      expect(gotoMock).toHaveBeenCalledWith("/status/tables", {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    });

    it("syncToUrl serializes array params as comma-separated", () => {
      const sync = createUrlFilterSync([{ key: "level", type: "array" }]);

      sync.syncToUrl({ level: ["error", "warn"] });

      expect(gotoMock).toHaveBeenCalledWith(
        "/status/tables?level=error%2Cwarn",
        {
          replaceState: true,
          noScroll: true,
          keepFocus: true,
        },
      );
    });

    it("syncToUrl removes array param when empty", () => {
      pageStore.set({
        url: new URL("http://localhost/status/tables?level=error"),
      });
      const sync = createUrlFilterSync([{ key: "level", type: "array" }]);

      sync.syncToUrl({ level: [] });

      expect(gotoMock).toHaveBeenCalledWith("/status/tables", {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    });

    it("syncToUrl omits enum param when value equals default", () => {
      const sync = createUrlFilterSync([
        { key: "type", type: "enum", defaultValue: "all" },
      ]);

      sync.syncToUrl({ type: "all" });

      expect(gotoMock).toHaveBeenCalledWith("/status/tables", {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    });

    it("syncToUrl sets enum param when value differs from default", () => {
      const sync = createUrlFilterSync([
        { key: "type", type: "enum", defaultValue: "all" },
      ]);

      sync.syncToUrl({ type: "view" });

      expect(gotoMock).toHaveBeenCalledWith("/status/tables?type=view", {
        replaceState: true,
        noScroll: true,
        keepFocus: true,
      });
    });

    it("syncToUrl handles multiple params together", () => {
      const sync = createUrlFilterSync([
        { key: "q", type: "string" },
        { key: "type", type: "enum", defaultValue: "all" },
      ]);

      sync.syncToUrl({ q: "users", type: "table" });

      expect(gotoMock).toHaveBeenCalledWith(
        "/status/tables?q=users&type=table",
        {
          replaceState: true,
          noScroll: true,
          keepFocus: true,
        },
      );
    });

    it("syncToUrl updates lastSyncedSearch so subsequent hasExternalNavigation is false", () => {
      const sync = createUrlFilterSync([{ key: "q", type: "string" }]);
      sync.init(new URL("http://localhost/status/tables"));

      sync.syncToUrl({ q: "test" });

      const urlAfterSync = new URL("http://localhost/status/tables?q=test");
      expect(sync.hasExternalNavigation(urlAfterSync)).toBe(false);
    });
  });
});
