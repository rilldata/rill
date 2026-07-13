import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";

/**
 * Mobile smoke manifest.
 *
 * web-admin serves the real device viewport on every route (see `app.html`).
 * The route ids listed here are the surfaces that were explicitly audited for
 * narrow screens and are exercised below at a phone viewport. When a new
 * surface is made responsive, add its route id here and a matching smoke URL
 * to {@link TEST_ROUTES}; the coverage guard fails otherwise.
 *
 * `/[organization]/[project]/-/ai/[conversationId]` is intentionally not
 * listed yet: its layout is covered by the `-/ai` surface, and the smoke test
 * has no seeded conversation to visit.
 */
const MOBILE_READY_ROUTES: string[] = [
  "/[organization]/[project]",
  "/[organization]/[project]/-/ai",
  "/[organization]/[project]/explore/[dashboard]",
  "/[organization]/[project]/canvas/[dashboard]",
];

/**
 * Maps each mobile-ready SvelteKit route id to a concrete URL in the e2e
 * environment (org `e2e`, projects `openrtb` / `adbids` seeded by `setup`).
 */
const TEST_ROUTES: Record<string, string> = {
  "/[organization]/[project]": "/e2e/openrtb",
  "/[organization]/[project]/-/ai": "/e2e/openrtb/-/ai",
  "/[organization]/[project]/explore/[dashboard]":
    "/e2e/openrtb/explore/auction_explore",
  "/[organization]/[project]/canvas/[dashboard]":
    "/e2e/openrtb/canvas/bids_canvas",
};

/**
 * Asserts that the page produces no horizontal overflow at the current
 * viewport width: neither the document scrolls horizontally, nor does any
 * visible element extend past the right edge of the viewport. Offending
 * element selectors are included in the failure message.
 */
async function assertNoHorizontalOverflow(page: Page) {
  const result = await page.evaluate(() => {
    const viewportWidth = window.innerWidth;
    const doc = document.scrollingElement;
    const documentScrolls = doc ? doc.scrollWidth > doc.clientWidth + 1 : false;

    const describe = (el: Element): string => {
      const tag = el.tagName.toLowerCase();
      const id = el.id ? `#${el.id}` : "";
      const cls =
        typeof el.className === "string" && el.className.trim()
          ? `.${el.className.trim().split(/\s+/).join(".")}`
          : "";
      return `${tag}${id}${cls}`;
    };

    // Collect visible elements whose right edge extends past the viewport.
    // Cap the list so a systemic failure does not produce an unwieldy message.
    const offenders: string[] = [];
    for (const el of Array.from(document.querySelectorAll("*"))) {
      const rect = el.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) continue;
      if (rect.right > viewportWidth + 1) {
        offenders.push(`${describe(el)} (right=${Math.round(rect.right)})`);
        if (offenders.length >= 10) break;
      }
    }

    return {
      viewportWidth,
      scrollWidth: doc?.scrollWidth ?? null,
      clientWidth: doc?.clientWidth ?? null,
      documentScrolls,
      offenders,
    };
  });

  expect(
    result.documentScrolls,
    `Document scrolls horizontally at ${result.viewportWidth}px ` +
      `(scrollWidth=${result.scrollWidth}, clientWidth=${result.clientWidth}).`,
  ).toBe(false);

  expect(
    result.offenders,
    `Elements overflow the ${result.viewportWidth}px viewport:\n` +
      result.offenders.map((o) => `  - ${o}`).join("\n"),
  ).toEqual([]);
}

test.describe("Mobile smoke", () => {
  // Coverage guard: every mobile-ready route must have a smoke URL so it is
  // actually exercised here. Fails loudly when a route is added to
  // MOBILE_READY_ROUTES without a corresponding TEST_ROUTES entry.
  test("every mobile-ready route has a smoke URL", () => {
    const missing = MOBILE_READY_ROUTES.filter(
      (routeId) => !(routeId in TEST_ROUTES),
    );
    expect(
      missing,
      `Routes in MOBILE_READY_ROUTES missing a TEST_ROUTES entry: ` +
        missing.join(", "),
    ).toEqual([]);
  });

  if (MOBILE_READY_ROUTES.length === 0) {
    // Keep the suite green (and explicit) until the first surface is migrated.
    test.skip("no mobile-ready routes yet", () => {});
  }

  for (const routeId of MOBILE_READY_ROUTES) {
    test(`no horizontal overflow: ${routeId}`, async ({ page }) => {
      const url = TEST_ROUTES[routeId];
      test.skip(!url, `No TEST_ROUTES entry for ${routeId}`);
      await page.goto(url);
      await assertNoHorizontalOverflow(page);
    });
  }
});
