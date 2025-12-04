import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get, writable, type Writable } from "svelte/store";
import { CanvasEntity } from "../stores/canvas-entity";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { CanvasResponse } from "../selector";
import {
  getQueryServiceResolveCanvasQueryOptions,
  type V1CanvasSpec,
  type V1MetricsView,
  type V1ResolveCanvasResponse,
} from "@rilldata/web-common/runtime-client";

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

const MAX_SPEC_RETRIES = 15;
const PollIntervalWhenDashboardFirstReconciling = 1000;
const PollIntervalWhenDashboardErrored = 5000;

export function handleCanvasStoreInitialization(
  canvasName: string,
  instanceId?: string,
): {
  canvasStoreStore: Writable<CanvasStore | null>;
  reconcilingStore: Writable<boolean>;
  errorMessageStore: Writable<string | undefined>;
} {
  const response = {
    canvasStoreStore: writable<CanvasStore | null>(null),
    reconcilingStore: writable<boolean>(false),
    errorMessageStore: writable<string | undefined>(undefined),
  };

  const initializeStoreAndFetchSpec = async () => {
    let store: CanvasStore;
    const currentInstanceId = instanceId ?? get(runtime).instanceId;

    try {
      store = getCanvasStore(canvasName, currentInstanceId);
      response.canvasStoreStore.set(store);
      return;
    } catch {
      // Store not found
    }

    try {
      const queryOptions = getQueryServiceResolveCanvasQueryOptions(
        currentInstanceId,
        canvasName,
        {},
      );

      const fetchOptions = {
        ...queryOptions,
        retry: 5,
      };

      let data: V1ResolveCanvasResponse | null = null;
      let validSpec: V1CanvasSpec | undefined = undefined;

      for (let attempt = 1; attempt <= MAX_SPEC_RETRIES; attempt++) {
        data = await queryClient.fetchQuery(fetchOptions);

        const resource = data?.canvas;

        validSpec = resource?.canvas?.state?.validSpec;
        const reconcileError = resource?.meta?.reconcileError;

        // Reconciliation is ongoing if no valid spec exists AND there's no error.
        const isReconciling = !validSpec && !reconcileError;

        // Error state is active if no valid spec exists AND a reconcile error is present.
        // We prioritize showing a previous valid spec over an error if it exists.
        const isErrored = !validSpec && !!reconcileError;

        response.errorMessageStore.set(reconcileError);

        if (attempt !== 1) {
          response.reconcilingStore.set(isReconciling);
        }

        if (validSpec) {
          response.reconcilingStore.set(false);
          response.errorMessageStore.set(undefined);
          break;
        }

        const delayLength = isReconciling
          ? PollIntervalWhenDashboardFirstReconciling
          : isErrored
            ? PollIntervalWhenDashboardErrored
            : 1000;

        if (attempt < MAX_SPEC_RETRIES) {
          await delay(delayLength);
        } else {
          throw new Error();
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

      store = setCanvasStore(canvasName, currentInstanceId, processed);

      response.canvasStoreStore.set(store);
    } catch {
      response.errorMessageStore.set(
        `The canvas dashboard you requested could not be found. Please check that you provided the name of a working canvas dashboard.`,
      );
    }
  };

  initializeStoreAndFetchSpec().catch(console.error);

  return response;
}
