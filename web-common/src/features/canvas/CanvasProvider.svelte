<script lang="ts">
  import {
    getCanvasStoreUnguarded,
    setCanvasStore,
    type CanvasStore,
  } from "./state-managers/state-managers";
  import { page } from "$app/stores";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { onNavigate } from "$app/navigation";
  import { writable } from "svelte/store";
  import DelayedSpinner from "../entity-management/DelayedSpinner.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";

  import {
    createQueryServiceResolveCanvas,
    type V1MetricsView,
    type V1ResolveCanvasResponse,
  } from "@rilldata/web-common/runtime-client";

  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;

  export let canvasName: string;
  export let instanceId: string;
  export let ready = false;
  export let showBanner = false;
  export let projectId: string | undefined = undefined;

  let resolvedStore: CanvasStore | undefined = undefined;

  $: ({ url } = $page);

  $: existingStore = getCanvasStoreUnguarded(canvasName, instanceId);

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

  $: fetchedCanvas = fetchedCanvasQuery ? $fetchedCanvasQuery?.data : undefined;

  $: validSpec = fetchedCanvas?.canvas?.canvas?.state?.validSpec;
  $: reconcileError = fetchedCanvas?.canvas?.meta?.reconcileError;

  $: isReconciling = !existingStore && !validSpec && !reconcileError;

  $: errorMessage = !validSpec && reconcileError;

  $: if (fetchedCanvas && !isReconciling) {
    const metricsViews: Record<string, V1MetricsView | undefined> = {};
    const refMetricsViews = fetchedCanvas?.referencedMetricsViews;
    if (refMetricsViews) {
      Object.keys(refMetricsViews).forEach((key) => {
        metricsViews[key] = refMetricsViews?.[key]?.metricsView;
      });
    }

    const processed = {
      canvas: fetchedCanvas?.canvas?.canvas?.state?.validSpec,
      components: fetchedCanvas?.resolvedComponents,
      metricsViews,
      filePath: fetchedCanvas?.canvas?.meta?.filePaths?.[0],
    };

    resolvedStore = setCanvasStore(canvasName, instanceId, processed);
  } else if (existingStore) {
    resolvedStore = existingStore;
  }

  $: ready = !!resolvedStore;

  $: if (resolvedStore) {
    resolvedStore.canvasEntity
      .onUrlChange({ url, projectId })
      .catch(console.error);
  }

  $: title = resolvedStore?.canvasEntity.titleStore || writable("");
  $: canvasTitle = $title;

  $: bannerStore = resolvedStore?.canvasEntity._banner || writable("");
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
</script>

<svelte:head>
  <title>{canvasTitle || `${canvasName} - Rill`}</title>
</svelte:head>

<div class="size-full justify-center items-center flex flex-col">
  {#if resolvedStore}
    <slot />
  {:else if isReconciling}
    <DashboardBuilding />
  {:else if errorMessage}
    <ErrorPage
      statusCode={404}
      header="Canvas not found"
      body={errorMessage || "An unknown error occurred."}
    />
  {:else}
    <header
      role="presentation"
      class="bg-background border-b py-4 px-2 w-full h-[100px] select-none z-50 flex items-center justify-center"
    ></header>
    <div class="size-full flex justify-center items-center">
      <DelayedSpinner isLoading={true} size="48px" />
    </div>
  {/if}
</div>
