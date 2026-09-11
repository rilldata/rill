import type { HoistedPageForComponentTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { useExploreTimeFiltersVariant } from "@rilldata/web-common/features/dashboards/time-controls/test/explore-time-filters-variant";
import {
  testComparisonTimeRangeFilters,
  testTimeRangeFilters,
} from "@rilldata/web-common/features/dashboards/time-controls/test/time-filters-suite";
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
// client router, which a component test does not boot. `PageMockForComponentTests` keeps this in
// sync with the `$app/stores` page below.
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

// Tests time filter interactions on the explore dashboard, where the time filter bar writes to the
// explore state and the url. Loading state from other sources is covered by
// DashboardStateManager.spec.ts.
//
// The test bodies live in `test/time-filters-suite.ts`, so that the other dashboards embedding the
// same time filter bar can run them as well.
describe("ExploreTimeFilters", () => {
  const variant = useExploreTimeFiltersVariant(hoistedPage);

  testTimeRangeFilters(variant);
  testComparisonTimeRangeFilters(variant);
});
