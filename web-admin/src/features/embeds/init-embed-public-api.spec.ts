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
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
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

const EMBED_URL = "http://localhost/-/embed";
const EXPLORE_ROUTE = "/[organization]/[project]/-/embed/explore/[name]";
const CANVAS_ROUTE = "/[organization]/[project]/-/embed/canvas/[name]";
const AD_BIDS_CANVAS_NAME = "AdBids_canvas";

describe("initEmbedPublicAPI", () => {
  let harness: EmbedPublicAPIHarness;
  let cleanup: () => void;

  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  const client = new RuntimeClient({
    host: "http://localhost",
    instanceId: "test",
  });

  beforeEach(() => {
    vi.useFakeTimers();

    // initEmbedPublicAPI reads the embed's config (theme mode, navigation) from the
    // EmbedStore singleton, which the embed layout initializes before calling it.
    EmbedStore.init(new URL(`${EMBED_URL}?navigation=true`));

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

  describe("navigateBack / navigateForward", () => {
    it("drives the browser history and returns true", async () => {
      const back = vi
        .spyOn(window.history, "back")
        .mockImplementation(() => {});
      const forward = vi
        .spyOn(window.history, "forward")
        .mockImplementation(() => {});

      expect((await harness.call("navigateBack")).result).toBe(true);
      expect(back).toHaveBeenCalledOnce();

      expect((await harness.call("navigateForward")).result).toBe(true);
      expect(forward).toHaveBeenCalledOnce();

      back.mockRestore();
      forward.mockRestore();
    });
  });

  describe("with navigation disabled", () => {
    // Re-initialize the embed without `navigation=true` and re-register the methods
    // against it, mirroring an embed configured with navigation disabled.
    beforeEach(() => {
      cleanup();
      EmbedStore.init(new URL(EMBED_URL));
      cleanup = initEmbedPublicAPI(client);
    });

    it.each(["navigateBack", "navigateForward", "navigateToDashboard"])(
      "returns a JSON-RPC error from %s",
      async (method) => {
        const gotoCountBefore = harness.gotoCalls.length;

        const response = await harness.call(method, {
          name: AD_BIDS_EXPLORE_NAME,
        });

        expect(response.result).toBeUndefined();
        expect(response.error?.message).toBe(
          "Navigation is disabled for this embed",
        );
        expect(harness.gotoCalls.length).toBe(gotoCountBefore);
      },
    );

    it("still allows state changes on the current dashboard", async () => {
      const response = await harness.call("setState", "foo=bar");

      expect(response.result).toBe(true);
      expect(harness.lastGoto()?.url.search).toBe("?foo=bar");
    });
  });

  // navigateToDashboard resolves the target dashboard's kind through ListResources and then
  // applies the state the same way setValidState does.
  describe("navigateToDashboard", () => {
    beforeEach(() => {
      queryClient.clear();
      sessionStorage.clear();
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
        ...AD_BIDS_EXPLORE_INIT,
        defaultPreset: AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
      });
      mocks.mockListResources([
        {
          meta: {
            name: {
              kind: ResourceKind.MetricsView,
              name: AD_BIDS_METRICS_NAME,
            },
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Explore, name: AD_BIDS_EXPLORE_NAME },
          },
        },
        {
          meta: {
            name: { kind: ResourceKind.Canvas, name: AD_BIDS_CANVAS_NAME },
          },
        },
      ]);
    });

    // The ListResources and buildValidatedExploreUrl fetches resolve on a real setTimeout,
    // so advance fake timers to let the RPC response settle before returning it.
    async function callNavigate(params: unknown) {
      const response = harness.call("navigateToDashboard", params);
      await vi.advanceTimersByTimeAsync(50);
      return response;
    }

    it("navigates to an explore dashboard with the validated state", async () => {
      harness.setRoute(CANVAS_ROUTE, { name: AD_BIDS_CANVAS_NAME });

      const response = await callNavigate({
        name: AD_BIDS_EXPLORE_NAME,
        state: "measures=impressions&dims=publisher",
      });

      expect(response.result).toEqual({
        success: true,
        appliedState: "measures=impressions&dims=publisher",
        errors: [],
      });
      const last = harness.lastGoto();
      expect(last?.url.pathname).toBe(
        `/-/embed/explore/${AD_BIDS_EXPLORE_NAME}`,
      );
      expect(last?.url.search).toBe("?measures=impressions&dims=publisher");
      // Navigating to another dashboard should be undoable via `navigateBack`.
      expect(last?.opts).toEqual({ replaceState: false });
    });

    it("navigates to a canvas dashboard applying the state as-is", async () => {
      harness.setRoute(EXPLORE_ROUTE, { name: AD_BIDS_EXPLORE_NAME });

      const response = await callNavigate({
        name: AD_BIDS_CANVAS_NAME,
        state: "foo=bar",
      });

      expect(response.result).toEqual({
        success: true,
        appliedState: "foo=bar",
        errors: [],
      });
      expect(harness.lastGoto()?.url.pathname).toBe(
        `/-/embed/canvas/${AD_BIDS_CANVAS_NAME}`,
      );
      expect(harness.lastGoto()?.url.search).toBe("?foo=bar");
    });

    it("navigates without any state when state is not given", async () => {
      harness.setRoute(CANVAS_ROUTE, { name: AD_BIDS_CANVAS_NAME });
      // Params of the dashboard being navigated away from should not carry over.
      harness.navigateTo("foo=bar");

      const response = await callNavigate({ name: AD_BIDS_EXPLORE_NAME });

      expect(response.result).toEqual({
        success: true,
        appliedState: "",
        errors: [],
      });
      const last = harness.lastGoto();
      expect(last?.url.pathname).toBe(
        `/-/embed/explore/${AD_BIDS_EXPLORE_NAME}`,
      );
      expect(last?.url.search).toBe("");
    });

    it("does not navigate on validation errors when failOnError is true", async () => {
      harness.setRoute(CANVAS_ROUTE, { name: AD_BIDS_CANVAS_NAME });
      const gotoCountBefore = harness.gotoCalls.length;

      const response = await callNavigate({
        name: AD_BIDS_EXPLORE_NAME,
        state: "measures=does_not_exist",
        failOnError: true,
      });

      expect(response.result).toEqual({
        success: false,
        errors: ['Selected measure: "does_not_exist" is not valid.'],
      });
      expect(harness.gotoCalls.length).toBe(gotoCountBefore);
    });

    it("clears prior embed session storage before navigating to an explore", async () => {
      const sessionKey = getKeyForSessionStore(
        AD_BIDS_EXPLORE_NAME,
        EmbedStorageNamespacePrefix,
        ExploreUrlWebView.Explore,
      );
      // Simulate state left over from an earlier visit to the target explore. Without clearing,
      // handleURLChange would restore it instead of the state we just applied.
      sessionStorage.setItem(sessionKey, "f=publisher+IN+%28%27Google%27%29");

      await callNavigate({ name: AD_BIDS_EXPLORE_NAME, state: "" });

      expect(sessionStorage.getItem(sessionKey)).toBeNull();
    });

    it("returns a JSON-RPC error when the dashboard does not exist", async () => {
      const response = await callNavigate({ name: "does_not_exist" });

      expect(response.result).toBeUndefined();
      expect(response.error?.message).toBe(
        'Dashboard "does_not_exist" not found',
      );
    });

    it("returns a JSON-RPC error when the name is not a resource of a dashboard kind", async () => {
      const response = await callNavigate({ name: AD_BIDS_METRICS_NAME });

      expect(response.error?.message).toBe(
        `Dashboard "${AD_BIDS_METRICS_NAME}" not found`,
      );
    });

    it("returns a JSON-RPC error when params is missing a string name", async () => {
      const notObject = await callNavigate(AD_BIDS_EXPLORE_NAME);
      expect(notObject.error?.message).toBe(
        "Expected params to be an object with a string `name` property",
      );

      const nonStringState = await callNavigate({
        name: AD_BIDS_EXPLORE_NAME,
        state: 123,
      });
      expect(nonStringState.error?.message).toBe(
        "Expected `state` to be a string",
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
