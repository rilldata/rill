<script lang="ts">
  import { page } from "$app/stores";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { createQueryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import ExamplePivot from "@rilldata/web-common/features/dashboards/pivot/ExamplePivot.svelte";

  $: metricViewName = $page.params.name;
  const queryClient = createQueryClient();
</script>

<QueryClientProvider client={queryClient}>
  <StateManagersProvider metricsViewName={metricViewName}>
    <div class="p-8">
      <ExamplePivot />
    </div>
  </StateManagersProvider>
</QueryClientProvider>
