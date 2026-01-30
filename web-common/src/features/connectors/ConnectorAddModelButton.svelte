<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";
  import { OLAP_ENGINES } from "@rilldata/web-common/features/sources/modal/constants";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Plus } from "lucide-svelte";

  export let resource: V1Resource | undefined;
  export let hasUnsavedChanges = false;

  $: connectorName = resource?.meta?.name?.name;
  $: driverName = resource?.connector?.spec?.driver;
  $: hasReconcileError = !!resource?.meta?.reconcileError;
  $: isOlapConnector = driverName ? OLAP_ENGINES.includes(driverName) : false;

  function openAddModel() {
    if (!driverName || !connectorName) return;
    addSourceModal.open(driverName, connectorName);
  }
</script>

{#if !isOlapConnector}
  <Tooltip distance={8}>
    <Button
      type="primary"
      onClick={openAddModel}
      disabled={hasUnsavedChanges || hasReconcileError || !driverName}
      label="Import data"
    >
      <Plus size="14px" />
      Import data
    </Button>
    <TooltipContent slot="tooltip-content">
      {#if hasUnsavedChanges}
        Save your changes first
      {:else if hasReconcileError}
        Fix connector errors first
      {:else}
        Import data using this connector
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
