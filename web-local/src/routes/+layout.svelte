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
  import {
    initPosthog,
    posthogIdentify,
  } from "@rilldata/web-common/lib/analytics/posthog";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    errorEventHandler,
    initMetrics,
  } from "@rilldata/web-common/metrics/initMetrics";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import { AppMode, previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";
  import { LOCAL_HOST, LOCAL_INSTANCE_ID } from "../lib/runtime-client";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import type { Query } from "@tanstack/query-core";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { LayoutData } from "./$types";
  import PreviewModeNav from "../features/preview/PreviewModeNav.svelte";
  import {
    isPreviewRoute,
    isDeveloperRoute,
    showPreviewNav,
  } from "./route-constants";
  import "@rilldata/web-common/app.css";

  export let data: LayoutData;

  const { deploy } = featureFlags;

  queryClient.getQueryCache().config.onError = (error: unknown, query: Query) =>
    errorEventHandler?.requestErrorEventHandler(error, query);
  initPylonWidget();

  // Preview mode store sync:
  // 1. Backend lock: if --preview flag is set, always true
  // 2. URL-derived: preview routes (/dashboards, /ai, /status) → true,
  //    developer routes (/, /files) → false
  // 3. Preserved: shared routes (/explore, /canvas, /deploy) keep previous value
  $: {
    if (data.previewMode) {
      previewModeStore.set(true);
    } else if (isPreviewRoute($page.url.pathname)) {
      previewModeStore.set(true);
    } else if (isDeveloperRoute($page.url.pathname)) {
      previewModeStore.set(false);
    }
  }

  let removeJavascriptListeners: () => void;
  onMount(async () => {
    const config = data.metadata;

    const shouldSendAnalytics =
      config.analyticsEnabled && !import.meta.env.VITE_PLAYWRIGHT_TEST && !dev;

    if (shouldSendAnalytics) {
      await initMetrics(config, host); // Proxies events through the Rill "intake" service
      initPosthog(config.version);
      posthogIdentify(config.userId, {
        installId: config.installId,
      });

      removeJavascriptListeners =
        errorEventHandler.addJavascriptErrorListeners();
    }

    featureFlags.set(false, "adminServer");
    featureFlags.set(config.readonly || data.previewMode, "readOnly");
  });

  /**
   * Async mount doesnt support an unsubscribe method.
   * So we need this to make sure javascript listeners for error handler is removed.
   */
  onMount(() => {
    return () => removeJavascriptListeners?.();
  });

  const host = LOCAL_HOST;
  const instanceId = LOCAL_INSTANCE_ID;

  $: ({ route } = $page);
  $: onDeployPage = isDeployPage($page);
  $: isPreviewMode = $previewModeStore;

  // Preview mode from store OR (viz) route group
  $: mode =
    isPreviewMode || route.id?.includes("(viz)")
      ? AppMode.Preview
      : AppMode.Developer;

  $: shouldShowPreviewNav =
    isPreviewMode && showPreviewNav($page.url.pathname) && !onDeployPage;

  $: onWelcomePage = route.id?.startsWith("/(misc)/welcome");
</script>

<Tooltip.Provider>
  <QueryClientProvider client={queryClient}>
    <RuntimeProvider {host} {instanceId}>
      <FileAndResourceWatcher {host} {instanceId}>
        <div
          class="body h-screen w-screen overflow-hidden absolute flex flex-col"
        >
          {#if data.initialized && !onWelcomePage}
            <BannerCenter />
            <RepresentingUserBanner />
            <ApplicationHeader {mode} />
            {#if shouldShowPreviewNav}
              <PreviewModeNav />
            {/if}
            {#if $deploy}
              <RemoteProjectManager />
            {/if}
          {/if}

          <slot />
        </div>
      </FileAndResourceWatcher>
    </RuntimeProvider>
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
</Tooltip.Provider>

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
