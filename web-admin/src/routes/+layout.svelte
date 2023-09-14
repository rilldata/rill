<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import {
    featureFlags,
    retainFeaturesFlags,
  } from "@rilldata/web-common/features/feature-flags";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import {
    QueryCache,
    QueryClient,
    QueryClientProvider,
  } from "@tanstack/svelte-query";
  import { globalErrorCallback } from "../components/errors/error-utils";
  import ErrorBoundary from "../components/errors/ErrorBoundary.svelte";
  import TopNavigationBar from "../components/navigation/TopNavigationBar.svelte";
  import { clearViewedAsUserAfterNavigate } from "../features/view-as-user/clearViewedAsUser";

  const queryClient = new QueryClient({
    queryCache: new QueryCache({
      // Motivation:
      // - https://tkdodo.eu/blog/breaking-react-querys-api-on-purpose#a-bad-api
      // - https://tkdodo.eu/blog/react-query-error-handling#the-global-callbacks
      onError: globalErrorCallback,
    }),
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
      },
    },
  });

  featureFlags.set({
    // Set read-only mode so that the user can't edit the dashboard
    readOnly: true,
  });

  beforeNavigate(retainFeaturesFlags);
  clearViewedAsUserAfterNavigate(queryClient);
</script>

<svelte:head>
  <meta content="Rill Cloud" name="description" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <main class="flex flex-col h-screen">
      <TopNavigationBar />
      <div class="flex-grow overflow-hidden">
        <ErrorBoundary>
          <slot />
        </ErrorBoundary>
      </div>
    </main>
  </QueryClientProvider>

  <NotificationCenter />
</RillTheme>
