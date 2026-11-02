import type { MarkdownCanvasComponent } from "@rilldata/web-common/features/canvas/components/markdown";
import { getResolveTemplatedStringQueryOptions } from "@rilldata/web-common/features/canvas/components/markdown/util";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get, writable } from "svelte/store";
import { describe, expect, it } from "vitest";

describe("getResolveTemplatedStringQueryOptions", () => {
  it("Should release the parent spec once unsubscribed, however often the filters changed", () => {
    const exprByMetricsView = writable<Record<string, V1Expression>>({});

    // Counts the live subscriptions to the parent canvas spec.
    let parentSpecSubscribers = 0;
    const parentSpec = writable({ data: { metricsViews: {} } });
    const parentSpecStore = {
      subscribe: (run: (value: unknown) => void) => {
        parentSpecSubscribers++;
        const unsubscribe = parentSpec.subscribe(run);
        return () => {
          parentSpecSubscribers--;
          unsubscribe();
        };
      },
    };

    const component = {
      id: "canvas--component-0-0",
      specStore: writable({ content: "{{ metrics_sql 'select 1' }}" }),
      resource: writable({ meta: { stateVersion: "1" } }),
      timeAndFilterStore: writable({ timeRange: undefined }),
      parent: {
        specStore: parentSpecStore,
        expressionFilterManager: {
          exprByMetricsViewStore: exprByMetricsView,
        },
        timeManager: { hasTimeSeriesStore: writable(false) },
      },
    } as unknown as MarkdownCanvasComponent;

    const optionsStore = getResolveTemplatedStringQueryOptions(
      component,
      new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
    );

    const unsubscribe = optionsStore.subscribe(() => {});
    // Every filter change re-runs the outer derived, which builds a new inner one.
    exprByMetricsView.set({ mv: { ident: "a" } });
    exprByMetricsView.set({ mv: { ident: "b" } });
    expect(parentSpecSubscribers).toBe(1);

    unsubscribe();
    expect(parentSpecSubscribers).toBe(0);
  });

  it("Should key the query on the resource's state version so a reconcile refetches the resolved text", () => {
    const specStore = writable({ content: "{{ metrics_sql 'select 1' }}" });
    const resource = writable({ meta: { stateVersion: "1" } });
    const component = {
      id: "canvas--component-0-0",
      specStore,
      resource,
      timeAndFilterStore: writable({ timeRange: undefined }),
      parent: {
        specStore: writable({ data: { metricsViews: {} } }),
        expressionFilterManager: {
          exprByMetricsViewStore: writable<Record<string, V1Expression>>({}),
        },
        timeManager: { hasTimeSeriesStore: writable(false) },
      },
    } as unknown as MarkdownCanvasComponent;

    const optionsStore = getResolveTemplatedStringQueryOptions(
      component,
      new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
    );

    const keyBefore = get(optionsStore).queryKey;
    expect(keyBefore).toContain("1");

    // A local edit changes the content before the runtime has reconciled it, so it must not change the key.
    specStore.set({ content: "{{ metrics_sql 'select 2' }}" });
    expect(get(optionsStore).queryKey).toEqual(keyBefore);

    resource.set({ meta: { stateVersion: "2" } });
    const keyAfter = get(optionsStore).queryKey;
    expect(keyAfter).not.toEqual(keyBefore);
    expect(keyAfter).toContain("2");
  });
});
