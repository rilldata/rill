<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { SOURCES } from "@rilldata/web-common/features/sources/modal/constants";
  import { getSchemaNameFromDriver } from "@rilldata/web-common/features/sources/modal/connector-schemas";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { Plus } from "lucide-svelte";
  import AddDataModal from "@rilldata/web-common/features/add-data/AddDataModal.svelte";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";

  export let resource: V1Resource | undefined;
  export let hasUnsavedChanges = false;

  $: connectorName = resource?.meta?.name?.name;
  $: driverName = resource?.connector?.spec?.driver;
  $: hasReconcileError = !!resource?.meta?.reconcileError;
  // Map driver name to schema name for connector lookup
  $: schemaName = driverName ? getSchemaNameFromDriver(driverName) : null;
  $: isDataSource = schemaName ? SOURCES.includes(schemaName) : false;
  $: isDisabled = hasUnsavedChanges || hasReconcileError || !driverName;

  let addModelOpen = false;
</script>

{#if isDataSource}
  <Tooltip distance={8} suppress={!isDisabled}>
    <Button
      type="primary"
      onClick={() => (addModelOpen = true)}
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

{#if schemaName && connectorName}
  <AddDataModal
    config={{
      medium: BehaviourEventMedium.Button,
      space: MetricsEventSpace.Workspace,
      screen: MetricsEventScreenName.Connector,
    }}
    bind:open={addModelOpen}
    schema={schemaName}
    connector={connectorName}
  />
{/if}
