<script lang="ts">
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import PivotOnlyDashboard from "@rilldata/web-common/features/dashboards/workspace/PivotOnlyDashboard.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { exploreName } = data;

  $: ({ instanceId } = $runtime);

  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewName = exploreSpec.metricsView ?? "";
</script>

{#if exploreName && metricsViewName}
  {#key exploreName}
    <StateManagersProvider {metricsViewName} {exploreName}>
      <DashboardStateManager
        {exploreName}
        disableSessionStorage
        disableMostRecentDashboardState
      >
        <DashboardThemeProvider>
          <PivotOnlyDashboard {metricsViewName} {exploreName} />
        </DashboardThemeProvider>
      </DashboardStateManager>
    </StateManagersProvider>
  {/key}
{/if}
