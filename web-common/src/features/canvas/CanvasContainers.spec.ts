import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import CanvasDashboardTest from "@rilldata/web-common/features/canvas/test/CanvasDashboardTest.svelte";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { mockResizeObserverForComponentTesting } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
    beforeNavigate: () => {},
  };
});
// The rune based url stores read `page` from here, and SvelteKit only populates it through its
// client router, which a component test does not boot.
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
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});

const INSTANCE_ID = "test";
const CANVAS_NAME = "story_canvas";
const HEADER = `${CANVAS_NAME}--component-0-0`;
// The parser names inline components inside the container at index 1 with the "c1-" prefix.
const INSIDE = `${CANVAS_NAME}--component-c1-1-0`;
const CANVAS_SPEC: V1CanvasSpec = {
  displayName: "Story canvas",
  rows: [
    {
      height: 40,
      heightUnit: "px",
      items: [{ component: HEADER, definedInCanvas: true, width: 12 }],
    },
    {
      container: {
        name: "revenue_story",
        rows: [
          { height: 40, heightUnit: "px", items: [] },
          {
            height: 40,
            heightUnit: "px",
            items: [{ component: INSIDE, definedInCanvas: true, width: 12 }],
          },
        ],
      },
    },
    // An empty container, as viewers get it when nothing prunes the canvas.
    { container: { name: "container-2" } },
  ],
};

function image(title: string) {
  return {
    renderer: "image",
    rendererProperties: { title, url: "https://example.com/logo.png" },
    definedInCanvas: true,
  };
}

describe("canvas containers in view mode", () => {
  mockResizeObserverForComponentTesting();
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  beforeEach(() => {
    new PageMockForComponentTests(hoistedPage);
    // Components latch their lazy load on an IntersectionObserver, which jsdom does not have.
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        observe() {}
        unobserve() {}
        disconnect() {}
      },
    );
    mocks.mockCanvas(
      CANVAS_NAME,
      CANVAS_SPEC,
      {},
      { [HEADER]: image("Header image"), [INSIDE]: image("Story image") },
    );
    queryClient.clear();
    removeCanvasStore(CANVAS_NAME, INSTANCE_ID);
  });

  it("renders a container's rows like free rows, with position-keyed instance ids", async () => {
    const runtimeClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    render(CanvasDashboardTest, {
      props: { canvasName: CANVAS_NAME },
      context: new Map<string | symbol, unknown>([
        ["$$_queryClient", queryClient],
        [RUNTIME_CONTEXT_KEY, runtimeClient],
      ]),
    });

    await waitFor(() =>
      expect(screen.getByText("Story image")).toBeInTheDocument(),
    );
    expect(screen.getByText("Header image")).toBeInTheDocument();
    // Viewers get no container chrome.
    expect(screen.queryByText("Container")).not.toBeInTheDocument();
    expect(screen.queryByText(/No valid component/)).not.toBeInTheDocument();

    const { canvasEntity } = getCanvasStore(CANVAS_NAME, INSTANCE_ID);
    const component = canvasEntity.componentsStore.getNonReactive(INSIDE);
    expect(component?.pathInYAML).toEqual([
      "rows",
      1,
      "rows",
      1,
      "items",
      0,
      "image",
    ]);

    const layout = get(canvasEntity.layout);
    expect(layout.map((block) => block.kind)).toEqual([
      "row",
      "container",
      "container",
    ]);
    const story = layout[1];
    expect(story.kind === "container" && story.container.name).toBe(
      "revenue_story",
    );
    expect(story.kind === "container" && story.container.rowsPath).toEqual([
      "rows",
      1,
      "rows",
    ]);
  });
});
