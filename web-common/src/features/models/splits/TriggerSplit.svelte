<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Button } from "../../../components/button";
  import {
    V1ReconcileStatus,
    createRuntimeServiceCreateTrigger,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { addLeadingSlash } from "../../entity-management/entity-mappers";
  import { fileArtifacts } from "../../entity-management/file-artifacts";

  const queryClient = useQueryClient();
  const triggerMutation = createRuntimeServiceCreateTrigger();

  export let splitKey: string;

  function trigger() {
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [
          {
            model: resource?.meta?.name?.name as string,
            splits: [splitKey],
          },
        ],
      },
    });
  }

  $: ({ instanceId } = $runtime);

  // We access the `GetResource` result store from directly within this component
  // This avoids needing to pass down the result through the table & needing to hack reactivity
  $: ({ params } = $page);
  $: fileArtifact = fileArtifacts.getFileArtifact(addLeadingSlash(params.file));
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;
</script>

<Button
  type="secondary"
  on:click={trigger}
  disabled={resource?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
  noWrap
>
  Refresh split
</Button>
