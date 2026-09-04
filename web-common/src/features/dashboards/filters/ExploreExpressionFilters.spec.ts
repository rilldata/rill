import { useExploreFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/explore-filters-variant";
import {
  testDimensionFilters,
  testMeasureFilters,
  testUrlLengthLimit,
  testURLNavigationFlows,
} from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
import type { HoistedPageForComponentTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import type { ActionResult } from "@sveltejs/kit";
import { describe, vi } from "vitest";

const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
    // superforms, which the measure filter form uses, registers a navigation guard for tainted
    // forms as soon as it is created.
    beforeNavigate: () => {},
  };
});
vi.mock("$app/forms", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$app/forms")>();
  return {
    ...actual,
    // The real `applyAction` needs the SvelteKit client runtime, which a component test does not
    // boot. The measure filter form reads its validation result back from the page, so hand the
    // result to the page mock instead.
    applyAction: (result: ActionResult) => {
      hoistedPage.applyAction(result);
      return Promise.resolve();
    },
  };
});
// The rune based url stores read `page` from here, and SvelteKit only populates it through its
// client router, which a component test does not boot. `PageMockForExploreTests` keeps this in sync
// with the `$app/stores` page below.
vi.mock("$app/state", async () => {
  return {
    page: (
      await import(
        "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
      )
    ).pageStateMock,
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
    // superforms cancels a submit that navigates to another route, so it subscribes to this while
    // the measure filter form is submitting. Nothing here navigates away from the dashboard.
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});

// Tests filter interactions on the explore dashboard, where the filter bar writes to the explore
// state and the url. Loading state from other sources is covered by DashboardStateManager.spec.ts.
//
// The test bodies live in `test/expression-filters-suite.ts`, so that
// StandaloneExpressionFilters.spec.ts can run them against a bar with no dashboard around it. See
// `test/README.md`.
describe("ExploreExpressionFilters", () => {
  const variant = useExploreFiltersVariant(hoistedPage);

  testDimensionFilters(variant);
  testUrlLengthLimit(variant);
  testMeasureFilters(variant);
  testURLNavigationFlows(variant);
});
