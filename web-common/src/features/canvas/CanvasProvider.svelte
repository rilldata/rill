<script lang="ts">
  import { handleCanvasStoreInitialization } from "./state-managers/state-managers";
  import { page } from "$app/stores";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import { onNavigate } from "$app/navigation";
  import { writable } from "svelte/store";
  import DelayedSpinner from "../entity-management/DelayedSpinner.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";

  export let canvasName: string;
  export let instanceId: string;
  export let ready = false;
  export let showBanner = false;

  $: ({ url } = $page);

  $: ({ canvasStoreStore, reconcilingStore, errorMessageStore } =
    handleCanvasStoreInitialization(canvasName, instanceId));

  $: errorMessage = $errorMessageStore;

  $: resolvedStore = $canvasStoreStore;

  $: ready = !!resolvedStore;

  $: if (resolvedStore) {
    resolvedStore.canvasEntity
      .onUrlChange({ url, loadFunction: false })
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
</script>

<svelte:head>
  <title>{canvasTitle || `${canvasName} - Rill`}</title>
</svelte:head>

<div class="size-full justify-center items-center flex flex-col">
  {#if resolvedStore}
    <slot />
  {:else if $reconcilingStore}
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
