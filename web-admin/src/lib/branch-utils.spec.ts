import { describe, it, expect, beforeEach } from "vitest";
import {
  extractBranchFromPath,
  injectBranchIntoPath,
  removeBranchFromPath,
  branchPathPrefix,
  requestSkipBranchInjection,
  consumeSkipBranchInjection,
} from "./branch-utils";

describe("branch-utils", () => {
  describe("extractBranchFromPath", () => {
    it("returns undefined for production paths (no @branch)", () => {
      expect(extractBranchFromPath("/acme/analytics")).toBeUndefined();
      expect(
        extractBranchFromPath("/acme/analytics/explore/revenue-overview"),
      ).toBeUndefined();
    });

    it("extracts a simple branch name", () => {
      expect(
        extractBranchFromPath(
          "/acme/analytics/@staging/explore/revenue-overview",
        ),
      ).toBe("staging");
    });

    it("extracts branch from path with no trailing segments", () => {
      expect(
        extractBranchFromPath("/acme/analytics/@q4-dashboard-refresh"),
      ).toBe("q4-dashboard-refresh");
    });

    it("decodes branches containing / (encoded as ~)", () => {
      expect(
        extractBranchFromPath(
          "/acme/analytics/@eric~revenue-metrics/explore/revenue-overview",
        ),
      ).toBe("eric/revenue-metrics");
    });

    it("decodes percent-encoded characters", () => {
      expect(
        extractBranchFromPath(
          "/acme/analytics/@WIP%20churn%20model/explore/churn",
        ),
      ).toBe("WIP churn model");
    });

    it("handles branches with multiple / segments", () => {
      expect(
        extractBranchFromPath(
          "/acme/analytics/@team~maya~funnel-rework/explore/conversion",
        ),
      ).toBe("team/maya/funnel-rework");
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

  describe("injectBranchIntoPath", () => {
    it("injects branch after the project segment", () => {
      expect(
        injectBranchIntoPath(
          "/acme/analytics/explore/revenue-overview",
          "staging",
        ),
      ).toBe("/acme/analytics/@staging/explore/revenue-overview");
    });

    it("injects branch into a path with only org/project", () => {
      expect(
        injectBranchIntoPath("/acme/analytics", "q4-dashboard-refresh"),
      ).toBe("/acme/analytics/@q4-dashboard-refresh");
    });

    it("encodes / in branch names as ~", () => {
      expect(
        injectBranchIntoPath(
          "/acme/analytics/explore/conversion",
          "eric/revenue-metrics",
        ),
      ).toBe("/acme/analytics/@eric~revenue-metrics/explore/conversion");
    });

    it("returns original path if fewer than 3 segments", () => {
      expect(injectBranchIntoPath("/acme", "staging")).toBe("/acme");
      expect(injectBranchIntoPath("/", "staging")).toBe("/");
    });
  });

  describe("removeBranchFromPath", () => {
    it("removes @branch from the path", () => {
      expect(
        removeBranchFromPath(
          "/acme/analytics/@staging/explore/revenue-overview",
        ),
      ).toBe("/acme/analytics/explore/revenue-overview");
    });

    it("removes @branch when it is the last segment", () => {
      expect(removeBranchFromPath("/acme/analytics/@staging")).toBe(
        "/acme/analytics",
      );
    });

    it("returns the path unchanged if no @branch present", () => {
      expect(
        removeBranchFromPath("/acme/analytics/explore/revenue-overview"),
      ).toBe("/acme/analytics/explore/revenue-overview");
    });

    it("handles encoded branch names", () => {
      expect(
        removeBranchFromPath(
          "/acme/analytics/@eric~revenue-metrics/explore/conversion",
        ),
      ).toBe("/acme/analytics/explore/conversion");
    });
  });

  describe("round-trip: inject then extract", () => {
    const branches = [
      "main",
      "staging",
      "eric/revenue-metrics",
      "team/maya/funnel-rework",
      "q4-dashboard-refresh",
    ];
    const basePath = "/acme/analytics/explore/revenue-overview";

    for (const branch of branches) {
      it(`round-trips "${branch}"`, () => {
        const injected = injectBranchIntoPath(basePath, branch);
        expect(extractBranchFromPath(injected)).toBe(branch);
      });
    }
  });

  describe("round-trip: inject then remove", () => {
    it("restores the original path", () => {
      const original = "/acme/analytics/explore/revenue-overview";
      const injected = injectBranchIntoPath(original, "eric/revenue-metrics");
      expect(removeBranchFromPath(injected)).toBe(original);
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

  describe("skipBranchInjection flag", () => {
    beforeEach(() => {
      consumeSkipBranchInjection();
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
  });
});
