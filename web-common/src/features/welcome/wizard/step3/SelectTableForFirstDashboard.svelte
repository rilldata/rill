<script lang="ts">
  import { goto } from "$app/navigation";
  import { connectorExplorerStore } from "@rilldata/web-common/features/connectors/connector-explorer-store";
  import { Button } from "../../../../components/button";
  import { BehaviourEventMedium } from "../../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../../metrics/service/MetricsTypes";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import ConnectorExplorer from "../../../connectors/ConnectorExplorer.svelte";
  import { useCreateMetricsViewFromTableUIAction } from "../../../metrics-views/ai-generation/generateMetricsView";

  const store = connectorExplorerStore.duplicateStore();

  const explorerState = store.store;
  $: selectedItem = $explorerState.selectedItem;

  async function createDashboard() {
    await createMetricsViewFromTable();
    clearWelcomeStateFromLocalStorage();
  }

  async function skip() {
    await goto("/");
    clearWelcomeStateFromLocalStorage();
  }

  let createMetricsViewFromTable: () => Promise<void>;
  $: {
    if (selectedItem && selectedItem?.table) {
      createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
        $runtime.instanceId,
        selectedItem.connector,
        selectedItem?.database ?? "",
        selectedItem?.databaseSchema ?? "",
        selectedItem.table,
        true,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace,
      );
    }
  }

  function clearWelcomeStateFromLocalStorage() {
    localStorage.removeItem("welcomeStep");
    localStorage.removeItem("welcomeManagementType");
    localStorage.removeItem("welcomeOlapDriver");
  }
</script>

<div class="w-[544px] p-6 flex flex-col gap-y-4">
  <h2 class="text-lead">Pick a table to power your first dashboard</h2>
  <div class="max-h-[300px] overflow-y-auto">
    <ConnectorExplorer {store} />
  </div>
  {#if selectedItem?.table}
    <Button type="primary" large on:click={createDashboard}>
      Create dashboard
    </Button>
  {:else}
    <Button type="secondary" large on:click={skip}>Skip</Button>
  {/if}
</div>
