<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Button } from "../../../components/button";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTrigger,
    type V1Resource,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { addLeadingSlash } from "../../entity-management/entity-mappers";
  import { fileArtifacts } from "../../entity-management/file-artifacts";

  const queryClient = useQueryClient();
  const triggerMutation = createRuntimeServiceCreateTrigger();

  export let partitionKey: string;
  export let resource: V1Resource | undefined = undefined;

  function trigger() {
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [
          {
            model: resolvedResource?.meta?.name?.name as string,
            partitions: [partitionKey],
          },
        ],
      },
    });
  }

  $: ({ instanceId } = $runtime);

  // If resource is passed as prop, use it directly; otherwise derive from URL params (web-local)
  $: ({ params } = $page);
  $: fileArtifact =
    params.file !== undefined
      ? fileArtifacts.getFileArtifact(addLeadingSlash(params.file))
      : undefined;
  $: resourceQuery = fileArtifact?.getResource(queryClient, instanceId);
  $: resolvedResource = resource ?? $resourceQuery?.data;
</script>

<Button
  type="secondary"
  onClick={trigger}
  disabled={resolvedResource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
  noWrap
>
  Refresh partition
</Button>
