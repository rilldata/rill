import type { QueryClient, QueryKey } from "@tanstack/svelte-query";
import { get, type Readable } from "svelte/store";
import { tick } from "svelte";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";

const DEFAULT_TIMEOUT_MS = 60_000;
// How often waitUntilQueriesIdle polls for in-flight queries. Frame-rate polling
// is needlessly frequent; a coarser interval is plenty to detect idleness.
const QUERY_POLL_INTERVAL_MS = 100;

function asyncRequestAnimationFrame(): Promise<void> {
  return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// Resolves once the boolean store equals `target`, or after `timeoutMs`.
// Returns true if the value was reached, false on timeout.
export function waitForStoreValue(
  store: Readable<boolean>,
  target: boolean,
  timeoutMs: number,
): Promise<boolean> {
  if (get(store) === target) return Promise.resolve(true);

  return new Promise((resolve) => {
    let settled = false;
    let shouldUnsubscribeAfterSubscribe = false;
    let unsub = () => {
      shouldUnsubscribeAfterSubscribe = true;
    };
    const finish = (value: boolean) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      unsub();
      resolve(value);
    };
    const timer = setTimeout(() => finish(false), timeoutMs);
    // subscribe() fires synchronously with the current value; guard against it
    // resolving (and calling unsub) before unsub is assigned below.
    unsub = store.subscribe((value) => {
      if (value === target) finish(true);
    });
    if (shouldUnsubscribeAfterSubscribe) unsub();
  });
}

// Resolves once there are no in-flight queries for `stablePolls` consecutive
// polls (debouncing the gaps between dependent queries), or on timeout.
export async function waitUntilQueriesIdle(
  queryClient: QueryClient,
  opts: { instanceId: string; stablePolls?: number; timeoutMs: number },
): Promise<boolean> {
  const stablePolls = opts.stablePolls ?? 2;
  const deadline = Date.now() + opts.timeoutMs;
  let stable = 0;
  while (Date.now() < deadline) {
    const fetching = queryClient.isFetching({
      predicate: (query) =>
        isCanvasExportQuery(query.queryKey, opts.instanceId),
    });
    if (fetching === 0) {
      stable += 1;
      if (stable >= stablePolls) return true;
    } else {
      stable = 0;
    }
    await sleep(QUERY_POLL_INTERVAL_MS);
  }
  return false;
}

export function isCanvasExportQuery(
  queryKey: QueryKey,
  instanceId: string,
): boolean {
  const [service] = queryKey;
  if (service === "metrics_sql") return true;
  if (service !== "QueryService") return false;
  // QueryService keys follow [ServiceName, methodName, instanceId, request]
  // (see runtime-client/invalidation.ts); match the instance at its fixed index
  // rather than scanning the whole key.
  return queryKey[2] === instanceId;
}

// Forces every canvas component to render (bypassing the IntersectionObserver
// lazy-load), then waits for the canvas, its data, and fonts to settle so the
// DOM is ready to be rasterized. Best-effort: returns even on timeout.
//
// Returns a function that restores each component's prior `visible` value.
// `visible` gates component queries, so without this an export would leave every
// below-the-fold component permanently visible, keeping their queries active on
// later filter/time changes. The IntersectionObserver keeps observing the
// restored-hidden components, so lazy-load still kicks in when they scroll in.
export async function prepareCanvasForCapture(
  canvasEntity: CanvasEntity,
  queryClient: QueryClient,
  opts: { instanceId: string; timeoutMs?: number },
): Promise<() => void> {
  const timeoutMs = opts.timeoutMs ?? DEFAULT_TIMEOUT_MS;

  // Force-render all components; idempotent with the IntersectionObserver.
  const previouslyHidden: BaseCanvasComponent[] = [];
  for (const component of canvasEntity.componentsStore.read().values()) {
    if (!get(component.visible)) previouslyHidden.push(component);
    component.visible.set(true);
  }
  const restoreVisibility = () => {
    for (const component of previouslyHidden) component.visible.set(false);
  };

  await tick();
  await asyncRequestAnimationFrame();
  await asyncRequestAnimationFrame();

  await waitForStoreValue(canvasEntity.firstLoad, false, timeoutMs);
  await waitUntilQueriesIdle(queryClient, {
    instanceId: opts.instanceId,
    timeoutMs,
  });

  if (document.fonts) {
    await document.fonts.ready;
  }
  // Give Vega/canvas renderers a couple of frames to flush their final paint.
  await asyncRequestAnimationFrame();
  await asyncRequestAnimationFrame();

  return restoreVisibility;
}
