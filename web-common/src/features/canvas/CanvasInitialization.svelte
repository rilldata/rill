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
    createQueryServiceResolveCanvas,
    createRuntimeServiceListResources,
    type V1MetricsView,
    type V1ResolveCanvasResponse,
  } from "@rilldata/web-common/runtime-client";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import { findRootCause } from "../entity-management/error-utils";

  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;

  export let canvasName: string;
  export let instanceId: string;
  export let showBanner = false;
  export let projectId: string | undefined = undefined;

  let resolvedStore: CanvasStore | undefined = undefined;

  $: ({ url } = $page);

  $: existingStore = getCanvasStoreUnguarded(canvasName, instanceId);

  $: resourceQuery = useResource(
    instanceId,
    canvasName,
    ResourceKind.Canvas,
    {},
  );

  $: fetchedCanvasQuery = !existingStore
    ? createQueryServiceResolveCanvas(
        instanceId,
        canvasName,
        {},
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

  $: resource = resourceQuery ? $resourceQuery?.data : undefined;

  $: errorMessage = !validSpec
    ? reconcileError || resource?.meta?.reconcileError
    : undefined;

  // Fetch all resources only when there's an error, to find the root cause
  $: allResourcesQuery = errorMessage
    ? createRuntimeServiceListResources(instanceId)
    : undefined;

  $: rootCause =
    errorMessage && resource && $allResourcesQuery?.data
      ? findRootCause(resource, $allResourcesQuery.data.resources ?? [])
      : undefined;

  // Replace the error message body with the root cause if found
  $: resolvedErrorMessage = rootCause?.meta?.reconcileError
    ? `${rootCause.meta.name?.name}: ${rootCause.meta.reconcileError}`
    : errorMessage;

  $: resolvedStore = getResolvedStore(
    fetchedCanvas,
    isReconciling,
    existingStore,
    instanceId,
  );

  $: ready = !!resolvedStore;

  $: if (resolvedStore) {
    resolvedStore.canvasEntity
      .onUrlChange({ url, projectId })
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
    if (fetchedCanvas && !isReconciling) {
      const metricsViews: Record<string, V1MetricsView | undefined> = {};
      const refMetricsViews = fetchedCanvas?.referencedMetricsViews;
      if (refMetricsViews) {
        Object.keys(refMetricsViews).forEach((key) => {
          metricsViews[key] = refMetricsViews?.[key]?.metricsView;
        });
      }

      const validSpec = fetchedCanvas?.canvas?.canvas?.state?.validSpec;

      if (validSpec) {
        const processed = {
          canvas: fetchedCanvas?.canvas?.canvas?.state?.validSpec,
          components: fetchedCanvas?.resolvedComponents,
          metricsViews,
          filePath: fetchedCanvas?.canvas?.meta?.filePaths?.[0],
        };

        return setCanvasStore(canvasName, instanceId, processed);
      }
    }

    return existingStore;
  }
</script>

<svelte:head>
  <title>{canvasTitle || `${canvasName} - Rill`}</title>
</svelte:head>

<slot {ready} errorMessage={resolvedErrorMessage} {isLoading} {isReconciling} />
