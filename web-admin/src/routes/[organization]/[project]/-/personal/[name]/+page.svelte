<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    createRuntimeServiceListResources,
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { page } from "$app/state";
  import type { PageData } from "./$types";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import CanvasPersonalFile from "@rilldata/web-admin/features/personal-files/canvas/CanvasPersonalFile.svelte";

  let { data }: { data: PageData } = $props();
  let { personalFile, fileIo } = $derived(data);

  const client = useRuntimeClient();

  let resourceName = $derived(page.params.name);
  let resourceQuery = $derived(
    createRuntimeServiceListResources(
      client,
      {},
      {
        query: {
          select: (data) =>
            data.resources?.filter(
              (res) => res.meta?.name?.name === resourceName,
            ) ?? [],
        },
      },
      queryClient,
    ),
  );
  let resourceKind = $derived($resourceQuery.data?.[0]?.meta?.name?.kind);

  let fileArtifact = $derived(fileArtifacts.getFileArtifact(personalFile.path));
  $effect(() => {
    fileArtifact.editorContent.set(personalFile.yaml);
  });

  async function onFileWrite({ name, kind }: { name: string; kind: string }) {
    // Invalidate the resource caches so the preview re-fetches the reconciled canvas
    // (otherwise the user might see a stale render right after their edit).
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGetResourceQueryKey(client.instanceId, {
        name: { kind, name },
      }),
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(client.instanceId, {}),
    });
  }

  onMount(() => fileIo.on("write", (args) => void onFileWrite(args)));
</script>

{#if resourceKind === ResourceKind.Canvas}
  <CanvasPersonalFile {fileArtifact} name={resourceName} />
{:else}
  Unsupported resource kind: {resourceKind}
{/if}
