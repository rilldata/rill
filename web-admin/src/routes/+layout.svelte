<script lang="ts">
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";

  const queryClient = new QueryClient();
</script>

<svelte:head>
  <meta name="description" content="Rill Cloud" />
</svelte:head>

<main>
  <RillTheme>
    <QueryClientProvider client={queryClient}>
      <slot />
    </QueryClientProvider>
  </RillTheme>
</main>

<footer>
  <p>Rill Data</p>
</footer>
