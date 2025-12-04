import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get, type Writable } from "svelte/store";
import { CanvasEntity } from "../stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { CanvasResponse } from "../selector";
import {
  getQueryServiceResolveCanvasQueryOptions,
  type V1CanvasSpec,
  type V1MetricsView,
  type V1ResolveCanvasResponse,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";

export type CanvasStore = {
  runtime: Writable<Runtime>;
  canvasEntity: CanvasEntity;
  queryClient: QueryClient;
};

type CanvasId = `${string}::${string}`;

const canvasRegistry = new Map<string, CanvasStore>();

function makeCanvasId(canvasName: string, instanceId: string): CanvasId {
  return `${instanceId}::${canvasName}`;
}

export function getCanvasStoreUnguarded(
  canvasName: string,
  instanceId: string,
): CanvasStore | undefined {
  const id = makeCanvasId(canvasName, instanceId);

  return canvasRegistry.get(id);
}

export function getCanvasStore(
  canvasName: string,
  instanceId: string,
): CanvasStore {
  const id = makeCanvasId(canvasName, instanceId);

  const store = canvasRegistry.get(id);

  if (!store) {
    throw new Error(
      `Canvas store for ID ${id} has not been initialized. Use 'setCanvasStore' first.`,
    );
  }

  return store;
}

export function removeCanvasStore(
  canvasName: string,
  instanceId: string,
): void {
  const id = makeCanvasId(canvasName, instanceId);
  canvasRegistry.delete(id);
}

export function setCanvasStore(
  canvasName: string,
  instanceId: string,
  response: CanvasResponse,
): CanvasStore {
  const id = makeCanvasId(canvasName, instanceId);

  if (canvasRegistry.has(id)) {
    console.warn(
      `Canvas store for ID ${id} already exists. Returning existing store.`,
    );
    return canvasRegistry.get(id)!;
  }

  const canvasEntity = new CanvasEntity(canvasName, instanceId, response);
  const store: CanvasStore = {
    runtime: runtime,
    canvasEntity,
    queryClient,
  };

  canvasRegistry.set(id, store);
  return store;
}

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

const MAX_SPEC_RETRIES = 10;
const RETRY_DELAY_MS = 1000;

export async function handleCanvasStoreInitialization(
  canvasName: string,
  instanceId?: string,
) {
  let store: CanvasStore;

  instanceId = instanceId ?? get(runtime).instanceId;

  try {
    store = getCanvasStore(canvasName, instanceId);
    return {
      store,
      canvasName,
    };
  } catch {
    //
  }

  try {
    const queryOptions = getQueryServiceResolveCanvasQueryOptions(
      instanceId,
      canvasName,
      {},
    );

    const fetchOptions = {
      ...queryOptions,
      retry: 10,
    };

    let data: V1ResolveCanvasResponse | null = null;
    let validSpec: V1CanvasSpec | undefined = undefined;

    for (let attempt = 1; attempt <= MAX_SPEC_RETRIES; attempt++) {
      data = await queryClient.fetchQuery(fetchOptions);

      validSpec = data?.canvas?.canvas?.state?.validSpec;

      if (validSpec) {
        break;
      }

      if (attempt < MAX_SPEC_RETRIES) {
        await delay(RETRY_DELAY_MS);
      } else {
        throw new Error(
          `Canvas spec not ready after ${MAX_SPEC_RETRIES} attempts.`,
        );
      }
    }

    const metricsViews: Record<string, V1MetricsView | undefined> = {};
    const refMetricsViews = data?.referencedMetricsViews;
    if (refMetricsViews) {
      Object.keys(refMetricsViews).forEach((key) => {
        metricsViews[key] = refMetricsViews?.[key]?.metricsView;
      });
    }

    const processed = {
      canvas: data?.canvas?.canvas?.state?.validSpec,
      components: data?.resolvedComponents,
      metricsViews,
      filePath: data?.canvas?.meta?.filePaths?.[0],
    };

    store = setCanvasStore(canvasName, instanceId, processed);

    return {
      store,
      canvasName,
    };
  } catch (e) {
    throw error(
      404,
      `Canvas '${canvasName}' not found or spec not ready in instance '${instanceId}'`,
    );
  }
}
