<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { isDashboardPage } from "@rilldata/web-common/features/navigation/nav-utils";
  import { initCloudMetrics } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import {
    featureFlags,
    retainFeaturesFlags,
  } from "@rilldata/web-common/features/feature-flags";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { errorEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import ErrorBoundary from "../features/errors/ErrorBoundary.svelte";
  import { createGlobalErrorCallback } from "../features/errors/error-utils";
  import TopNavigationBar from "../features/navigation/TopNavigationBar.svelte";
  import { clearViewedAsUserAfterNavigate } from "../features/view-as-user/clearViewedAsUser";

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
      },
    },
  });
  // Motivation:
  // - https://tkdodo.eu/blog/breaking-react-querys-api-on-purpose#a-bad-api
  // - https://tkdodo.eu/blog/react-query-error-handling#the-global-callbacks
  queryClient.getQueryCache().config.onError =
    createGlobalErrorCallback(queryClient);

  // The admin server enables some dashboard features like scheduled reports and alerts
  // Set read-only mode so that the user can't edit the dashboard
  featureFlags.set(true, "adminServer", "readOnly");

  // Temporary flag to show/hide the WIP alerts feature
  if (localStorage.getItem("alerts") === "true") {
    featureFlags.set(true, "alerts");
  }

  beforeNavigate(retainFeaturesFlags);
  clearViewedAsUserAfterNavigate(queryClient);
  initCloudMetrics();

  onMount(() => errorEvent?.addJavascriptErrorListeners());

  $: isEmbed = $page.url.pathname === "/-/embed";

  // The Dashboard component assumes a page height of `h-screen`. This is somehow motivated by
  // making the line charts and leaderboards scroll independently.
  // However, `h-screen` screws up overflow/scroll on all other pages, so we only apply it to the dashboard.
  // (This all feels hacky and should not be considered optimal.)
  $: onDashboardPage = isDashboardPage($page);
</script>

<svelte:head>
  <meta content="Rill Cloud" name="description" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <main class="flex flex-col min-h-screen {onDashboardPage && 'h-screen'}">
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
