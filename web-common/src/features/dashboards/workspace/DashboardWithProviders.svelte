<script lang="ts">
  import DashboardURLStateSyncWrapper from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSyncWrapper.svelte";
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
      <DashboardURLStateSyncWrapper>
        <DashboardThemeProvider>
          <Dashboard {metricsViewName} {exploreName} />
        </DashboardThemeProvider>
      </DashboardURLStateSyncWrapper>
    </StateManagersProvider>
  {/key}
{:else}
  <Spinner size="48px" />
{/if}
