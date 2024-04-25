<script lang="ts">
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  const queryClient = useQueryClient();

  export let data;

  resetSelectedMockUserAfterNavigate(queryClient);

  $: metricsViewName = data.metricsView.meta?.name?.name as string;
</script>

<svelte:head>
  <title>Rill Developer | {metricsViewName}</title>
</svelte:head>

{#key metricsViewName}
  <StateManagersProvider {metricsViewName}>
    <DashboardStateProvider metricViewName={metricsViewName}>
      <DashboardURLStateProvider metricViewName={metricsViewName}>
        <DashboardThemeProvider>
          <Dashboard metricViewName={metricsViewName} />
        </DashboardThemeProvider>
      </DashboardURLStateProvider>
    </DashboardStateProvider>
  </StateManagersProvider>
{/key}
