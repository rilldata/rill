import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
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
import { getQueryServiceResolveCanvasQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import { render, screen, waitFor } from "@testing-library/svelte";
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
const CANVAS_NAME = "revisited_canvas";
const COMPONENT_NAME = `${CANVAS_NAME}--component-0-0`;
const COMPONENT_TITLE = "Revisited image";
const CANVAS_SPEC: V1CanvasSpec = {
  displayName: "Revisited canvas",
  rows: [
    {
      height: 40,
      heightUnit: "px",
      items: [{ component: COMPONENT_NAME, definedInCanvas: true, width: 12 }],
    },
  ],
};

/**
 * Covers coming back to a canvas after leaving it.
 * The canvas store outlives the page that built it, so the second visit must not be served
 * by what the first visit left behind.
 */
describe("CanvasInitialization", () => {
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
      {
        // Not a markdown component: its templated content query subscribes to the canvas spec
        // for good, which would keep the canvas query cached and hide what this covers.
        [COMPONENT_NAME]: {
          renderer: "image",
          rendererProperties: {
            title: COMPONENT_TITLE,
            url: "https://example.com/logo.png",
          },
          definedInCanvas: true,
        },
      },
    );

    queryClient.clear();
    removeCanvasStore(CANVAS_NAME, INSTANCE_ID);
    lastVisitedState.clear();
  });

  function renderCanvas(runtimeClient: RuntimeClient) {
    return render(CanvasDashboardTest, {
      props: { canvasName: CANVAS_NAME },
      // TODO: we need to make sure every single query uses an explicit queryClient instead of the
      //       global one. Only then we can use a fresh client here.
      context: new Map<string | symbol, unknown>([
        ["$$_queryClient", queryClient],
        [RUNTIME_CONTEXT_KEY, runtimeClient],
      ]),
    });
  }

  async function assertComponentRendered() {
    await waitFor(() =>
      expect(screen.getByText(COMPONENT_TITLE)).toBeInTheDocument(),
    );
    expect(screen.queryByText(/No valid component/)).not.toBeInTheDocument();
  }

  it("Should render the components when revisited after the canvas query was dropped from the cache", async () => {
    const runtimeClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    const { unmount } = renderCanvas(runtimeClient);
    await assertComponentRendered();
    unmount();

    // The query cache drops a query once it has had no observers for `gcTime` (5 minutes),
    // so this is the canvas a user returns to after spending that long elsewhere in the project.
    queryClient.removeQueries({
      queryKey: getQueryServiceResolveCanvasQueryKey(INSTANCE_ID, {
        canvas: CANVAS_NAME,
        unsafe: false,
      }),
    });
    renderCanvas(runtimeClient);

    await assertComponentRendered();
  });

  it("Should load the canvas through the current runtime client when revisited", async () => {
    // Leaving the project disposes its runtime client, and coming back creates a new one.
    const firstClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    const { unmount } = renderCanvas(firstClient);
    await assertComponentRendered();
    unmount();
    firstClient.dispose();

    const secondClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    renderCanvas(secondClient);

    await assertComponentRendered();
    expect(getCanvasStore(CANVAS_NAME, INSTANCE_ID).canvasEntity.client).toBe(
      secondClient,
    );
  });
});
