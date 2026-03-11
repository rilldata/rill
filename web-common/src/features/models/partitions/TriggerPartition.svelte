<script lang="ts">
  import { page } from "$app/stores";
  import { onDestroy } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Button } from "../../../components/button";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTriggerMutation,
    getRuntimeServiceGetModelPartitionsQueryKey,
    type V1Resource,
  } from "../../../runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { addLeadingSlash } from "../../entity-management/entity-mappers";
  import { fileArtifacts } from "../../entity-management/file-artifacts";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  export let partitionKey: string;
  export let resource: V1Resource | undefined = undefined;

  let pollInterval: ReturnType<typeof setInterval> | null = null;

  onDestroy(() => {
    if (pollInterval) {
      clearInterval(pollInterval);
    }
  });

  async function trigger() {
    const modelName = resolvedResource?.meta?.name?.name;
    if (!modelName) return;

    // Clear any existing poll interval
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }

    await $triggerMutation.mutateAsync({
      models: [
        {
          model: modelName,
          partitions: [partitionKey],
        },
      ],
    });

    // Poll for updates since partition execution happens asynchronously
    const invalidate = () =>
      queryClient.invalidateQueries({
        queryKey: getRuntimeServiceGetModelPartitionsQueryKey(instanceId, {
          model: modelName,
        }),
      });

    await invalidate();

    // Poll every 2 seconds for up to 30 seconds.
    // Note: we don't early-exit on IDLE because in web-admin, `resource` is a
    // snapshot (not a live query), so reconcileStatus would be stale.
    let pollCount = 0;
    const maxPolls = 15;
    pollInterval = setInterval(async () => {
      pollCount++;
      await invalidate();
      if (pollCount >= maxPolls && pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    }, 2000);
  }

  $: ({ instanceId } = runtimeClient);
  $: isLoading = $triggerMutation.isPending;

  // If resource is passed as prop, use it directly; otherwise derive from URL params (web-local)
  $: ({ params } = $page);
  $: fileArtifact =
    params.file !== undefined
      ? fileArtifacts.getFileArtifact(addLeadingSlash(params.file))
      : undefined;
  $: resourceQuery = fileArtifact?.getResource(queryClient);
  $: resolvedResource = resource ?? $resourceQuery?.data;
</script>

<Button
  type="secondary"
  onClick={trigger}
  disabled={isLoading ||
    resolvedResource?.meta?.reconcileStatus ===
      V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
  loading={isLoading}
  noWrap
>
  Refresh partition
</Button>
