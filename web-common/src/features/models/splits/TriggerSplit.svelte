<script lang="ts">
  import { Button } from "../../../components/button";
  import {
    V1ModelSplit,
    V1ReconcileStatus,
    V1Resource,
    createRuntimeServiceCreateTrigger,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let resource: V1Resource;
  export let split: V1ModelSplit;

  $: splitKey = split.key as string;
  $: ({ instanceId } = $runtime);

  const triggerMutation = createRuntimeServiceCreateTrigger();

  function trigger() {
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [
          {
            model: resource.meta?.name?.name as string,
            splits: [splitKey],
          },
        ],
      },
    });
  }
</script>

<Button
  on:click={trigger}
  disabled={resource.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
  loading={resource.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
>
  Refresh split
</Button>
