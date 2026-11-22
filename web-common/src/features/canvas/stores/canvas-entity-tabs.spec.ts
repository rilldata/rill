import {
  CanvasEntity,
  lastVisitedState,
} from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1CanvasRow } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get, writable, type Writable } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  goto: vi.fn<
    (url: string | URL, opts?: { replaceState?: boolean }) => Promise<void>
  >(),
  pageStore: null as unknown as Writable<{ url: URL }>,
}));

vi.mock("$app/navigation", () => ({
  goto: (url: string | URL, opts?: { replaceState?: boolean }) =>
    hoisted.goto(url, opts),
}));

vi.mock("$app/stores", () => ({
  page: {
    subscribe: (run: (value: { url: URL }) => void) =>
      hoisted.pageStore.subscribe(run),
  },
}));

vi.mock("$app/state", async () => ({
  page: (
    await import(
      "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
    )
  ).pageStateMock,
}));

vi.mock("svelte", async (importOriginal) => {
  const actual = await importOriginal<typeof import("svelte")>();
  return {
    ...actual,
    getContext: (key: unknown) =>
      key === "$$_queryClient" ? queryClient : actual.getContext(key),
  };
});

function makeTabGroupRow(
  name: string,
  tabs: Array<{ name: string; displayName?: string }>,
): V1CanvasRow {
  return {
    tabGroup: {
      name,
      tabs: tabs.map((t) => ({
        name: t.name,
        displayName: t.displayName ?? t.name,
        rows: [],
      })),
    },
  };
}

describe("CanvasEntity tab URL state", () => {
  const createdEntities: CanvasEntity[] = [];

  function setBrowserUrl(pathAndSearch: string) {
    const nextUrl = new URL(pathAndSearch, "http://localhost");
    window.history.replaceState({}, "", nextUrl.pathname + nextUrl.search);
    hoisted.pageStore.set({ url: nextUrl });
    return nextUrl;
  }

  function createEntity(
    rows: V1CanvasRow[],
    { allowUnvalidatedSpec = false, name = "tabbed_canvas" } = {},
  ) {
    const client = new RuntimeClient({
      host: "http://localhost",
      instanceId: "test",
    });
    const entity = new CanvasEntity(
      name,
      "test",
      {
        canvas: { displayName: "Tabbed Canvas", rows },
        components: {},
        metricsViews: {},
        filePath: undefined,
      },
      client,
      allowUnvalidatedSpec,
    );
    createdEntities.push(entity);
    return entity;
  }

  /**
   * Wire `goto` to update `window.location` and re-enter `entity.onUrlChange`,
   * mirroring SvelteKit's navigation + `CanvasInitialization`'s reactive `$page` hook.
   */
  function wireNavigation(entity: CanvasEntity) {
    hoisted.goto.mockImplementation(async (target) => {
      const nextUrl = new URL(
        typeof target === "string" ? target : target.toString(),
        window.location.href,
      );
      window.history.replaceState({}, "", nextUrl.pathname + nextUrl.search);
      hoisted.pageStore.set({ url: nextUrl });
      await entity.onUrlChange({ url: nextUrl });
    });
  }

  async function flushNavigation() {
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
  }

  beforeEach(() => {
    hoisted.pageStore = writable({
      url: new URL("http://localhost/canvas/tabbed_canvas"),
    });
    hoisted.goto.mockReset();
    hoisted.goto.mockResolvedValue(undefined);
    setBrowserUrl("/canvas/tabbed_canvas");
    lastVisitedState.clear();
    queryClient.clear();
  });

  afterEach(() => {
    while (createdEntities.length) {
      createdEntities.pop()?.dispose();
    }
    lastVisitedState.clear();
  });

  it("can switch back to the first tab when no other URL params are present", async () => {
    const entity = createEntity([
      makeTabGroupRow("pages", [{ name: "forms" }, { name: "profiles" }]),
    ]);
    wireNavigation(entity);

    // Initial visit with bare URL: first tab is active.
    await entity.onUrlChange({ url: new URL(window.location.href) });
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(0);

    // Switch to second tab: URL gets `?tabs=pages.profiles`, snapshot is saved.
    entity.setActiveTabInURL("pages", "profiles");
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.profiles");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(1);
    expect(lastVisitedState.get("tabbed_canvas")).toBe("tabs=pages.profiles");

    // Switch back to first tab: must NOT emit an empty query that bounces back via lastVisitedState.
    entity.setActiveTabInURL("pages", "forms");
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.forms");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(0);
    expect(lastVisitedState.get("tabbed_canvas")).toBe("tabs=pages.forms");

    // Returning to the bare URL restores the first tab rather than bouncing to the second tab.
    setBrowserUrl("/canvas/tabbed_canvas");
    await entity.onUrlChange({ url: new URL(window.location.href) });
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.forms");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(0);

    // Switching back to the second tab keeps the URL clean.
    entity.setActiveTabInURL("pages", "profiles");
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.profiles");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(1);
  });

  it("can switch back to the first tab in edit mode without other URL params", async () => {
    setBrowserUrl("/files/dashboards/tabbed_canvas.yaml");
    const entity = createEntity(
      [makeTabGroupRow("pages", [{ name: "forms" }, { name: "profiles" }])],
      { allowUnvalidatedSpec: true },
    );
    wireNavigation(entity);

    await entity.onUrlChange({ url: new URL(window.location.href) });

    entity.setActiveTabInURL("pages", "profiles");
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.profiles");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(1);

    entity.setActiveTabInURL("pages", "forms");
    await flushNavigation();
    expect(window.location.search).toBe("?tabs=pages.forms");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(0);
  });

  it("still omits the tabs param for the first tab when other URL params are present", async () => {
    setBrowserUrl("/canvas/tabbed_canvas?tr=P7D");
    const entity = createEntity([
      makeTabGroupRow("pages", [{ name: "forms" }, { name: "profiles" }]),
    ]);
    wireNavigation(entity);

    await entity.onUrlChange({ url: new URL(window.location.href) });

    entity.setActiveTabInURL("pages", "profiles");
    await flushNavigation();
    expect(window.location.search).toBe("?tr=P7D&tabs=pages.profiles");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(1);

    entity.setActiveTabInURL("pages", "forms");
    await flushNavigation();
    expect(window.location.search).toBe("?tr=P7D");
    expect(get(entity.getTabGroup("pages")!.activeTabIndex)).toBe(0);
  });

  it("handles multiple tab groups without other URL params", async () => {
    const entity = createEntity([
      makeTabGroupRow("g1", [{ name: "a" }, { name: "b" }]),
      makeTabGroupRow("g2", [{ name: "c" }, { name: "d" }]),
    ]);
    wireNavigation(entity);

    await entity.onUrlChange({ url: new URL(window.location.href) });

    // Both groups moved off first tab.
    entity.setActiveTabInURL("g1", "b");
    await flushNavigation();
    entity.setActiveTabInURL("g2", "d");
    await flushNavigation();
    expect(new URLSearchParams(window.location.search).get("tabs")).toBe(
      "g1.b,g2.d",
    );

    // Moving g1 back to first tab keeps URL short (only g2.d).
    entity.setActiveTabInURL("g1", "a");
    await flushNavigation();
    expect(new URLSearchParams(window.location.search).get("tabs")).toBe(
      "g2.d",
    );
    expect(get(entity.getTabGroup("g1")!.activeTabIndex)).toBe(0);
    expect(get(entity.getTabGroup("g2")!.activeTabIndex)).toBe(1);

    // Moving g2 back to first tab leaves no other params: write all active tabs explicitly.
    entity.setActiveTabInURL("g2", "c");
    await flushNavigation();
    expect(new URLSearchParams(window.location.search).get("tabs")).toBe(
      "g1.a,g2.c",
    );
    expect(get(entity.getTabGroup("g1")!.activeTabIndex)).toBe(0);
    expect(get(entity.getTabGroup("g2")!.activeTabIndex)).toBe(0);
  });

  it("encodes dots and commas in group and tab names when falling back to explicit first tab", async () => {
    const entity = createEntity([
      makeTabGroupRow("g.1", [{ name: "a,b" }, { name: "c.d" }]),
    ]);
    wireNavigation(entity);

    await entity.onUrlChange({ url: new URL(window.location.href) });

    entity.setActiveTabInURL("g.1", "c.d");
    await flushNavigation();
    expect(get(entity.getTabGroup("g.1")!.activeTabIndex)).toBe(1);

    entity.setActiveTabInURL("g.1", "a,b");
    await flushNavigation();
    expect(get(entity.getTabGroup("g.1")!.activeTabIndex)).toBe(0);
    expect(new URLSearchParams(window.location.search).get("tabs")).toBe(
      "g%2E1.a%2Cb",
    );
  });
});
