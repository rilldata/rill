<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { addSourceModal } from "@rilldata/web-common/features/sources/modal/add-source-visibility";
  import { SOURCES } from "@rilldata/web-common/features/sources/modal/constants";
  import { getSchemaNameFromDriver } from "@rilldata/web-common/features/sources/modal/connector-schemas";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Plus } from "lucide-svelte";

  export let resource: V1Resource | undefined;
  export let hasUnsavedChanges = false;

  $: connectorName = resource?.meta?.name?.name;
  $: driverName = resource?.connector?.spec?.driver;
  $: hasReconcileError = !!resource?.meta?.reconcileError;
  // Map driver name to schema name for connector lookup
  $: schemaName = driverName ? getSchemaNameFromDriver(driverName) : null;
  $: isDataSource = schemaName ? SOURCES.includes(schemaName) : false;
  $: isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;

  /**
   * Opens the Add Data modal pre-configured for this connector.
   * Passes the schema name (for form lookup) and connector instance name
   * so the modal can skip to the import step with the connector pre-selected.
   */
  function openAddModel() {
    if (!schemaName || !connectorName) return;
    addSourceModal.open(schemaName, connectorName);
  }
</script>

{#if isDataSource}
  <Tooltip distance={8} suppress={!isDisabled}>
    <Button
      type="primary"
      onClick={openAddModel}
      disabled={isDisabled}
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
      {/if}
    </TooltipContent>
  </Tooltip>
{/if}
