<script lang="ts">
  import DashboardStateLoader from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";
  import Spinner from "../../entity-management/Spinner.svelte";
  import DashboardThemeProvider from "../DashboardThemeProvider.svelte";
  import StateManagersProvider from "../state-managers/StateManagersProvider.svelte";
  import Dashboard from "./Dashboard.svelte";

  export let metricsViewName: string;
  export let exploreName: string;
</script>

{#if metricsViewName}
  {#key exploreName + metricsViewName}
    <StateManagersProvider {metricsViewName} {exploreName} visualEditing>
      <DashboardStateLoader {exploreName}>
        <DashboardThemeProvider>
          <Dashboard {metricsViewName} {exploreName} />
        </DashboardThemeProvider>
      </DashboardStateLoader>
    </StateManagersProvider>
  {/key}
{:else}
  <Spinner size="48px" />
{/if}
