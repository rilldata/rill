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
    <main>
      <TopNavigationBar />
      <slot />
    </main>

    <footer>
      <p>Rill Data</p>
    </footer>
  </QueryClientProvider>
</RillTheme>
