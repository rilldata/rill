import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  createStaleBuildReloadHandler,
  STALE_BUILD_RELOAD_COOLDOWN_MS,
  type StaleBuildReloadDeps,
} from "./reload-on-stale-build";

// The module reads SvelteKit's app state only in `getReloadUrl` and when it registers the real listener.
const appState = vi.hoisted(() => ({
  navigating: { to: null as { url: URL } | null },
  page: {
    url: new URL("https://ui.rilldata.com/acme/sales/explore/revenue?tr=P7D"),
  },
}));
vi.mock("$app/state", () => ({
  updated: {},
  navigating: appState.navigating,
  page: appState.page,
}));

const reloadUrl = new URL(
  "https://ui.rilldata.com/acme/sales/explore/revenue?tr=P7D",
);

function fakeStorage(): StaleBuildReloadDeps["storage"] {
  const values = new Map<string, string>();
  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => void values.set(key, value),
  };
}

function setup(overrides: Partial<StaleBuildReloadDeps> = {}) {
  const deps: StaleBuildReloadDeps = {
    isUpdated: vi.fn(() => Promise.resolve(true)),
    getReloadUrl: () => reloadUrl,
    navigate: vi.fn(),
    storage: fakeStorage(),
    now: () => 1_700_000_000_000,
    ...overrides,
  };
  return { deps, handle: createStaleBuildReloadHandler(deps) };
}

describe("createStaleBuildReloadHandler", () => {
  it("does a full-page navigation once a new build is detected", async () => {
    const { deps, handle } = setup();
    await handle();
    expect(deps.navigate).toHaveBeenCalledTimes(1);
    expect(deps.navigate).toHaveBeenCalledWith(reloadUrl);
  });

  it("leaves the page alone while the running build is still the latest", async () => {
    const isUpdated = vi
      .fn<() => Promise<boolean>>()
      .mockResolvedValueOnce(false)
      .mockResolvedValue(true);
    const { deps, handle } = setup({ isUpdated });
    await handle();
    expect(deps.navigate).not.toHaveBeenCalled();

    // The failure that was not caused by a deploy did not count towards the cooldown.
    await handle();
    expect(deps.navigate).toHaveBeenCalledTimes(1);
  });

  it("leaves the page alone when there is no page to reload", async () => {
    const { deps, handle } = setup({ getReloadUrl: () => null });
    await handle();
    expect(deps.isUpdated).not.toHaveBeenCalled();
    expect(deps.navigate).not.toHaveBeenCalled();
  });

  it("reloads at most once per cooldown, also across the reload itself", async () => {
    const storage = fakeStorage();
    const reloadedAt = 1_700_000_000_000;
    const first = setup({ storage, now: () => reloadedAt });
    await first.handle();
    expect(first.deps.navigate).toHaveBeenCalledTimes(1);

    // The reloaded page creates a new handler but shares the tab's storage.
    const second = setup({
      storage,
      now: () => reloadedAt + STALE_BUILD_RELOAD_COOLDOWN_MS - 1,
    });
    await second.handle();
    expect(second.deps.navigate).not.toHaveBeenCalled();

    const third = setup({
      storage,
      now: () => reloadedAt + STALE_BUILD_RELOAD_COOLDOWN_MS,
    });
    await third.handle();
    expect(third.deps.navigate).toHaveBeenCalledTimes(1);
  });

  it("coalesces failures that arrive while the version check is pending", async () => {
    let finishCheck!: (updated: boolean) => void;
    const { deps, handle } = setup({
      isUpdated: () => new Promise((resolve) => (finishCheck = resolve)),
    });
    const pending = Promise.all([handle(), handle(), handle()]);
    finishCheck(true);
    await pending;
    expect(deps.navigate).toHaveBeenCalledTimes(1);
  });

  it("still reloads once when storage access is denied", async () => {
    const denied: StaleBuildReloadDeps["storage"] = {
      getItem: () => {
        throw new Error("Access is denied for this document");
      },
      setItem: () => {
        throw new Error("Access is denied for this document");
      },
    };
    const { deps, handle } = setup({ storage: denied });
    await handle();
    await handle();
    expect(deps.navigate).toHaveBeenCalledTimes(1);
  });
});

describe("getReloadUrl", () => {
  beforeEach(() => {
    // A fresh module registry gives each test an EmbedStore that has not been initialized.
    vi.resetModules();
    appState.navigating.to = null;
  });

  it("is the current page", async () => {
    const { getReloadUrl } = await import("./reload-on-stale-build");
    expect(getReloadUrl()?.href).toBe(appState.page.url.href);
  });

  it("is the destination of an in-flight navigation", async () => {
    appState.navigating.to = {
      url: new URL("https://ui.rilldata.com/acme/sales/explore/orders"),
    };
    const { getReloadUrl } = await import("./reload-on-stale-build");
    expect(getReloadUrl()?.href).toBe(
      "https://ui.rilldata.com/acme/sales/explore/orders",
    );
  });

  it("is null inside an embed, whose access token a reload would lose", async () => {
    const { EmbedStore } = await import(
      "@rilldata/web-common/features/embeds/embed-store.ts"
    );
    EmbedStore.init(
      new URL(
        "https://ui.rilldata.com/-/embed?instance_id=inst&runtime_host=https%3A%2F%2Fruntime.example.com&access_token=token&resource=revenue",
      ),
    );
    const { getReloadUrl } = await import("./reload-on-stale-build");
    expect(getReloadUrl()).toBeNull();
  });
});
