<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    V1ReconcileStatus,
    type V1Resource,
    createRuntimeServiceCreateTrigger,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let resource: V1Resource | undefined;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  $: ({ instanceId } = $runtime);
  $: connectorName = resource?.meta?.name?.name;
  $: isReconciling =
    resource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING;

  function refreshConnector() {
    if (!connectorName) return;
    $triggerMutation.mutate({
      instanceId,
      data: {
        resources: [{ kind: ResourceKind.Connector, name: connectorName }],
      },
    });
  }
</script>

<div class="flex items-center gap-x-2">
  <Button
    type="secondary"
    onClick={refreshConnector}
    disabled={$triggerMutation.isPending || isReconciling}
    loading={$triggerMutation.isPending}
    loadingCopy="Refreshing"
  >
    <RefreshIcon size="14px" />
    {isReconciling ? "Refreshing…" : "Refresh"}
  </Button>
</div>
