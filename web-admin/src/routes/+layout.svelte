<script lang="ts">
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    QueryCache,
    QueryClient,
    QueryClientProvider,
  } from "@tanstack/svelte-query";
  import { globalErrorCallback } from "../components/errors/error-utils";
  import ErrorBoundary from "../components/errors/ErrorBoundary.svelte";
  import TopNavigationBar from "../components/navigation/TopNavigationBar.svelte";

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
        placeholderData: {}, // there's an issue somewhere in the Leaderboard components that depends on this setting
      },
    },
  });

  featureFlags.set({
    // Set read-only mode so that the user can't edit the dashboard
    readOnly: true,
  });
</script>

<svelte:head>
  <meta name="description" content="Rill Cloud" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <div class="flex flex-col h-screen">
      <main class="flex-grow flex flex-col">
        <TopNavigationBar />
        <div class="flex-grow overflow-auto">
          <ErrorBoundary>
            <slot />
          </ErrorBoundary>
        </div>
      </main>
    </div>
  </QueryClientProvider>
</RillTheme>
