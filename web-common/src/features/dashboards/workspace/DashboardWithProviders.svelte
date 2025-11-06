<script lang="ts">
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import Spinner from "../../entity-management/Spinner.svelte";
  import StateManagersProvider from "../state-managers/StateManagersProvider.svelte";
  import Dashboard from "./Dashboard.svelte";

  export let metricsViewName: string;
  export let exploreName: string;
</script>

{#if metricsViewName}
  {#key exploreName + metricsViewName}
    <StateManagersProvider {metricsViewName} {exploreName} visualEditing>
      <DashboardStateManager {exploreName}>
        <Dashboard {metricsViewName} {exploreName} />
      </DashboardStateManager>
    </StateManagersProvider>
  {/key}
{:else}
  <Spinner size="48px" />
{/if}
