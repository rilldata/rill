<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ exploreName } = data);

  const client = useRuntimeClient();

  $: explore = createRuntimeServiceGetExplore(
    client,
    { name: exploreName },
    {
      query: { enabled: !!exploreName },
    },
  );

  $: metricsViewName = $explore.data?.metricsView?.meta?.name?.name;
  $: measures =
    $explore.data?.explore?.explore?.state?.validSpec?.measures ?? [];
</script>

<svelte:head>
  <title>Rill | {exploreName}</title>
</svelte:head>

{#if measures.length === 0}
  <ErrorPage
    statusCode={undefined}
    header="Error fetching dashboard"
    body="No measures available"
  />
{:else if metricsViewName}
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
