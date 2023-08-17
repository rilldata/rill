<script lang="ts">
  import { page } from "$app/stores";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import TimeDimensionDetailsTableRT from "@rilldata/web-common/features/dashboards/time-dimension-details-rt/TimeDimensionDetailsTableRT.svelte";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { createQueryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";

  $: metricViewName = $page.params.name;
  const queryClient = createQueryClient();
</script>

<QueryClientProvider client={queryClient}>
  <StateManagersProvider metricsViewName={metricViewName}>
    <div class="p-8">
      <!-- <TimeDimensionDetailsTable /> -->
      <TimeDimensionDetailsTableRT />
    </div>
  </StateManagersProvider>
</QueryClientProvider>
