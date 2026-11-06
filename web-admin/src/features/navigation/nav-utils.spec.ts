import {
  isPublicAIPage,
  isPublicURLPage,
} from "@rilldata/web-admin/features/navigation/nav-utils";
import type { Page } from "@sveltejs/kit";
import { describe, expect, it } from "vitest";

function pageFor(routeId: string, url: string): Page {
  return { route: { id: routeId }, url: new URL(url) } as Page;
}

describe("isPublicURLPage", () => {
  it("treats share token routes as public", () => {
    expect(
      isPublicURLPage(
        pageFor(
          "/[organization]/[project]/-/share/[token]/explore/[dashboard]",
          "https://ui.rilldata.com/org/proj/-/share/abc/explore/e1",
        ),
      ),
    ).toBe(true);
  });

  it("treats an AI conversation opened with a token as public", () => {
    const page = pageFor(
      "/[organization]/[project]/-/ai/[conversationId]",
      "https://ui.rilldata.com/org/proj/-/ai/session-1?token=abc",
    );
    expect(isPublicAIPage(page)).toBe(true);
    expect(isPublicURLPage(page)).toBe(true);
  });

  it("does not treat an AI conversation without a token as public", () => {
    const page = pageFor(
      "/[organization]/[project]/-/ai/[conversationId]",
      "https://ui.rilldata.com/org/proj/-/ai/session-1",
    );
    expect(isPublicAIPage(page)).toBe(false);
    expect(isPublicURLPage(page)).toBe(false);
  });

  it("does not treat the AI landing page as public even with a token", () => {
    expect(
      isPublicAIPage(
        pageFor(
          "/[organization]/[project]/-/ai",
          "https://ui.rilldata.com/org/proj/-/ai?token=abc",
        ),
      ),
    ).toBe(false);
  });
});
