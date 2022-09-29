<script lang="ts">
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";
  import "../app.css";

  const queryClient = new QueryClient();
</script>

<svelte:head>
  <meta name="description" content="Rill Cloud" />
</svelte:head>

<main>
  <QueryClientProvider client={queryClient}>
    <slot />
  </QueryClientProvider>
</main>

<footer>
  <p>Rill Data</p>
</footer>

<style>
  main {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 1rem;
    width: 100%;
    max-width: 1024px;
    margin: 0 auto;
    box-sizing: border-box;
  }

  footer {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 40px;
  }

  @media (min-width: 480px) {
    footer {
      padding: 40px 0;
    }
  }
</style>
