import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import RuntimeContextHarness from "@rilldata/web-common/features/metrics-views/providers/test/RuntimeContextHarness.svelte";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { QueryClient } from "@tanstack/svelte-query";
import { mount, unmount } from "svelte";

/**
 * Helpers for unit tests that need real metrics view specs rather than hand-stubbed ones.
 *
 *   useMetricsViewMocks({ [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT });
 *
 *   it("filters by publisher", async () => {
 *     const { value: provider, destroy } = await createTestMetricsViewsProvider([
 *       AD_BIDS_METRICS_NAME,
 *     ]);
 *     ...
 *     destroy();
 *   });
 */

export type MetricsViewSpecs = Record<string, V1MetricsViewSpec>;

export type RuntimeTestContext = {
  runtimeClient: RuntimeClient;
  queryClient: QueryClient;
};

export type RenderedInRuntimeContext<T> = RuntimeTestContext & {
  /** Whatever `init` returned. */
  value: T;
  destroy: () => void;
};

/**
 * Serves `specs` from ListResources for the duration of the test file.
 *
 * Call this at module or `describe` scope, since it registers the `beforeAll` hook
 * that stubs `fetch`. The returned mocks can be used to add responses for other
 * endpoints, such as aggregation queries backing dimension value lists.
 */
export function useMetricsViewMocks(specs: MetricsViewSpecs) {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  for (const [metricsViewName, spec] of Object.entries(specs)) {
    mocks.mockMetricsView(metricsViewName, spec);
  }
  return mocks;
}

/**
 * Runs `init` inside an effect root, which is all a class needs when it only calls
 * `$effect` in its constructor, such as ExpressionFilterManager. Classes that also
 * call `createQuery` need a QueryClient context, so use {@link renderInRuntimeContext}
 * for those.
 */
export function createInEffectRoot<T>(init: () => T): {
  value: T;
  destroy: () => void;
} {
  let value: T | undefined;
  const destroy = $effect.root(() => {
    value = init();
  });
  return { value: value as T, destroy };
}

/**
 * Runs `init` inside a mounted component so that it has a QueryClient context and an
 * effect owner. Classes like MetricsViewsProvider call `createQuery` in their
 * constructors, so they can only be built here and not directly from a test body.
 *
 * The caller owns the returned `destroy`; call it once the test is done to unmount
 * the harness and stop the queries.
 */
export function renderInRuntimeContext<T>(
  init: (ctx: RuntimeTestContext) => T,
): RenderedInRuntimeContext<T> {
  const runtimeClient = new RuntimeClient({
    host: "http://localhost",
    instanceId: "test",
  });
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: "always",
      },
    },
  });

  let value: T | undefined;
  const component = mount(RuntimeContextHarness, {
    target: document.createElement("div"),
    props: {
      queryClient,
      runtimeClient,
      init: () => {
        value = init({ runtimeClient, queryClient });
      },
    },
  });

  return {
    value: value as T,
    runtimeClient,
    queryClient,
    destroy: () => {
      void unmount(component);
      queryClient.clear();
      runtimeClient.dispose();
    },
  };
}

/**
 * Creates a MetricsViewsProvider for `metricsViewNames` and resolves once the specs
 * mocked by {@link useMetricsViewMocks} have landed.
 */
export async function createTestMetricsViewsProvider(
  metricsViewNames: string[],
): Promise<RenderedInRuntimeContext<MetricsViewsProvider>> {
  const rendered = renderInRuntimeContext(({ runtimeClient }) => {
    return new MetricsViewsProvider(runtimeClient, metricsViewNames);
  });

  await waitForMetricsViewSpecs(rendered.value, metricsViewNames);

  return rendered;
}

/** Resolves once every requested metrics view has a spec, or throws on timeout. */
export async function waitForMetricsViewSpecs(
  metricsViewsProvider: MetricsViewsProvider,
  metricsViewNames: string[],
  timeout = 5000,
) {
  const loaded = await waitUntil(
    () => metricsViewNames.every((name) => !!metricsViewsProvider.specs[name]),
    timeout,
    10,
  );
  if (!loaded) {
    const missing = metricsViewNames.filter(
      (name) => !metricsViewsProvider.specs[name],
    );
    throw new Error(
      `Timed out waiting for metrics view specs: ${missing.join(", ")}. ` +
        `Did the test call useMetricsViewMocks with these metrics views?`,
    );
  }
}
