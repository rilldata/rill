<script lang="ts">
  import Spinner from "../../entity-management/Spinner.svelte";
  import DashboardThemeProvider from "../DashboardThemeProvider.svelte";
  import DashboardUrlStateProvider from "../proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "../state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "../stores/DashboardStateProvider.svelte";
  import Dashboard from "./Dashboard.svelte";

  export let metricsViewName: string;
  export let exploreName: string;
</script>

{#if metricsViewName}
  {#key exploreName}
    <StateManagersProvider {metricsViewName} {exploreName} visualEditing>
      <DashboardStateProvider {exploreName}>
        <DashboardUrlStateProvider {metricsViewName}>
          <DashboardThemeProvider>
            <Dashboard {metricsViewName} {exploreName} />
          </DashboardThemeProvider>
        </DashboardUrlStateProvider>
      </DashboardStateProvider>
    </StateManagersProvider>
  {/key}
{:else}
  <Spinner size="48px" />
{/if}
