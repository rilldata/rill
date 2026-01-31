<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";
  import { OLAP_ENGINES } from "@rilldata/web-common/features/sources/modal/constants";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Database } from "lucide-svelte";

  export let resource: V1Resource | undefined;
  export let hasUnsavedChanges = false;

  $: connectorName = resource?.meta?.name?.name;
  $: driverName = resource?.connector?.spec?.driver;
  $: hasReconcileError = !!resource?.meta?.reconcileError;
  $: isOlapConnector = driverName ? OLAP_ENGINES.includes(driverName) : false;

  function openDataExplorer() {
    if (!driverName || !connectorName) return;
    // Create a V1ConnectorDriver-compatible object for openExplorerForConnector
    const connectorDriver = {
      name: driverName,
      displayName: connectorName,
      implementsOlap: true,
    };
    addSourceModal.openExplorerForConnector(connectorDriver, connectorName);
  }
</script>

{#if isOlapConnector}
  <Tooltip distance={8}>
    <Button
      type="primary"
      onClick={openDataExplorer}
      disabled={hasUnsavedChanges || hasReconcileError || !driverName}
      label="Open Data Explorer"
    >
      <Database size="14px" />
      Open Data Explorer
    </Button>
    <TooltipContent slot="tooltip-content">
      {#if hasUnsavedChanges}
        Save your changes first
      {:else if hasReconcileError}
        Fix connector errors first
      {:else}
        Browse tables and create models from this connector
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
