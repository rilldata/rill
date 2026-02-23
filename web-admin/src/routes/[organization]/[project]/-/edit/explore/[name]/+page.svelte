<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ metricsView, explore, exploreName } = data);

  $: metricsViewName = metricsView?.meta?.name?.name as string;

  $: ({ instanceId } = $runtime);

  $: exploreQuery = useExploreValidSpec(instanceId, exploreName);
  $: measures = $exploreQuery.data?.explore?.measures ?? [];
</script>

<svelte:head>
  <title>Rill | {exploreName}</title>
</svelte:head>

{#if measures.length === 0}
  <ErrorPage
    statusCode={$exploreQuery.error?.response?.status}
    header="Error fetching dashboard"
    body="No measures available"
  />
{:else}
  <div class="h-full overflow-hidden">
    {#key exploreName}
      <StateManagersProvider {metricsViewName} {exploreName}>
        <DashboardStateManager {exploreName}>
          <Dashboard {metricsViewName} {exploreName} />
        </DashboardStateManager>
      </StateManagersProvider>
    {/key}
  </div>
{/if}
