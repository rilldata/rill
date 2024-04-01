<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  // import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  // import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useQueryClient } from "@tanstack/svelte-query";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";

  const queryClient = useQueryClient();

  // const { readOnly } = featureFlags;

  export let data;

  $: metricViewName = $page.params.name;

  resetSelectedMockUserAfterNavigate(queryClient);
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<!-- {#if ($resourceStatusStore.status === ResourceStatus.Idle) || showErrorPage} -->
{#key metricViewName}
  <StateManagersProvider metricsViewName={metricViewName}>
    <DashboardStateProvider {metricViewName}>
      <!-- <DashboardURLStateProvider {metricViewName}> -->
      <DashboardThemeProvider>
        <Dashboard {metricViewName} dimensions={data.dimensions} />
      </DashboardThemeProvider>
      <!-- </DashboardURLStateProvider> -->
    </DashboardStateProvider>
  </StateManagersProvider>
{/key}
<!-- {:else if $resourceStatusStore.status === ResourceStatus.Busy}
  <div class="grid h-screen w-full place-content-center">
    <ReconcilingSpinner />
  </div>
{/if} -->
