import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import type { BeforeNavigate } from "@sveltejs/kit";
import {
  extractBranchFromPath,
  injectBranchIntoPath,
  removeBranchFromPath,
  branchPathPrefix,
  requestSkipBranchInjection,
  consumeSkipBranchInjection,
  getBranchRedirect,
  handleBranchNavigation,
} from "./branch-utils";

// Shared test data: every entry exercises extract, inject, and remove together.
const branchPaths: {
  name: string;
  branchPath: string;
  basePath: string;
  branch: string;
}[] = [
  {
    name: "simple branch with trailing route",
    branchPath: "/acme/analytics/@staging/explore/revenue-overview",
    basePath: "/acme/analytics/explore/revenue-overview",
    branch: "staging",
  },
  {
    name: "branch as last segment (no trailing route)",
    branchPath: "/acme/analytics/@q4-dashboard-refresh",
    basePath: "/acme/analytics",
    branch: "q4-dashboard-refresh",
  },
  {
    name: "branch containing / (encoded as ~)",
    branchPath:
      "/acme/analytics/@eric~revenue-metrics/explore/revenue-overview",
    basePath: "/acme/analytics/explore/revenue-overview",
    branch: "eric/revenue-metrics",
  },
  {
    name: "branch with multiple / segments",
    branchPath: "/acme/analytics/@team~maya~funnel-rework/explore/conversion",
    basePath: "/acme/analytics/explore/conversion",
    branch: "team/maya/funnel-rework",
  },
  {
    name: "percent-encoded characters in branch",
    branchPath: "/acme/analytics/@WIP%20churn%20model/explore/churn",
    basePath: "/acme/analytics/explore/churn",
    branch: "WIP churn model",
  },
];

describe("branch-utils", () => {
  describe("path manipulation (data-driven)", () => {
    for (const { name, branchPath, basePath, branch } of branchPaths) {
      describe(name, () => {
        it("extractBranchFromPath returns the branch", () => {
          expect(extractBranchFromPath(branchPath)).toBe(branch);
        });

        it("removeBranchFromPath returns the base path", () => {
          expect(removeBranchFromPath(branchPath)).toBe(basePath);
        });

        it("injectBranchIntoPath → extractBranchFromPath round-trips", () => {
          const injected = injectBranchIntoPath(basePath, branch);
          expect(extractBranchFromPath(injected)).toBe(branch);
        });

        it("injectBranchIntoPath → removeBranchFromPath restores the stripped path", () => {
          const injected = injectBranchIntoPath(basePath, branch);
          expect(removeBranchFromPath(injected)).toBe(basePath);
        });
      });
    }
  });

  describe("extractBranchFromPath edge cases", () => {
    it("returns undefined for production paths (no @branch)", () => {
      expect(extractBranchFromPath("/acme/analytics")).toBeUndefined();
      expect(
        extractBranchFromPath("/acme/analytics/explore/revenue-overview"),
      ).toBeUndefined();
    });

    it("returns undefined for @ in wrong position", () => {
      expect(
        extractBranchFromPath("/@staging/analytics/explore"),
      ).toBeUndefined();
      expect(extractBranchFromPath("/acme/@staging/explore")).toBeUndefined();
      expect(
        extractBranchFromPath("/acme/analytics/explore/@staging"),
      ).toBeUndefined();
    });

    it("returns undefined for empty pathname", () => {
      expect(extractBranchFromPath("")).toBeUndefined();
    });

    it("returns undefined for root path", () => {
      expect(extractBranchFromPath("/")).toBeUndefined();
    });
  });

  describe("injectBranchIntoPath edge cases", () => {
    it("returns original path if fewer than 3 segments", () => {
      expect(injectBranchIntoPath("/acme", "staging")).toBe("/acme");
      expect(injectBranchIntoPath("/", "staging")).toBe("/");
    });
  });

  describe("removeBranchFromPath edge cases", () => {
    it("returns the path unchanged if no @branch present", () => {
      expect(
        removeBranchFromPath("/acme/analytics/explore/revenue-overview"),
      ).toBe("/acme/analytics/explore/revenue-overview");
    });
  });

  describe("branchPathPrefix", () => {
    it("returns empty string for undefined", () => {
      expect(branchPathPrefix(undefined)).toBe("");
    });

    it("returns empty string for empty string", () => {
      expect(branchPathPrefix("")).toBe("");
    });

    it("returns /@encoded-branch for a simple branch", () => {
      expect(branchPathPrefix("staging")).toBe("/@staging");
    });

    it("encodes / in branch names", () => {
      expect(branchPathPrefix("eric/revenue-metrics")).toBe(
        "/@eric~revenue-metrics",
      );
    });
  });

  describe("getBranchRedirect", () => {
    const org = "acme";
    const proj = "analytics";
    const branch = "staging";
    const url = (path: string) => new URL(`http://localhost${path}`);

    it("returns redirect URL for a project-internal path missing @branch", () => {
      expect(
        getBranchRedirect(
          url("/acme/analytics/explore/revenue-overview"),
          branch,
          org,
          proj,
        ),
      ).toBe("/acme/analytics/@staging/explore/revenue-overview");
    });

    it("preserves search params and hash", () => {
      expect(
        getBranchRedirect(
          url("/acme/analytics/explore/revenue-overview?filter=us#section"),
          branch,
          org,
          proj,
        ),
      ).toBe(
        "/acme/analytics/@staging/explore/revenue-overview?filter=us#section",
      );
    });

    it("returns null if path already has @branch", () => {
      expect(
        getBranchRedirect(
          url("/acme/analytics/@staging/explore/revenue-overview"),
          branch,
          org,
          proj,
        ),
      ).toBeNull();
    });

    it("returns null for paths outside the project", () => {
      expect(
        getBranchRedirect(url("/other-org/other-project"), branch, org, proj),
      ).toBeNull();
    });

    it("returns null for a different project that shares a name prefix", () => {
      expect(
        getBranchRedirect(
          url("/acme/analytics-v2/explore/dashboard"),
          branch,
          org,
          proj,
        ),
      ).toBeNull();
    });

    it("returns null for public share URLs", () => {
      expect(
        getBranchRedirect(
          url("/acme/analytics/-/share/abc123"),
          branch,
          org,
          proj,
        ),
      ).toBeNull();
    });

    it("handles the bare project path", () => {
      expect(getBranchRedirect(url("/acme/analytics"), branch, org, proj)).toBe(
        "/acme/analytics/@staging",
      );
    });
  });

  describe("handleBranchNavigation", () => {
    const org = "acme";
    const proj = "analytics";
    const branch = "staging";

    function makeNav(pathname: string, type: BeforeNavigate["type"] = "link") {
      const nav = {
        from: null,
        to: {
          params: {},
          route: { id: null },
          url: new URL(`http://localhost${pathname}`),
        },
        type,
        willUnload: false,
        complete: Promise.resolve(),
        cancel: vi.fn(),
      } as unknown as BeforeNavigate;
      const navigateFn = vi
        .fn<(url: string) => Promise<void>>()
        .mockResolvedValue(undefined);
      return { nav, navigateFn };
    }

    beforeEach(() => {
      consumeSkipBranchInjection();
    });

    it("redirects a branch-unaware project URL", () => {
      const { nav, navigateFn } = makeNav(
        "/acme/analytics/explore/revenue-overview",
      );
      handleBranchNavigation(nav, branch, org, proj, navigateFn);
      expect(nav.cancel).toHaveBeenCalled();
      expect(navigateFn).toHaveBeenCalledWith(
        "/acme/analytics/@staging/explore/revenue-overview",
      );
    });

    it("does nothing when there is no active branch", () => {
      const { nav, navigateFn } = makeNav(
        "/acme/analytics/explore/revenue-overview",
      );
      handleBranchNavigation(nav, undefined, org, proj, navigateFn);
      expect(nav.cancel).not.toHaveBeenCalled();
      expect(navigateFn).not.toHaveBeenCalled();
    });

    it("does nothing for popstate navigations", () => {
      const { nav, navigateFn } = makeNav(
        "/acme/analytics/explore/revenue-overview",
        "popstate",
      );
      handleBranchNavigation(nav, branch, org, proj, navigateFn);
      expect(nav.cancel).not.toHaveBeenCalled();
      expect(navigateFn).not.toHaveBeenCalled();
    });

    it("preserves search params and hash in redirect", () => {
      const { nav, navigateFn } = makeNav(
        "/acme/analytics/explore/revenue-overview?filter=us#section",
      );
      handleBranchNavigation(nav, branch, org, proj, navigateFn);
      expect(nav.cancel).toHaveBeenCalled();
      expect(navigateFn).toHaveBeenCalledWith(
        "/acme/analytics/@staging/explore/revenue-overview?filter=us#section",
      );
    });

    it("respects the skip flag", () => {
      requestSkipBranchInjection();
      const { nav, navigateFn } = makeNav(
        "/acme/analytics/explore/revenue-overview",
      );
      handleBranchNavigation(nav, branch, org, proj, navigateFn);
      expect(nav.cancel).not.toHaveBeenCalled();
      expect(navigateFn).not.toHaveBeenCalled();
    });

    it("does nothing when nav.to is null", () => {
      const nav = {
        from: null,
        to: null,
        type: "link" as const,
        willUnload: false,
        complete: Promise.resolve(),
        cancel: vi.fn(),
      } as unknown as BeforeNavigate;
      const navigateFn = vi.fn();
      handleBranchNavigation(nav, branch, org, proj, navigateFn);
      expect(nav.cancel).not.toHaveBeenCalled();
      expect(navigateFn).not.toHaveBeenCalled();
    });
  });

  describe("skipBranchInjection flag", () => {
    beforeEach(() => {
      vi.useFakeTimers();
      consumeSkipBranchInjection();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("returns false when not requested", () => {
      expect(consumeSkipBranchInjection()).toBe(false);
    });

    it("returns true after request, then resets", () => {
      requestSkipBranchInjection();
      expect(consumeSkipBranchInjection()).toBe(true);
      expect(consumeSkipBranchInjection()).toBe(false);
    });

    it("only fires once per request", () => {
      requestSkipBranchInjection();
      consumeSkipBranchInjection();
      expect(consumeSkipBranchInjection()).toBe(false);
    });

    it("expires after 500ms if not consumed", () => {
      requestSkipBranchInjection();
      vi.advanceTimersByTime(501);
      expect(consumeSkipBranchInjection()).toBe(false);
    });
  });
});
