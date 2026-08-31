import { testCommonFilterFlows } from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
import { useStandaloneFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/standalone-filters-variant";
import type { HoistedPageForExploreTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import type { ActionResult } from "@sveltejs/kit";
import { describe, vi } from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
// Everything else lives in the variant and the shared suite.
const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

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

// Tests the filter bar rendered on its own, with the filter manager as the only source of state.
// The explore variant of the same flows lives in ExploreExpressionFilters.spec.ts.
describe("StandaloneExpressionFilters", () => {
  testCommonFilterFlows(useStandaloneFiltersVariant(hoistedPage));
});
