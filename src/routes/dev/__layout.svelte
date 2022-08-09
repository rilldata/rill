<script lang="ts">
  import { QueryClient, QueryClientProvider } from "@sveltestack/svelte-query";

  const queryClient = new QueryClient();
</script>

<QueryClientProvider client={queryClient}>
  <div class="bg-white p-8" style:min-width="100vw" style:min-height="100vh">
    <slot />
  </div>
</QueryClientProvider>
