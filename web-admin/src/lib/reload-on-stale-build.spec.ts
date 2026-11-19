import type { BeforeNavigate } from "@sveltejs/kit";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const appState = vi.hoisted(() => ({
  updated: { current: false, check: vi.fn<() => Promise<boolean>>() },
  navigating: { to: null as { url: URL } | null },
}));
vi.mock("$app/state", () => appState);

const pageUrl = "https://ui.rilldata.com/acme/sales/explore/revenue?tr=P7D";
const otherPageUrl = "https://ui.rilldata.com/acme/sales/explore/orders";
const embedUrl =
  "https://ui.rilldata.com/-/embed?instance_id=inst&runtime_host=https%3A%2F%2Fruntime.example.com&access_token=token&resource=revenue";
const startTime = 1_700_000_000_000;

let reloadOnStaleBuild: typeof import("./reload-on-stale-build");
/** The full page loads the tab started: the assigned URL, or "reload". */
let pageLoads: string[];
let storedValues: Map<string, string>;

// The unit tests run in Node, which has no `window`, `location` or `sessionStorage`, so each test stubs them.
function stubLocation(href: string) {
  vi.stubGlobal("location", {
    get href() {
      return href;
    },
    set href(url: string) {
      pageLoads.push(url);
    },
    reload: () => pageLoads.push("reload"),
  });
}

/** A page of the tab: a fresh `window` with the listener registered. */
function openPage() {
  vi.stubGlobal("window", new EventTarget());
  reloadOnStaleBuild.registerStaleBuildReload();
}

/** Dispatches what Vite dispatches when a dynamic import fails. */
function dispatchPreloadError() {
  window.dispatchEvent(new Event("vite:preloadError"));
}

/** Lets the listener finish its version check. */
function settle() {
  return new Promise((resolve) => setTimeout(resolve));
}

async function failImport() {
  dispatchPreloadError();
  await settle();
}

async function initEmbed() {
  const { EmbedStore } = await import(
    "@rilldata/web-common/features/embeds/embed-store.ts"
  );
  EmbedStore.init(new URL(embedUrl));
}

beforeEach(async () => {
  // A fresh module registry gives each test an EmbedStore that has not been initialized.
  vi.resetModules();
  reloadOnStaleBuild = await import("./reload-on-stale-build");

  appState.updated.current = false;
  appState.updated.check.mockReset().mockResolvedValue(true);
  appState.navigating.to = null;

  pageLoads = [];
  storedValues = new Map();
  stubLocation(pageUrl);
  vi.stubGlobal("sessionStorage", {
    getItem: (key: string) => storedValues.get(key) ?? null,
    setItem: (key: string, value: string) => void storedValues.set(key, value),
  });
  vi.setSystemTime(startTime);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("registerStaleBuildReload", () => {
  it("reloads the current page once a new build is detected", async () => {
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);
  });

  it("reloads the current page when its URL has a fragment, where assigning the URL would only scroll", async () => {
    stubLocation(`${pageUrl}#chart`);
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);
  });

  it("loads the destination of an in-flight navigation", async () => {
    appState.navigating.to = { url: new URL(otherPageUrl) };
    openPage();
    await failImport();
    expect(pageLoads).toEqual([otherPageUrl]);
  });

  it("reloads when the in-flight destination is the current page apart from its fragment", async () => {
    appState.navigating.to = { url: new URL(`${pageUrl}#chart`) };
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);
  });

  it("leaves the page alone while the running build is still the latest", async () => {
    appState.updated.check.mockResolvedValueOnce(false);
    openPage();
    await failImport();
    expect(pageLoads).toEqual([]);

    // The failure that was not caused by a deploy did not count towards the cooldown.
    await failImport();
    expect(pageLoads).toEqual(["reload"]);
  });

  it("leaves embeds alone, whose access token a reload would lose", async () => {
    await initEmbed();
    openPage();
    await failImport();
    expect(appState.updated.check).not.toHaveBeenCalled();
    expect(pageLoads).toEqual([]);
  });

  it("reloads at most once per cooldown, also across the reload itself", async () => {
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);

    // The reloaded page registers its own listener but shares the tab's storage.
    vi.setSystemTime(
      startTime + reloadOnStaleBuild.STALE_BUILD_RELOAD_COOLDOWN_MS - 1,
    );
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);

    vi.setSystemTime(
      startTime + reloadOnStaleBuild.STALE_BUILD_RELOAD_COOLDOWN_MS,
    );
    openPage();
    await failImport();
    expect(pageLoads).toEqual(["reload", "reload"]);
  });

  it("coalesces failures that arrive while the version check is pending", async () => {
    let finishCheck!: (updated: boolean) => void;
    appState.updated.check.mockReturnValue(
      new Promise((resolve) => (finishCheck = resolve)),
    );
    openPage();
    dispatchPreloadError();
    dispatchPreloadError();
    dispatchPreloadError();
    finishCheck(true);
    await settle();
    expect(appState.updated.check).toHaveBeenCalledTimes(1);
    expect(pageLoads).toEqual(["reload"]);
  });

  it("still reloads once when storage access is denied", async () => {
    const denied = () => {
      throw new Error("Access is denied for this document");
    };
    vi.stubGlobal("sessionStorage", { getItem: denied, setItem: denied });
    openPage();
    await failImport();
    await failImport();
    expect(pageLoads).toEqual(["reload"]);
  });

  it("stops listening once removed", async () => {
    vi.stubGlobal("window", new EventTarget());
    const remove = reloadOnStaleBuild.registerStaleBuildReload();
    remove();
    await failImport();
    expect(appState.updated.check).not.toHaveBeenCalled();
  });
});

describe("navigateToNewBuild", () => {
  function navigation(to: string, willUnload = false) {
    return { to: { url: new URL(to) }, willUnload } as BeforeNavigate;
  }

  it("turns the navigation into a full page load once polling has found a newer build", () => {
    appState.updated.current = true;
    reloadOnStaleBuild.navigateToNewBuild(navigation(otherPageUrl));
    expect(pageLoads).toEqual([otherPageUrl]);
  });

  it("reloads on a back or forward navigation, where the location already is the destination", () => {
    appState.updated.current = true;
    stubLocation(`${otherPageUrl}#chart`);
    reloadOnStaleBuild.navigateToNewBuild(navigation(`${otherPageUrl}#chart`));
    expect(pageLoads).toEqual(["reload"]);
  });

  it("leaves the navigation alone while the running build is still the latest", () => {
    reloadOnStaleBuild.navigateToNewBuild(navigation(otherPageUrl));
    expect(pageLoads).toEqual([]);
  });

  it("leaves navigations that unload the page alone", () => {
    appState.updated.current = true;
    reloadOnStaleBuild.navigateToNewBuild(
      navigation("https://docs.rilldata.com", true),
    );
    expect(pageLoads).toEqual([]);
  });

  it("leaves embeds alone, whose access token a full page load would lose", async () => {
    appState.updated.current = true;
    await initEmbed();
    reloadOnStaleBuild.navigateToNewBuild(navigation(otherPageUrl));
    expect(pageLoads).toEqual([]);
  });
});
