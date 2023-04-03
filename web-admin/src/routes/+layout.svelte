<script lang="ts">
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { featureFlags } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";
  import TopNavigationBar from "../components/navigation/TopNavigationBar.svelte";

  const queryClient = new QueryClient({
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
    <div class="flex flex-col min-h-screen">
      <main class="flex-grow flex flex-col">
        <TopNavigationBar />
        <div class="flex-grow overflow-auto">
          <slot />
        </div>
      </main>
      <footer class="text-center">
        <p>Rill Data</p>
      </footer>
    </div>
  </QueryClientProvider>
</RillTheme>
