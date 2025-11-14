<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    V1ReconcileStatus,
    type V1Resource,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let resource: V1Resource | undefined;
  export let hasUnsavedChanges = false;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  $: ({ instanceId } = $runtime);
  $: connectorName = resource?.meta?.name?.name;
  $: isReconciling =
    resource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING;

  function refreshConnector() {
    if (!connectorName) return;
    void $triggerMutation.mutateAsync({
      instanceId,
      data: {
        resources: [{ kind: ResourceKind.Connector, name: connectorName }],
      },
    });
  }
</script>

<div class="flex items-center gap-x-2">
  <Tooltip distance={8}>
    <Button
      square
      type="secondary"
      onClick={refreshConnector}
      disabled={$triggerMutation.isPending ||
        isReconciling ||
        hasUnsavedChanges}
      loading={$triggerMutation.isPending}
      loadingCopy="Refreshing"
      label="Refresh Connector"
    >
      <RefreshIcon size="14px" />
    </Button>
    <TooltipContent slot="tooltip-content">
      {#if hasUnsavedChanges}
        Save your changes to refresh
      {:else}
        Refresh connector
      {/if}
    </TooltipContent>
  </Tooltip>
</div>
