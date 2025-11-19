<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";
  import { writable } from "svelte/store";

  const PollIntervalWhenDashboardFirstReconciling = 1000;
  const PollIntervalWhenDashboardErrored = 5000;

  export let data;

  $: ({
    canvasName,
    project: { name: project },
    organization,
  } = data);

  $: ({ instanceId } = $runtime);

  const canvasNameStore = writable(canvasName);
  $: canvasNameStore.set(canvasName);
  const orgAndProjectNameStore = writable({ organization: "", project: "" });
  $: orgAndProjectNameStore.set({
    organization: organization,
    project: project,
  });

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas, {
    refetchInterval: (query) => {
      const resource = query?.state?.data?.resource;
      if (!resource) return false;
      if (isCanvasReconcilingForFirstTime(resource))
        return PollIntervalWhenDashboardFirstReconciling;
      if (isCanvasErrored(resource)) return PollIntervalWhenDashboardErrored;
      return false;
    },
  });

  $: canvasQueryResponse = $canvasQuery;

  $: ({ isError, error, data: canvasResource } = canvasQueryResponse);

  $: canvasTitle = canvasResource?.canvas?.state?.validSpec?.displayName;
  $: hasBanner = !!canvasResource?.canvas?.state?.validSpec?.banner;

  $: isCanvasNotFound =
    !canvasResource && isError && error?.response?.status === 404;

  // If no canvas dashboard is found, show a 404 page
  $: if (isCanvasNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Canvas not found",
      body: `The canvas dashboard you requested could not be found. Please check that you provided the name of a working canvas dashboard.`,
    });
  }

  // Display a dashboard banner
  $: if (hasBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: canvasResource?.canvas?.state?.validSpec?.banner,
        iconType: "alert",
      },
    });
  }

  onNavigate(({ from, to }) => {
    const changedDashboard =
      !from || !to || from.params.dashboard !== to.params.dashboard;
    // Clear out any dashboard banners
    if (hasBanner && changedDashboard) {
      eventBus.emit("remove-banner", DashboardBannerID);
    }
  });

  function isCanvasReconcilingForFirstTime(canvasResource: V1Resource) {
    if (!canvasResource) return undefined;
    const isCanvasReconcilingForFirstTime =
      !canvasResource.canvas?.state?.validSpec &&
      !canvasResource?.meta?.reconcileError;
    return isCanvasReconcilingForFirstTime;
  }

  function isCanvasErrored(canvasResource: V1Resource) {
    if (!canvasResource) return undefined;
    // We only consider a dashboard errored (from the end-user perspective) when BOTH a reconcile error exists AND a validSpec does not exist.
    // If there's any validSpec (which can persist from a previous, non-current spec), then we serve that version of the dashboard to the user,
    // so the user does not see an error state.
    const isCanvasErrored =
      !canvasResource.canvas?.state?.validSpec &&
      !!canvasResource?.meta?.reconcileError;
    return isCanvasErrored;
  }
</script>

<svelte:head>
  <title>{canvasTitle || `${canvasName} - Rill`}</title>
</svelte:head>

{#if isCanvasReconcilingForFirstTime(canvasResource)}
  <DashboardBuilding />
{:else}
  <CanvasDashboardEmbed resource={canvasResource} {canvasName} />
{/if}
