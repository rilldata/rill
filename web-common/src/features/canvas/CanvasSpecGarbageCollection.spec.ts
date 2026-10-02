import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import CanvasEmbedTest from "@rilldata/web-common/features/canvas/test/CanvasEmbedTest.svelte";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  mockResizeObserverForComponentTesting,
  useDashboardFetchMocksForComponentTests,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { AD_BIDS_METRICS_INIT } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type {
  V1CanvasSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

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
const CANVAS_NAME = "gc_canvas";
const COMPONENT_NAME = `${CANVAS_NAME}--component-0-0`;
const METRICS_VIEW_NAME = "AdBids_metrics";

// An image, since it renders from its own spec alone and issues no query of its own.
const CANVAS_INIT: V1CanvasSpec = {
  displayName: "GC canvas",
  rows: [{ items: [{ component: COMPONENT_NAME, width: 12 }] }],
};
const IMAGE_COMPONENT: V1Resource = {
  meta: { name: { kind: ResourceKind.Component, name: COMPONENT_NAME } },
  component: {
    state: {
      validSpec: {
        renderer: "image",
        rendererProperties: { url: "https://example.com/logo.png" },
      },
    },
  },
};

/**
 * The canvas entity is cached across navigations, so it outlives the pages that render it.
 * Once the last page releases it, its spec query has no observers, and TanStack garbage
 * collects it after `gcTime`. Coming back must still render the components, rather than an
 * empty layout of "No valid component" placeholders.
 */
describe("Canvas spec after garbage collection", () => {
  mockAnimationsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  beforeAll(() => {
    // Canvas components lazy load behind an IntersectionObserver, which jsdom lacks.
    vi.stubGlobal(
      "IntersectionObserver",
      class IntersectionObserver {
        observe() {}
        unobserve() {}
        disconnect() {}
      },
    );
  });

  beforeEach(() => {
    new PageMockForComponentTests(hoistedPage);

    mocks.mockMetricsView(METRICS_VIEW_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockCanvas(
      CANVAS_NAME,
      CANVAS_INIT,
      { [METRICS_VIEW_NAME]: AD_BIDS_METRICS_INIT },
      { [COMPONENT_NAME]: IMAGE_COMPONENT },
    );

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    removeCanvasStore(CANVAS_NAME, INSTANCE_ID);
    lastVisitedState.clear();
  });

  function renderCanvas() {
    return render(CanvasEmbedTest, {
      props: { canvasName: CANVAS_NAME },
      context: new Map<string | symbol, unknown>([
        ["$$_queryClient", queryClient],
        [
          RUNTIME_CONTEXT_KEY,
          new RuntimeClient({
            host: "http://localhost",
            instanceId: INSTANCE_ID,
          }),
        ],
      ]),
    });
  }

  // The canvas entity's spec query, as opposed to the other ResolveCanvas queries on the page.
  function specQueries() {
    return queryClient.getQueryCache().findAll({
      predicate: (q) =>
        q.queryKey[1] === "resolveCanvas" &&
        (q.queryKey[3] as { unsafe?: boolean })?.unsafe === false,
    });
  }

  async function assertComponentRendered() {
    await waitFor(() =>
      expect(
        getCanvasStore(CANVAS_NAME, INSTANCE_ID)
          .canvasEntity.componentsStore.read()
          .has(COMPONENT_NAME),
      ).toBe(true),
    );
    expect(screen.queryByText(/No valid component/)).toBeNull();
  }

  it("Should render the components when returning after the spec query was garbage collected", async () => {
    const { unmount } = renderCanvas();
    await assertComponentRendered();

    unmount();
    // Nothing observes the spec query any more, so this is what `gcTime` does to it.
    const queries = specQueries();
    expect(queries).toHaveLength(1);
    for (const query of queries) {
      expect(query.getObserversCount()).toBe(0);
      queryClient.getQueryCache().remove(query);
    }

    // The same cached entity is reused on the way back.
    renderCanvas();

    await assertComponentRendered();
  });
});
