<script lang="ts">
  import {
    getCanvasStoreUnguarded,
    setCanvasStore,
    type CanvasStore,
  } from "./state-managers/state-managers";
  import { page } from "$app/stores";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { onNavigate } from "$app/navigation";
  import { writable } from "svelte/store";
  import {
    type V1MetricsView,
    type V1ResolveCanvasResponse,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQueryServiceResolveCanvas } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";
  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;

  export let canvasName: string;
  export let instanceId: string;
  export let showBanner = false;
  export let projectId: string | undefined = undefined;
  export let allowUnvalidatedSpec = false;
  // When set, relative time ranges are anchored at this time instead of now/latest
  // (e.g. for scheduled report exports). Applied when the store is resolved,
  // before any component subscribes to the time interval.
  export let executionTime: string | undefined = undefined;
  // Isolated consumers (e.g. the scheduled report dialog and export page) apply canvas
  // state without page-URL side effects: no redirects that would rewrite the host page's
  // URL and no last-visited snapshot. Implied when urlStateOverride is set.
  export let isolated = false;
  // Canvas state (a URL search string) to apply instead of the page URL.
  // Used by isolated consumers that render a canvas with stored state (e.g. a report's
  // captured filters) on a page whose own URL does not carry canvas state.
  export let urlStateOverride: string | undefined = undefined;

  const client = useRuntimeClient();

  let resolvedStore: CanvasStore | undefined = undefined;

  $: ({ url } = $page);

  $: existingStore = getCanvasStoreUnguarded(
    canvasName,
    instanceId,
    allowUnvalidatedSpec,
  );

  $: fetchedCanvasQuery = !existingStore
    ? createQueryServiceResolveCanvas(
        client,
        { canvas: canvasName, unsafe: allowUnvalidatedSpec },
        {
          query: {
            retry: 5,
            refetchInterval: (query) => {
              const resource = query?.state?.data;
              if (!resource) return false;
              if (isCanvasReconcilingForFirstTime(resource))
                return PollIntervalWhenDashboardFirstReconciling;
              if (isCanvasErrored(resource))
                return PollIntervalWhenDashboardErrored;
              return false;
            },
          },
        },
      )
    : undefined;

  $: isLoading = fetchedCanvasQuery ? $fetchedCanvasQuery?.isLoading : false;

  $: fetchedCanvas = fetchedCanvasQuery ? $fetchedCanvasQuery?.data : undefined;

  $: validSpec = fetchedCanvas?.canvas?.canvas?.state?.validSpec;
  $: reconcileError = fetchedCanvas?.canvas?.meta?.reconcileError;

  $: isReconciling =
    !existingStore && !validSpec && !reconcileError && !isLoading;

  $: reconcileErrorMessage = !validSpec ? reconcileError : undefined;

  $: resolvedStore = getResolvedStore(
    fetchedCanvas,
    isReconciling,
    existingStore,
    instanceId,
  );

  $: ready = !!resolvedStore;

  $: effectiveUrl =
    urlStateOverride !== undefined
      ? new URL(
          `${url.pathname}${urlStateOverride ? `?${urlStateOverride}` : ""}`,
          url.origin,
        )
      : url;

  $: if (resolvedStore) {
    resolvedStore.canvasEntity
      .onUrlChange({
        url: effectiveUrl,
        projectId,
        isolated: isolated || urlStateOverride !== undefined,
      })
      .catch(console.error);
  }

  $: title = resolvedStore?.canvasEntity.titleStore || writable("");
  $: canvasTitle = $title;

  $: bannerStore = resolvedStore?.canvasEntity.bannerStore || writable("");
  $: banner = $bannerStore;

  $: hasBanner = !!banner;

  $: if (hasBanner && showBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: banner ?? "",
        iconType: "alert",
      },
    });
  }

  let release: (() => void) | undefined = undefined;

  onNavigate(() => {
    if (hasBanner) {
      eventBus.emit("remove-banner", DashboardBannerID);
    }
  });

  function isCanvasReconcilingForFirstTime(
    canvasResource: V1ResolveCanvasResponse,
  ) {
    if (!canvasResource) return undefined;
    const isCanvasReconcilingForFirstTime =
      !canvasResource.canvas?.canvas?.state?.validSpec &&
      !canvasResource?.canvas?.meta?.reconcileError;
    return isCanvasReconcilingForFirstTime;
  }

  function isCanvasErrored(canvasResource: V1ResolveCanvasResponse) {
    if (!canvasResource) return undefined;
    // We only consider a dashboard errored (from the end-user perspective) when BOTH a reconcile error exists AND a validSpec does not exist.
    // If there's any validSpec (which can persist from a previous, non-current spec), then we serve that version of the dashboard to the user,
    // so the user does not see an error state.
    const isCanvasErrored =
      !canvasResource.canvas?.canvas?.state?.validSpec &&
      !!canvasResource?.canvas?.meta?.reconcileError;
    return isCanvasErrored;
  }

  function getResolvedStore(
    fetchedCanvas: V1ResolveCanvasResponse | undefined,
    isReconciling: boolean,
    existingStore: CanvasStore | undefined,
    instanceId: string,
  ) {
    // Always set (not just when defined): the store is cached and shared across surfaces,
    // so an undefined executionTime must reset an anchor left behind by a previous consumer
    // (e.g. the report export page), or the live dashboard would keep showing stale data.
    existingStore?.canvasEntity.timeManager.executionTimeStore.set(
      executionTime,
    );
    if (fetchedCanvas && !isReconciling) {
      const metricsViews: Record<string, V1MetricsView | undefined> = {};
      const refMetricsViews = fetchedCanvas?.referencedMetricsViews;
      if (refMetricsViews) {
        Object.keys(refMetricsViews).forEach((key) => {
          metricsViews[key] = refMetricsViews?.[key]?.metricsView;
        });
      }

      const canvasSpec =
        fetchedCanvas?.canvas?.canvas?.state?.validSpec ??
        (allowUnvalidatedSpec
          ? fetchedCanvas?.canvas?.canvas?.spec
          : undefined);

      if (canvasSpec) {
        const processed = {
          canvas: canvasSpec,
          components: fetchedCanvas?.resolvedComponents,
          metricsViews,
          filePath: fetchedCanvas?.canvas?.meta?.filePaths?.[0],
        };

        const newStore = setCanvasStore(
          canvasName,
          instanceId,
          processed,
          client,
          allowUnvalidatedSpec,
        );
        newStore.canvasEntity.timeManager.executionTimeStore.set(executionTime);
        newStore.canvasEntity.acquire();
        release?.(); // release our reference to the previous entity, if any
        release = newStore.canvasEntity.release;
        return newStore;
      }
    }

    existingStore?.canvasEntity?.acquire();
    release?.(); // release our reference to the previous entity, if any
    release = existingStore?.canvasEntity?.release;

    return existingStore;
  }

  onDestroy(() => {
    release?.();
  });
</script>

<svelte:head>
  <title>{canvasTitle || `${canvasName} - Rill`}</title>
</svelte:head>

<slot {ready} {reconcileErrorMessage} {isLoading} {isReconciling} />
