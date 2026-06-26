import type { QueryClient, QueryKey } from "@tanstack/svelte-query";
import { get, type Readable } from "svelte/store";
import { tick } from "svelte";
import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

const DEFAULT_TIMEOUT_MS = 60_000;

function raf(): Promise<void> {
  return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}

// Resolves once `predicate` holds for the store's value, or after `timeoutMs`.
// Returns true if the predicate was satisfied, false on timeout.
export function waitForStore<T>(
  store: Readable<T>,
  predicate: (value: T) => boolean,
  timeoutMs: number,
): Promise<boolean> {
  if (predicate(get(store))) return Promise.resolve(true);

  return new Promise((resolve) => {
    let settled = false;
    const finish = (value: boolean) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      unsub();
      resolve(value);
    };
    const timer = setTimeout(() => finish(false), timeoutMs);
    const unsub = store.subscribe((value) => {
      if (predicate(value)) finish(true);
    });
  });
}

// Resolves once there are no in-flight queries for `stableFrames` consecutive
// animation frames (debouncing the gaps between dependent queries), or on timeout.
export async function waitUntilQueriesIdle(
  queryClient: QueryClient,
  opts: { instanceId: string; stableFrames?: number; timeoutMs: number },
): Promise<boolean> {
  const stableFrames = opts.stableFrames ?? 2;
  const deadline = Date.now() + opts.timeoutMs;
  let stable = 0;
  while (Date.now() < deadline) {
    const fetching = queryClient.isFetching({
      predicate: (query) =>
        isCanvasExportQuery(query.queryKey, opts.instanceId),
    });
    if (fetching === 0) {
      stable += 1;
      if (stable >= stableFrames) return true;
    } else {
      stable = 0;
    }
    await raf();
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
  return queryKey.includes(instanceId);
}

// Forces every canvas component to render (bypassing the IntersectionObserver
// lazy-load), then waits for the canvas, its data, and fonts to settle so the
// DOM is ready to be rasterized. Best-effort: returns even on timeout.
export async function prepareCanvasForCapture(
  canvasEntity: CanvasEntity,
  queryClient: QueryClient,
  opts: { instanceId: string; timeoutMs?: number },
): Promise<void> {
  const timeoutMs = opts.timeoutMs ?? DEFAULT_TIMEOUT_MS;

  // Force-render all components; idempotent with the IntersectionObserver.
  for (const component of canvasEntity.componentsStore.read().values()) {
    component.visible.set(true);
  }

  await tick();
  await raf();
  await raf();

  await waitForStore(canvasEntity.firstLoad, (v) => v === false, timeoutMs);
  await waitUntilQueriesIdle(queryClient, {
    instanceId: opts.instanceId,
    timeoutMs,
  });

  if (document.fonts) {
    await document.fonts.ready;
  }
  // Give Vega/canvas renderers a couple of frames to flush their final paint.
  await raf();
  await raf();
}
