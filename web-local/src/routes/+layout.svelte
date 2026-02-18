<script lang="ts">
  import { dev } from "$app/environment";
  import { page } from "$app/stores";
  import BannerCenter from "@rilldata/web-common/components/banner/BannerCenter.svelte";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import RepresentingUserBanner from "@rilldata/web-common/features/authentication/RepresentingUserBanner.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { initPylonWidget } from "@rilldata/web-common/features/help/initPylonWidget";
  import RemoteProjectManager from "@rilldata/web-common/features/project/RemoteProjectManager.svelte";
  import ApplicationHeader from "@rilldata/web-common/layout/ApplicationHeader.svelte";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import {
    initPosthog,
    posthogIdentify,
  } from "@rilldata/web-common/lib/analytics/posthog";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    errorEventHandler,
    initMetrics,
  } from "@rilldata/web-common/metrics/initMetrics";
  import { localServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { AxiosError } from "axios";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import type { LayoutData } from "./$types";
  import PreviewModeNav from "./PreviewModeNav.svelte";
  import {
    isPreviewRoute,
    isDeveloperRoute,
    showPreviewNav,
  } from "./route-constants";
  import "@rilldata/web-common/app.css";

  export let data: LayoutData;

  const { deploy } = featureFlags;

  queryClient.getQueryCache().config.onError = (
    error: AxiosError,
    query: Query,
  ) => errorEventHandler?.requestErrorEventHandler(error, query);
  initPylonWidget();

  let removeJavascriptListeners: () => void;

  // Sync preview mode:
  // - If --preview or --previewer flag is set, always lock to preview mode
  // - Otherwise, infer from the current URL so refresh on preview pages stays in preview mode
  //   and shared routes (/explore, /canvas) preserve the current mode
  $: {
    const serverLocked =
      (data.previewMode ?? false) || (data.previewerMode ?? false);
    if (serverLocked) {
      previewModeStore.set(true);
    } else if (isPreviewRoute($page.url.pathname)) {
      previewModeStore.set(true);
    } else if (isDeveloperRoute($page.url.pathname)) {
      previewModeStore.set(false);
    }
    // For shared routes (/explore, /canvas, /deploy), keep current store value
  }

  onMount(async () => {
    // If in preview mode and on root, redirect to /home
    if ($previewModeStore && window.location.pathname === "/") {
      goto("/home");
    }

    const config = await localServiceGetMetadata();

    const shouldSendAnalytics =
      config.analyticsEnabled && !import.meta.env.VITE_PLAYWRIGHT_TEST && !dev;

    if (shouldSendAnalytics) {
      await initMetrics(config); // Proxies events through the Rill "intake" service
      initPosthog(config.version);
      posthogIdentify(config.userId, {
        installId: config.installId,
      });

      removeJavascriptListeners =
        errorEventHandler.addJavascriptErrorListeners();
    }

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly, "readOnly");
  });

  /**
   * Async mount doesnt support an unsubscribe method.
   * So we need this to make sure javascript listeners for error handler is removed.
   */
  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  $: ({ host, instanceId } = $runtime);

  $: onDeployPage = isDeployPage($page);

  $: isPreviewMode = $previewModeStore;

  $: shouldShowPreviewNav =
    isPreviewMode && showPreviewNav($page.url.pathname) && !onDeployPage;
</script>

<QueryClientProvider client={queryClient}>
  <FileAndResourceWatcher {host} {instanceId}>
    <div class="body h-screen w-screen overflow-hidden absolute flex flex-col">
      {#if data.initialized}
        <BannerCenter />
        <RepresentingUserBanner />
        <ApplicationHeader
          logoHref={isPreviewMode ? "/home" : "/"}
          breadcrumbResourceHref={isPreviewMode
            ? (name, kind) => `/${kind}/${name}`
            : undefined}
          noBorder={isPreviewMode}
          previewerMode={data.previewerMode ?? false}
        />
        {#if shouldShowPreviewNav}
          <PreviewModeNav />
        {/if}
        {#if $deploy}
          <RemoteProjectManager />
        {/if}
      {/if}

      <div class="flex-1 overflow-hidden" class:bg-white={onDeployPage}>
        <slot />
      </div>
    </div>
  </FileAndResourceWatcher>
</QueryClientProvider>

{#if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
