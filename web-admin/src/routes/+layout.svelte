<script lang="ts">
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";
  import TopNavigationBar from "../components/navigation/TopNavigationBar.svelte";

  const queryClient = new QueryClient();
</script>

<svelte:head>
  <meta name="description" content="Rill Cloud" />
</svelte:head>

<RillTheme>
  <QueryClientProvider client={queryClient}>
    <div class="flex flex-col h-screen">
      <main class="flex-grow flex flex-col h-full">
        <TopNavigationBar />
        <div class="flex-grow">
          <slot />
        </div>
      </main>
      <footer class="text-center">
        <p>Rill Data</p>
      </footer>
    </div>
  </QueryClientProvider>
</RillTheme>
