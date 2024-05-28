<script lang="ts">
  import { page } from "$app/stores";
  import { initCloudMetrics } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import ErrorBoundary from "../features/errors/ErrorBoundary.svelte";
  import { createGlobalErrorCallback } from "../features/errors/error-utils";
  import { initPylonWidget } from "../features/help/initPylonWidget";
  import TopNavigationBar from "../features/navigation/TopNavigationBar.svelte";
  import { clearViewedAsUserAfterNavigate } from "../features/view-as-user/clearViewedAsUser";

  // Motivation:
  // - https://tkdodo.eu/blog/breaking-react-querys-api-on-purpose#a-bad-api
  // - https://tkdodo.eu/blog/react-query-error-handling#the-global-callbacks
  queryClient.getQueryCache().config.onError =
    createGlobalErrorCallback(queryClient);

  // The admin server enables some dashboard features like scheduled reports and alerts
  // Set read-only mode so that the user can't edit the dashboard
  featureFlags.set(true, "adminServer", "readOnly");

  clearViewedAsUserAfterNavigate(queryClient);

  initCloudMetrics();
  initPylonWidget();

  onMount(() => errorEventHandler?.addJavascriptErrorListeners());

  $: isEmbed = $page.url.pathname === "/-/embed";
</script>

<svelte:head>
  <meta content="Rill Cloud" name="description" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <main class="flex flex-col min-h-screen h-screen">
      {#if !isEmbed}
        <TopNavigationBar />
      {/if}
      <ErrorBoundary>
        <slot />
      </ErrorBoundary>
    </main>
  </QueryClientProvider>

  <NotificationCenter />
</RillTheme>
