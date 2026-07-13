// @vitest-environment jsdom
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getKeyForSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store.ts";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import initEmbedPublicAPI from "./init-embed-public-api";
import { EmbedStorageNamespacePrefix } from "./constants.ts";
import {
  EmbedPublicAPIHarness,
  type HoistedEmbedPage,
} from "./test/EmbedPublicAPIHarness";

// Mirrors STATE_CHANGE_THROTTLE_TIMEOUT in init-embed-public-api.ts (not exported).
const STATE_CHANGE_THROTTLE_TIMEOUT = 200;

// vi.hoisted runs before the mock factories and before init-embed-public-api is imported.
const { hoistedPage } = vi.hoisted(() => ({
  hoistedPage: {} as HoistedEmbedPage,
}));

// The SvelteKit page store and navigation cannot be reached through `window`,
// so they remain module-level mocks. Everything RPC-related goes through the
// real transport and is observed via `window.parent.postMessage` in the harness.
vi.mock("$app/navigation", () => ({
  goto: (url: URL, opts?: { replaceState?: boolean }) =>
    hoistedPage.goto(url, opts),
}));
vi.mock("$app/stores", () => ({
  page: hoistedPage,
}));

// theme-control reads window.matchMedia at import time and neither theme module is
// relevant to setState/stateChange, so stub them to keep the harness isolated.
vi.mock("@rilldata/web-common/features/themes/theme-control", () => ({
  themeControl: {
    set: { light: vi.fn(), dark: vi.fn(), system: vi.fn() },
    preference: { subscribe: () => () => {} },
  },
}));
vi.mock("@rilldata/web-common/features/embeds/embed-theme", () => ({
  getEmbedThemeStoreInstance: () => ({
    subscribe: () => () => {},
    set: vi.fn(),
  }),
}));

describe("initEmbedPublicAPI", () => {
  let harness: EmbedPublicAPIHarness;
  let cleanup: () => void;

  const client = new RuntimeClient({
    host: "http://localhost",
    instanceId: "test",
  });

  beforeEach(() => {
    vi.useFakeTimers();

    // Construct the harness (mocks window.parent, installs the RPC handler)
    // before init so the "ready" and initial notifications are captured.
    harness = new EmbedPublicAPIHarness(hoistedPage);
    cleanup = initEmbedPublicAPI(client);
  });

  afterEach(() => {
    cleanup();
    harness.destroy();
    vi.useRealTimers();
  });

  it("emits the ready notification during init", () => {
    expect(harness.notifications("ready")).toHaveLength(1);
  });

  describe("setState", () => {
    it("replaces the url search via replaceState and returns true", async () => {
      const response = await harness.call("setState", "a=1&b=2");

      expect(response.result).toBe(true);

      const last = harness.lastGoto();
      expect(last?.url.pathname).toBe("/-/embed");
      expect(last?.url.search).toBe("?a=1&b=2");
      expect(last?.opts).toEqual({ replaceState: true });
    });

    it("replaces existing search params rather than merging them", async () => {
      harness.navigateTo("existing=value");

      await harness.call("setState", "a=1");

      expect(harness.lastGoto()?.url.search).toBe("?a=1");
    });

    it("clears the search when given an empty string", async () => {
      harness.navigateTo("existing=value");

      await harness.call("setState", "");

      expect(harness.lastGoto()?.url.search).toBe("");
    });

    it("returns a JSON-RPC error when state is not a string", async () => {
      const response = await harness.call("setState", 123);

      expect(response.result).toBeUndefined();
      expect(response.error?.message).toBe("Expected state to be a string");
    });
  });

  // setValidState runs the real buildValidatedExploreUrl for explore routes, so the
  // runtime GetExplore/metrics fetches are mocked with the AD_BIDS fixtures. See
  // DashboardStateManager.spec.ts for the same API mocking approach.
  describe("setValidState", () => {
    const EXPLORE_ROUTE = "/[organization]/[project]/-/embed/explore/[name]";
    const CANVAS_ROUTE = "/[organization]/[project]/-/embed/canvas/[name]";

    const mocks = DashboardFetchMocks.useDashboardFetchMocks();

    beforeEach(() => {
      queryClient.clear();
      sessionStorage.clear();
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
      });
    });

    function onExploreRoute() {
      harness.setRoute(EXPLORE_ROUTE, { name: AD_BIDS_EXPLORE_NAME });
    }

    // buildValidatedExploreUrl awaits fetches that resolve on a real setTimeout,
    // so advance fake timers to let the RPC response settle before returning it.
    async function callSetValidState(params: unknown) {
      const response = harness.call("setValidState", params);
      await vi.advanceTimersByTimeAsync(50);
      return response;
    }

    it("applies state as-is on non-explore routes without validating", async () => {
      harness.setRoute(CANVAS_ROUTE, { name: "my_canvas" });

      const response = await callSetValidState({ state: "foo=bar" });

      expect(response.result).toEqual({
        success: true,
        appliedState: "foo=bar",
        errors: [],
      });
      expect(harness.lastGoto()?.url.search).toBe("?foo=bar");
      expect(harness.lastGoto()?.opts).toEqual({ replaceState: true });
    });

    it("applies the validated, canonicalized url on explore routes", async () => {
      onExploreRoute();

      const response = await callSetValidState({
        state: "measures=impressions&dims=publisher",
      });

      expect(response.result).toEqual({
        success: true,
        appliedState: "measures=impressions&dims=publisher",
        errors: [],
      });
      // Params matching rill defaults (sort etc.) are stripped, leaving the canonical url.
      expect(harness.lastGoto()?.url.search).toBe(
        "?measures=impressions&dims=publisher",
      );
      expect(harness.lastGoto()?.opts).toEqual({ replaceState: true });
    });

    it("does not navigate on validation errors when failOnError is true", async () => {
      onExploreRoute();
      const gotoCountBefore = harness.gotoCalls.length;

      const response = await callSetValidState({
        state: "measures=does_not_exist",
        failOnError: true,
      });

      expect(response.result).toEqual({
        success: false,
        errors: ['Selected measure: "does_not_exist" is not valid.'],
      });
      // No navigation happened.
      expect(harness.gotoCalls.length).toBe(gotoCountBefore);
    });

    it("applies the cleaned url despite errors when failOnError defaults to false", async () => {
      onExploreRoute();

      const response = await callSetValidState({
        state: "dims=publisher&measures=does_not_exist",
      });

      // The invalid measure is dropped and its error reported, but the valid
      // dimension survives and the cleaned url is applied.
      expect(response.result).toEqual({
        success: true,
        appliedState: "dims=publisher",
        errors: ['Selected measure: "does_not_exist" is not valid.'],
      });
      expect(harness.lastGoto()?.url.search).toBe("?dims=publisher");
    });

    it("clears prior embed session storage before applying the validated state", async () => {
      onExploreRoute();

      const sessionKey = getKeyForSessionStore(
        AD_BIDS_EXPLORE_NAME,
        EmbedStorageNamespacePrefix,
        ExploreUrlWebView.Explore,
      );
      // Simulate a filter left over from a prior interaction. Without clearing, applying an
      // empty / view-only url lets handleURLChange restore this stale state instead of the
      // validated state we just returned in `appliedState`.
      sessionStorage.setItem(sessionKey, "f=publisher+IN+%28%27Google%27%29");

      await callSetValidState({ state: "" });

      expect(sessionStorage.getItem(sessionKey)).toBeNull();
    });

    it("returns a JSON-RPC error when params is not an object with a string state", async () => {
      onExploreRoute();

      const notObject = await callSetValidState("foo=bar");
      expect(notObject.error?.message).toBe(
        "Expected params to be an object with a string `state` property",
      );

      const missingState = await callSetValidState({});
      expect(missingState.error?.message).toBe(
        "Expected params to be an object with a string `state` property",
      );
    });
  });

  describe("stateChange notification", () => {
    // The page.subscribe callback fires immediately on subscribe during init,
    // which schedules the first (throttled) emission.
    // Flush and clear it so each test starts from a clean slate.
    function flushInitialEmission() {
      vi.advanceTimersByTime(STATE_CHANGE_THROTTLE_TIMEOUT);
      harness.clearMessages();
    }

    it("emits stateChange with embed params stripped when the url changes", () => {
      flushInitialEmission();

      harness.navigateTo(
        "view=explore&instance_id=abc&access_token=xyz&foo=bar",
      );
      vi.advanceTimersByTime(STATE_CHANGE_THROTTLE_TIMEOUT);

      const events = harness.notifications("stateChange");
      expect(events).toHaveLength(1);
      expect(events[0].params).toEqual({ state: "view=explore&foo=bar" });
    });

    it("coalesces rapid url changes into a single emission with the latest state", () => {
      flushInitialEmission();

      harness.navigateTo("step=1");
      harness.navigateTo("step=2");
      harness.navigateTo("step=3");
      // Only one timer is scheduled for the burst.
      vi.advanceTimersByTime(STATE_CHANGE_THROTTLE_TIMEOUT);

      const events = harness.notifications("stateChange");
      expect(events).toHaveLength(1);
      expect(events[0].params).toEqual({ state: "step=3" });
    });

    it("does not emit until the throttle window elapses", () => {
      flushInitialEmission();

      harness.navigateTo("foo=bar");
      vi.advanceTimersByTime(STATE_CHANGE_THROTTLE_TIMEOUT - 1);
      expect(harness.notifications("stateChange")).toHaveLength(0);

      vi.advanceTimersByTime(1);
      expect(harness.notifications("stateChange")).toHaveLength(1);
    });

    it("stops emitting stateChange after cleanup", () => {
      flushInitialEmission();
      cleanup();

      harness.navigateTo("foo=bar");
      vi.advanceTimersByTime(STATE_CHANGE_THROTTLE_TIMEOUT);

      expect(harness.notifications("stateChange")).toHaveLength(0);
    });
  });
});
